package ado

import (
	mocks "adoscanner/mocks/ado"
	"fmt"
	"github.com/google/uuid"
	"github.com/microsoft/azure-devops-go-api/azuredevops/core"
	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"io/ioutil"
	"strings"
	"testing"
)

const (
	GetProjectsFuncName = "GetProjects"
	GetAdditionalProjectFuncName = "GetAdditionalProjects"
	GetRepositoriesFuncName = "GetRepositories"
	GetItemsFuncName = "GetItems"
	GetItemContentFuncName = "GetItemContent"

)

func sProjects(connections Service) *ScanProjects {
	scanProjects := ScanProjects{
		adoService: connections,
		criteria: &SearchCriteria{
			ProjectNamePattern: "Project",
			FileNamePattern:    "File",
			ContentPattern:     "Content",
		},
	}

	return &scanProjects
}

func TestScanWithNotProjectsFound(t *testing.T) {
	mockConnection := new(mocks.Service)
	mockConnection.On(GetProjectsFuncName).Return(new(core.GetProjectsResponseValue), nil)
	results, err := sProjects(mockConnection).Scan()
	assert.NotNil(t, results)
	assert.Nil(t, err)

	mockConnection.AssertNumberOfCalls(t, GetProjectsFuncName, 1)
	mockConnection.AssertNumberOfCalls(t, GetAdditionalProjectFuncName, 0)
	mockConnection.AssertNumberOfCalls(t, GetRepositoriesFuncName, 0)
	mockConnection.AssertNumberOfCalls(t, GetItemsFuncName, 0)
	mockConnection.AssertNumberOfCalls(t, GetItemContentFuncName, 0)
	mockConnection.AssertExpectations(t)
}

func TestScanWithOneProjectFound(t *testing.T) {
	numOfProjects := 1
	mockConnection := new(mocks.Service)
	mockConnection.On(GetProjectsFuncName).Return(getProjectTestData(numOfProjects, ""), nil)
	mockConnection.On(GetRepositoriesFuncName, "Project0").Return(new([]git.GitRepository), nil)
	results, err := sProjects(mockConnection).Scan()
	assert.NotNil(t, results)
	assert.Nil(t, err)

	mockConnection.AssertNumberOfCalls(t, GetProjectsFuncName, 1)
	mockConnection.AssertNumberOfCalls(t, GetAdditionalProjectFuncName, 0)
	mockConnection.AssertNumberOfCalls(t, GetRepositoriesFuncName, numOfProjects)
	mockConnection.AssertNumberOfCalls(t, GetItemsFuncName, 0)
	mockConnection.AssertNumberOfCalls(t, GetItemContentFuncName, 0)
	mockConnection.AssertExpectations(t)
}

func TestScanWithTwoProjectsFound(t *testing.T) {
	numOfProjects := 2
	mockConnection := new(mocks.Service)
	mockConnection.On(GetProjectsFuncName).Return(getProjectTestData(numOfProjects, ""), nil)
	mockConnection.On(GetRepositoriesFuncName, mock.Anything).Return(new([]git.GitRepository), nil)
	results, err := sProjects(mockConnection).Scan()
	assert.NotNil(t, results)
	assert.Nil(t, err)

	mockConnection.AssertNumberOfCalls(t, GetProjectsFuncName, 1)
	mockConnection.AssertNumberOfCalls(t, GetAdditionalProjectFuncName, 0)
	mockConnection.AssertNumberOfCalls(t, GetRepositoriesFuncName, numOfProjects)
	mockConnection.AssertNumberOfCalls(t, GetItemsFuncName, 0)
	mockConnection.AssertNumberOfCalls(t, GetItemContentFuncName, 0)
	mockConnection.AssertExpectations(t)
}

func TestScanWithFourProjectsFoundUsingAdditionalProjects(t *testing.T) {
	numOfProjects := 2
	mockConnection := new(mocks.Service)
	mockConnection.On(GetProjectsFuncName).Return(getProjectTestData(numOfProjects, "yes"), nil)
	mockConnection.On(GetAdditionalProjectFuncName, mock.Anything).Return(getProjectTestData(numOfProjects, ""), nil)
	mockConnection.On(GetRepositoriesFuncName, mock.Anything).Return(new([]git.GitRepository), nil)
	results, err := sProjects(mockConnection).Scan()
	assert.NotNil(t, results)
	assert.Nil(t, err)

	mockConnection.AssertNumberOfCalls(t, GetProjectsFuncName, 1)
	mockConnection.AssertNumberOfCalls(t, GetAdditionalProjectFuncName, 1)
	mockConnection.AssertNumberOfCalls(t, GetRepositoriesFuncName, numOfProjects*2)
	mockConnection.AssertNumberOfCalls(t, GetItemsFuncName, 0)
	mockConnection.AssertNumberOfCalls(t, GetItemContentFuncName, 0)
	mockConnection.AssertExpectations(t)
}

func TestScanWithOneProjectFoundWithOneItem(t *testing.T) {
	numOfProjects := 1
	numOfRepos := 1
	numOfItems := 1
	expectedResults := Results{
		Projects: &[]Project{{
			Name:         "Project0",
			Repositories: &[]Repository{{
				Name:  "Repo0",
				Files: &[]Item{{
					Name: "File0",
					Lines: &[]string{"Content To Test"},
				}},
			}},
		}}}
	mockConnection := new(mocks.Service)
	mockConnection.On(GetProjectsFuncName).Return(getProjectTestData(numOfProjects, ""), nil)
	mockConnection.On(GetRepositoriesFuncName, "Project0").Return(getRepositoryTestData(numOfRepos), nil)
	mockConnection.On(GetItemsFuncName, mock.Anything, mock.Anything).Return(getItemTestData(numOfItems), nil)
	mockConnection.On(GetItemContentFuncName, mock.Anything, mock.Anything, mock.Anything).Return(getItemContentTestData(), nil)
	results, err := sProjects(mockConnection).Scan()
	assert.NotNil(t, results)
	assert.Equal(t, &expectedResults, results)
	assert.Nil(t, err)

	mockConnection.AssertNumberOfCalls(t, GetProjectsFuncName, 1)
	mockConnection.AssertNumberOfCalls(t, GetAdditionalProjectFuncName, 0)
	mockConnection.AssertNumberOfCalls(t, GetRepositoriesFuncName, numOfProjects)
	mockConnection.AssertNumberOfCalls(t, GetItemsFuncName, numOfItems)
	mockConnection.AssertNumberOfCalls(t, GetItemContentFuncName, numOfItems)
	mockConnection.AssertExpectations(t)
}

func getProjectTestData(numOfProjects int, continuationToken string) *core.GetProjectsResponseValue {

	var projectReferences []core.TeamProjectReference

	for i := 0; i < numOfProjects; i++ {
		projectName := fmt.Sprintf("Project%d", i)
		id := uuid.New()
		projectReferences = append(projectReferences, core.TeamProjectReference{
			Id:                  &id,
			Name:                &projectName,
		})
	}

	projectRespValue := core.GetProjectsResponseValue{
		Value:             projectReferences,
		ContinuationToken: continuationToken,
	}
	return &projectRespValue
}

func getRepositoryTestData(numOfRepos int) *[]git.GitRepository {
	var repos []git.GitRepository
	for i := 0; i < numOfRepos; i++ {
		repoName := fmt.Sprintf("Repo%d", i)
		repos = append(repos, git.GitRepository{
			Name: &repoName,
		})
	}
	return &repos
}

func getItemTestData(numOfItems int) *[]git.GitItem {
	var gitItems []git.GitItem
	for i := 0; i < numOfItems; i++ {
		itemName := fmt.Sprintf("File%d", i)
		gitItems = append(gitItems, git.GitItem{
			Path: &itemName,
			GitObjectType: &git.GitObjectTypeValues.Blob,
		})
	}
	return &gitItems
}

func getItemContentTestData() io.ReadCloser {
	return ioutil.NopCloser(strings.NewReader("Content To Test\nboo"))
}