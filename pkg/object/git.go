package object

import (
	"context"
	"time"
)

type GitDetails interface {
	FetchRepo(ctx context.Context, owner, repo string) (*Repository, int64, error)
	FetchCommits(ctx context.Context, owner, repo string) ([]Commit, int64, error)
}

type Repository struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	URL             string `json:"html_url"`
	Language        string `json:"language"`
	ForksCount      int    `json:"forks_count"`
	StarsCount      int    `json:"stargazers_count"`
	OpenIssuesCount int    `json:"open_issues_count"`
	WatchersCount   int    `json:"watchers_count"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

type Commit struct {
	RepoID      uint      `json:"repo_id"`
	SHA         string    `json:"sha"`
	AuthorName  string    `json:"author_name"`
	AuthorEmail string    `json:"author_email"`
	Message     string    `json:"message"`
	Date        time.Time `json:"date"`
}
