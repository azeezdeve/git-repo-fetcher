package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/project/internal/model"
	"github.com/project/internal/repository"
	"github.com/project/pkg/object"
	"log"
	"sync"
	"time"
)

type IGitInfo interface {
	SearchRepos(ctx context.Context, interest string) error
	FetchRepo(ctx context.Context, name, repo string) (*model.Repository, error)
	UpdateRepo(ctx context.Context) error
	GetCommit(ctx context.Context, name, repo string) ([]model.Commit, error)
	GetRepoByLanguage(ctx context.Context, language string) ([]model.Repository, error)
	GetTopNRepoByStarCount(ctx context.Context, n int) ([]model.Repository, error)
}

type gitInfo struct {
	repo       repository.IGitRepo
	gitDetails object.GitDetails
}

func NewGitInfo(repo repository.IGitRepo, gitDetails object.GitDetails) IGitInfo {
	return gitInfo{repo: repo, gitDetails: gitDetails}
}

func (g gitInfo) SearchRepos(ctx context.Context, interest string) error {
	var (
		repoResp []object.Repository
		rate     int64
		err      error
	)
	for {
		repoResp, rate, err = g.gitDetails.SearchRepos(ctx, interest)
		if err != nil {
			if err.Error() == "rate_limit" {
				time.Sleep(time.Duration(rate) * time.Minute)
				continue
			}
			log.Printf("error fetching repo, err %v", err)
			return errors.New("unable to process")
		}

		break
	}

	for _, rr := range repoResp {
		err = g.upsertRepo(ctx, rr)
		if err != nil {
			log.Printf("error processing repository %s/%s, error: %v", rr.Owner, rr.Name, err)
		}
	}

	return nil
}

func (g gitInfo) FetchRepo(ctx context.Context, owner, repo string) (*model.Repository, error) {
	resp, err := g.repo.GetRepo(ctx, owner, repo)
	if err != nil {
		log.Printf("error fetching repo, err %v", err)
		return nil, errors.New("unable to process")
	}

	gitDetail := g.gitDetails

	var (
		repoResp *object.Repository
		rate     int64
	)
	for {
		repoResp, rate, err = gitDetail.FetchRepo(ctx, owner, repo)
		if err != nil {
			if err.Error() == "rate_limit" {
				time.Sleep(time.Duration(rate) * time.Minute)
				continue
			}
			log.Printf("error fetching repo, err %v", err)
			return nil, errors.New("unable to process")
		}

		break
	}

	payload := model.Repository{
		ID:              uuid.New(),
		Name:            repoResp.Name,
		Owner:           owner,
		Description:     repoResp.Description,
		URL:             repoResp.URL,
		Language:        repoResp.Language,
		ForksCount:      repoResp.ForksCount,
		StarsCount:      repoResp.StarsCount,
		OpenIssuesCount: repoResp.OpenIssuesCount,
		WatchersCount:   repoResp.WatchersCount,
		CreatedAt:       repoResp.CreatedAt,
		UpdatedAt:       repoResp.UpdatedAt,
	}

	if resp != nil {
		payload.ID = resp.ID
		err = g.repo.UpdateRepoRecord(ctx, payload)
		return resp, err
	}

	err = g.repo.CreateRepoRecord(ctx, payload)
	if err != nil {
		return nil, err
	}

	return &payload, nil
}

func (g gitInfo) UpdateRepo(ctx context.Context) error {
	var wg sync.WaitGroup
	repoChan := make(chan model.Repository, 10)
	fetchMore := make(chan struct{}, 1)
	done := make(chan struct{})
	var maxPages int

	// Start worker goroutines
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for repo := range repoChan {
				_, err := g.GetCommit(ctx, repo.Owner, repo.Name)
				if err != nil {
					log.Printf("Error fetching commit: %v", err)
				}
				if len(repoChan) == 0 {
					fetchMore <- struct{}{}
				}
			}
		}()
	}

	sendRepos := func(page int) {
		repos, totalPage, err := g.repo.GetRepos(ctx, 10, page)
		if err != nil {
			log.Printf("Error fetching repos: %v", err)
			close(done)
			return
		}

		mutex := sync.Mutex{}
		mutex.Lock()
		maxPages = int(totalPage)
		mutex.Unlock()

		for _, repo := range repos {
			repoChan <- repo
		}
	}

	// Send initial batch of repositories
	go sendRepos(1)

	// Monitor for fetchMore signals and send the next page of repositories
	go func() {
		page := 2
		for {
			select {
			case <-fetchMore:
				if page > maxPages {
					close(done)
					return
				}
				sendRepos(page)
				page++
			case <-done:
				return
			}
		}
	}()

	// Wait for all goroutines to finish
	wg.Wait()
	close(repoChan)
	close(fetchMore)

	return nil
}

func (g gitInfo) GetCommit(ctx context.Context, name, repo string) ([]model.Commit, error) {
	repoResp, err := g.FetchRepo(ctx, name, repo)
	if err != nil {
		return nil, err
	}

	gitDetail := g.gitDetails

	var (
		commitResp []model.Commit
		rate       int64
	)

	for {
		commits, callrate, err := gitDetail.FetchCommits(ctx, name, repo)
		if err != nil {
			if err.Error() == "rate_limit" {
				time.Sleep(time.Duration(rate) * time.Minute)
				continue
			}
			log.Printf("error fetching repo, err %v", err)
			return nil, errors.New("unable to process")
		}

		rate = callrate

		for _, commit := range commits {
			commitResp = append(commitResp, model.Commit{
				ID:          uuid.New(),
				RepoID:      repoResp.ID,
				SHA:         commit.SHA,
				AuthorEmail: commit.AuthorEmail,
				AuthorName:  commit.AuthorName,
				Message:     commit.Message,
				CommitDate:  commit.Date,
			})
		}

		break
	}

	err = g.repo.CreateCommitRecord(ctx, commitResp)
	if err != nil {
		return nil, err
	}

	return commitResp, nil
}

func (g gitInfo) GetRepoByLanguage(ctx context.Context, language string) ([]model.Repository, error) {
	return g.repo.GetReposByLanguage(ctx, language)
}

func (g gitInfo) GetTopNRepoByStarCount(ctx context.Context, n int) ([]model.Repository, error) {
	return g.repo.GetTopNRepoByStarCount(ctx, n)
}

func (g gitInfo) upsertRepo(ctx context.Context, rr object.Repository) error {
	repo, err := g.repo.GetRepo(ctx, rr.Name, rr.Owner)
	if err != nil {
		return err
	}

	if repo != nil {
		err = g.repo.UpdateRepoRecord(ctx, *repo)
		if err != nil {
			log.Printf("error updating record with id: %s, error: %v", repo.ID, err)
			return err
		}
	} else {
		data := model.Repository{
			ID:              uuid.New(),
			Name:            rr.Name,
			Owner:           rr.Owner,
			Description:     rr.Description,
			URL:             rr.URL,
			Language:        rr.Language,
			ForksCount:      rr.ForksCount,
			StarsCount:      rr.StarsCount,
			OpenIssuesCount: rr.OpenIssuesCount,
			WatchersCount:   rr.WatchersCount,
			CreatedAt:       rr.CreatedAt,
			UpdatedAt:       rr.UpdatedAt,
		}
		err = g.repo.CreateRepoRecord(ctx, data)
		if err != nil {
			log.Printf("error creating record, error: %v", err)
			return err
		}
	}
	return nil
}
