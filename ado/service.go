package ado


import (
	"context"
	"errors"
	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/core"
	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"io"
)

// Service interface is used to provide you access the API you intend to scan and enable it to be mocked for testing
type Service interface {
	GetProjects() (*core.GetProjectsResponseValue, error)
	GetAdditionalProjects(continuationToken string) (*core.GetProjectsResponseValue, error)
	GetRepositories(projectName string) (*[]git.GitRepository, error)
	GetItems(projectName string, repoName string) (*[]git.GitItem, error)
	GetItemContent(projectName string, repoName string, path string) (io.ReadCloser, error)
	CreateConnection(orgURL, pat string) error
}

// AzureDevOpsService implements the Service interface and provides you the access to the Azure DevOps APIs
type AzureDevOpsService struct {
	connection *azuredevops.Connection
}

// CreateConnection establishes the connection used by the other methods in the interface
func (conn *AzureDevOpsService) CreateConnection(orgURL, pat string) error {
	conn.connection = azuredevops.NewPatConnection(orgURL, pat)
	if conn.connection == nil {
		return errors.New("unable to connect to azure devops")
	}

	return nil
}

func (conn *AzureDevOpsService) getCoreClient() (core.Client, error) {
	coreClient, err := core.NewClient(context.Background(), conn.connection)
	if err != nil {
		return nil, err
	}
	return coreClient, nil
}

func (conn *AzureDevOpsService) getGitClient() (git.Client, error) {
	gitClient, err := git.NewClient(context.Background(), conn.connection)
	if err != nil {
			return nil, err
		}
	return gitClient, nil
}

// GetProjects scans for projects that match the Project Search Criteria
func (conn *AzureDevOpsService) GetProjects() (*core.GetProjectsResponseValue, error) {
	coreClient, err := conn.getCoreClient()
	if err != nil {
		return nil, err
	}

	response, err := coreClient.GetProjects(context.Background(), core.GetProjectsArgs{})
	if err != nil {
		return nil, err
	}

	return response, nil
}

// GetAdditionalProjects uses a continuation token to scan for more projects
func (conn *AzureDevOpsService) GetAdditionalProjects(continuationToken string) (*core.GetProjectsResponseValue, error) {
	coreClient, err := conn.getCoreClient()
	if err != nil {
		return nil, err
	}

	response, err := coreClient.GetProjects(context.Background(), core.GetProjectsArgs{ContinuationToken: &continuationToken})
	if err != nil {
		return nil, err
	}

	return response, nil
}

// GetRepositories scans all repositories for a project and processes any repo that has a default branch
func (conn *AzureDevOpsService) GetRepositories(projectName string) (*[]git.GitRepository, error) {
	gitClient, err := conn.getGitClient()
	if err != nil {
		return nil, err
	}

	repos, err := gitClient.GetRepositories(context.Background(), git.GetRepositoriesArgs{Project: &projectName})
	if err != nil {
		return nil, err
	}

	return repos, nil
}

// GetItems scans all items in repository that matches the search criteria for file name
func (conn *AzureDevOpsService) GetItems(projectName, repoName string) (*[]git.GitItem, error){
	gitClient, err := conn.getGitClient()
	if err != nil {
		return nil, err
	}

	itemsReference, err := gitClient.GetItems(context.Background(), git.GetItemsArgs{RepositoryId: &repoName, Project: &projectName, RecursionLevel: &git.VersionControlRecursionTypeValues.Full})
	if err != nil {
		return nil, err
	}

	return itemsReference, nil
}

// GetItemContent scans all lines in a file and returns a list of each line that contains the search criteria
func (conn *AzureDevOpsService) GetItemContent(projectName, repoName, path string) (io.ReadCloser, error){
	gitClient, err := conn.getGitClient()
	if err != nil {
		return nil, err
	}

	includeContent := true
	item, err := gitClient.GetItemContent(context.Background(), git.GetItemContentArgs{RepositoryId: &repoName, Project: &projectName, Path: &path, IncludeContent: &includeContent})
	if err != nil {
		return nil, err
	}

	return item, nil
}