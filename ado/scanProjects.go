package ado

import (
	"bufio"
	"fmt"
	"github.com/microsoft/azure-devops-go-api/azuredevops/core"
	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"io"
	"regexp"
	"strings"
	"sync"
)

// ScanProjects uses the Service interface to interact with Azure DevOps and does all the heavy lifting to find data
type ScanProjects struct {
	adoService Service
	criteria   *SearchCriteria
	logger Logging
}

// Scan triggers the scan and aggregates all the Results into the Results struct for easy JSON marshaling to client
func (s *ScanProjects) Scan() (*Results, error) {
	projectsToScan, err := s.getProjects()
	if err != nil {
		return nil, err
	}

	errs := make(chan error)
	ch := make(chan Project, len(projectsToScan))
	wg := sync.WaitGroup{}
	projects := make([]Project, 0, len(projectsToScan))

	for _, project := range projectsToScan {
		wg.Add(1)
		go s.findContent(project.Name, ch, errs, &wg)
	}

	wg.Wait()
	close(ch)
	close(errs)

	for proj := range ch {
		projects = append(projects, proj)
	}

	return &Results{Projects: &projects}, nil
}

func (s *ScanProjects) getProjects() ([]core.TeamProjectReference, error) {
	var projects []core.TeamProjectReference
	response, err := s.adoService.GetProjects()
	if err != nil {
		return nil, err
	}

	for response != nil {
		projectsFiltered, err := s.processProjectResponse(response.Value)
		if err != nil {
			return nil, err
		}

		projects = append(projects, projectsFiltered...)

		if response.ContinuationToken != "" {
			response, err = s.adoService.GetAdditionalProjects(response.ContinuationToken)
			if err != nil {
				return nil, err
			}
		} else {
			response = nil
		}
	}
	return projects, nil
}

func (s *ScanProjects) processProjectResponse(projectToFilter []core.TeamProjectReference) (projectsFiltered []core.TeamProjectReference, err error) {
	for _, project := range projectToFilter {
		matchResults, err := regexp.MatchString(s.criteria.ProjectNamePattern, *project.Name)
		if err != nil {
			return nil, err
		}
		if matchResults {
			projectsFiltered = append(projectsFiltered, project)
		}
	}
	return projectsFiltered, nil
}

func (s *ScanProjects) findContent(projectName *string, project chan Project, errs chan error, parentWg *sync.WaitGroup) {
	repos, err := s.adoService.GetRepositories(*projectName)
	if err != nil {
		errs <- err
		return
	}

	ch := make(chan Repository, len(*repos))
	wg := sync.WaitGroup{}
	repositories := make([]Repository, 0, len(*repos))

	for _, repo := range *repos {
		wg.Add(1)
		go s.findFiles(repo.Name, projectName, ch, errs, &wg)
	}

	wg.Wait()
	close(ch)

	for repo := range ch {
		repositories = append(repositories, repo)
	}

	if len(repositories) > 0 {
		project <- Project{
			Name:         *projectName,
			Repositories: &repositories,
		}
	}
	
	parentWg.Done()
}

func (s *ScanProjects) findFiles(repoName, projectName *string, repository chan Repository, errs chan error, parentWg *sync.WaitGroup) {
	itemsReference, err := s.adoService.GetItems(*projectName, *repoName)
	if err != nil {
		if !strings.Contains(err.Error(), "Cannot find any branches for the") {
			s.logger.LogError(err)
			fmt.Printf("Error Found: %s\n", err)
		}
	}

	if itemsReference != nil {
		items, err := s.findContentInFile(repoName, projectName, itemsReference, errs)
		if err != nil {
			errs <- err
			return
		}

		if len(items) > 0 {
			repository <- Repository{
				Name:  *repoName,
				Files: &items,
			}
		}
	}

	parentWg.Done()
}

func (s *ScanProjects) findContentInFile(repoName, projectName *string, itemsReference *[]git.GitItem, errs chan error) ([]Item, error) {
	ch := make(chan Item, len(*itemsReference))
	wg := sync.WaitGroup{}
	items := make([]Item, 0, len(*itemsReference))

	for _, itemRef := range *itemsReference {
		matchResults, err := regexp.MatchString(s.criteria.FileNamePattern, *itemRef.Path)
		if err != nil {
			return nil, err
		}
		if *itemRef.GitObjectType == "blob" && matchResults {
			wg.Add(1)
			item, err := s.adoService.GetItemContent(*projectName, *repoName, *itemRef.Path)
			if err != nil {
				return nil, err
			}
			go s.processFile(itemRef.Path, item, ch, errs, &wg)
		}
	}
	wg.Wait()
	close(ch)

	for item := range ch {
		items = append(items, item)
	}
	return items, nil
}

func (s *ScanProjects) processFile(itemName *string, file io.ReadCloser, item chan Item, errs chan error, parentWg *sync.WaitGroup) {

	var lines []string
	srcScanner := bufio.NewScanner(file)
	srcScanner.Split(bufio.ScanLines)
	for srcScanner.Scan() {
		line := srcScanner.Text()
		matchResults, err := regexp.MatchString(s.criteria.ContentPattern, line)
		if err != nil {
			errs <- err
			return
		}
		if matchResults {
			lines = append(lines, line)
		}
	}

	if len(lines) > 0 {
		item <- Item{
			Name:  *itemName,
			Lines: &lines,
		}
	}

	parentWg.Done()
}