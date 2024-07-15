package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/project/pkg/object"
	"os"
	"strconv"
	"time"
)

const (
	rateLimitingRemainingHeader = "X-RateLimit-Remaining"
	rateLimitingResetHeader     = "X-RateLimit-Reset"
)

type github struct {
}

func NewGithub() object.GitDetails {
	return github{}
}

func (g github) SearchRepos(ctx context.Context, interest string) ([]object.Repository, int64, error) {
	var (
		response Repositories
		result   []object.Repository
	)

	client := resty.New()
	resp, err := client.R().
		SetContext(ctx).
		Get(fmt.Sprintf("%s/search/repositories?q=%s", os.Getenv("GITHUB_BASE_URL"), interest))
	if err != nil {
		return nil, 0, err
	}

	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, 0, err
	}

	for _, rr := range response.Items {
		result = append(result, object.Repository{
			Name:            rr.Name,
			Owner:           rr.Owner.Login,
			Description:     rr.Description,
			URL:             rr.Url,
			Language:        rr.Language,
			ForksCount:      rr.ForksCount,
			StarsCount:      rr.StargazersCount,
			OpenIssuesCount: rr.OpenIssuesCount,
			WatchersCount:   rr.WatchersCount,
		})
	}

	rateLimitReset := resp.Header().Get(rateLimitingResetHeader)
	rateLimitRemaining := resp.Header().Get(rateLimitingRemainingHeader)
	if rateLimitRemaining == "0" {
		resetTime, err := strconv.ParseInt(rateLimitReset, 10, 64)
		if err != nil {
			return nil, 0, err
		}

		return nil, resetTime, errors.New("rate_limit")
	}

	return result, 0, nil
}

func (github) FetchRepo(ctx context.Context, owner, repo string) (*object.Repository, int64, error) {
	client := resty.New()
	resp, err := client.R().
		SetContext(ctx).
		Get(fmt.Sprintf("%s/repos/%s/%s", os.Getenv("GITHUB_BASE_URL"), owner, repo))
	if err != nil {
		return nil, 0, err
	}

	var repository object.Repository
	if err := json.Unmarshal(resp.Body(), &repository); err != nil {
		return nil, 0, err
	}

	rateLimitReset := resp.Header().Get(rateLimitingResetHeader)
	rateLimitRemaining := resp.Header().Get(rateLimitingRemainingHeader)
	if rateLimitRemaining == "0" {
		resetTime, err := strconv.ParseInt(rateLimitReset, 10, 64)
		if err != nil {
			return nil, 0, err
		}

		return nil, resetTime, errors.New("rate_limit")
	}
	return &repository, 0, nil
}

func (github) FetchCommits(ctx context.Context, owner, repo string) ([]object.Commit, int64, error) {
	client := resty.New()
	resp, err := client.R().
		SetContext(ctx).
		Get(fmt.Sprintf("%s/repos/%s/%s/commits", os.Getenv("GITHUB_BASE_URL"), owner, repo))
	if err != nil {
		return nil, 0, err
	}

	var commits []struct {
		SHA    string `json:"sha"`
		Commit struct {
			Author struct {
				Name  string    `json:"name"`
				Email string    `json:"email"`
				Date  time.Time `json:"date"`
			} `json:"author"`
			Message string `json:"message"`
		} `json:"commit"`
	}
	if err := json.Unmarshal(resp.Body(), &commits); err != nil {
		return nil, 0, err
	}

	var commitList []object.Commit
	for _, c := range commits {
		commitList = append(commitList, object.Commit{
			SHA:         c.SHA,
			AuthorName:  c.Commit.Author.Name,
			AuthorEmail: c.Commit.Author.Email,
			Message:     c.Commit.Message,
			Date:        c.Commit.Author.Date,
		})
	}

	rateLimitReset := resp.Header().Get(rateLimitingResetHeader)
	rateLimitRemaining := resp.Header().Get(rateLimitingRemainingHeader)
	if rateLimitRemaining == "0" {
		resetTime, err := strconv.ParseInt(rateLimitReset, 10, 64)
		if err != nil {
			return nil, 0, err
		}
		return nil, resetTime, errors.New("rate_limit")
	}
	return commitList, 0, nil
}
