package model

import (
	"github.com/google/uuid"
	"time"
)

type Repository struct {
	ID              uuid.UUID
	Name            string `json:"name" gorm:"uniqueIndex:idx_name_owner"`
	Owner           string `json:"owner" gorm:"uniqueIndex:idx_name_owner"`
	Description     string `json:"description"`
	URL             string `json:"html_url"`
	Language        string `gorm:"index" json:"language"`
	ForksCount      int    `json:"forks_count"`
	StarsCount      int    `json:"stargazers_count"`
	OpenIssuesCount int    `json:"open_issues_count"`
	WatchersCount   int    `json:"watchers_count"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

type Commit struct {
	ID          uuid.UUID
	RepoID      uuid.UUID `json:"repo_id"`
	SHA         string    `json:"sha"`
	AuthorName  string    `json:"author_name"`
	AuthorEmail string    `json:"author_email"`
	Message     string    `json:"message"`
	CommitDate  time.Time `json:"commit_date"`
}
