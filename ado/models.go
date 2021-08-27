package ado

// Results is the keeper of all the projects scanned to be used to create the JSON blob that gets returned
type Results struct {
	Projects *[]Project
}

// Project contains the name of the project and all repositories that contained information that matched the criteria
type Project struct {
	Name string
	Repositories *[]Repository
}

// Repository contains the name of the repo and all the items that contained information that matched the criteria
type Repository struct {
	Name string
	Files *[]Item
}

// Item contains the name of the item and all the lines that matched the search criteria
type Item struct {
	Name string
	Lines *[]string
}

// SearchCriteria is the payload that gets sent in the post to search for the Project, File, and Contents
type SearchCriteria struct {
	ProjectNamePattern string
	FileNamePattern string
	ContentPattern string
}
