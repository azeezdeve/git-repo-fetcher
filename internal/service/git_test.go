package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/project/internal/model"
	"github.com/project/internal/service/mock_data"
	"github.com/project/pkg/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

// Mock repository
type MockGitRepo struct {
	mock.Mock
}

func (m *MockGitRepo) CreateCommitRecord(ctx context.Context, commit []model.Commit) error {
	return m.Called(ctx, commit).Error(0)
}

func (m *MockGitRepo) GetReposByLanguage(ctx context.Context, s string) ([]model.Repository, error) {
	args := m.Called(ctx, s)
	return args.Get(0).([]model.Repository), args.Error(1)
}

func (m *MockGitRepo) GetTopNRepoByStarCount(ctx context.Context, i int) ([]model.Repository, error) {
	args := m.Called(ctx, i)
	return args.Get(0).([]model.Repository), args.Error(1)
}

func (m *MockGitRepo) GetRepos(ctx context.Context, limit, page int) ([]model.Repository, int64, error) {
	args := m.Called(ctx, limit, page)
	return args.Get(0).([]model.Repository), args.Get(1).(int64), args.Error(2)
}

func (m *MockGitRepo) GetRepo(ctx context.Context, owner, repo string) (*model.Repository, error) {
	return &model.Repository{
		ID: uuid.New(),
	}, nil
}

func (m *MockGitRepo) CreateRepoRecord(ctx context.Context, repo model.Repository) error {
	return m.Called(ctx, repo).Error(0)
}

func (m *MockGitRepo) UpdateRepoRecord(ctx context.Context, repo model.Repository) error {
	return nil
}

// Mock GitDetails for FetchRepo testing
type MockGitDetails struct {
	mock.Mock
}

func (m *MockGitDetails) FetchRepo(ctx context.Context, owner, repo string) (*object.Repository, int64, error) {
	return &object.Repository{
		Name: repo,
	}, 1, nil
}

func (m *MockGitDetails) FetchCommits(ctx context.Context, owner, repo string) ([]object.Commit, int64, error) {
	args := m.Called(ctx, owner, repo)
	return args.Get(0).([]object.Commit), args.Get(1).(int64), args.Error(2)
}

// Test FetchRepo method
func TestFetchRepo(t *testing.T) {
	mockRepo := new(MockGitRepo)
	mockGitDetails := new(MockGitDetails)
	mockDetails := &mock_data.MockGitDetails{
		FetchRepoFunc: func(ctx context.Context, owner, repo string) (*object.Repository, int64, error) {
			return &object.Repository{
				Name:        repo,
				Description: "A sample repository",
				URL:         "https://github.com/" + owner + "/" + repo,
				Language:    "Go",
				ForksCount:  10,
				StarsCount:  100,
			}, 0, nil
		},
	}
	gitService := gitInfo{repo: mockRepo, gitDetails: mockDetails}

	ctx := context.Background()
	owner := "owner"
	repo := "repo"
	repoResp := &object.Repository{
		Name:            "repo",
		Description:     "description",
		URL:             "url",
		Language:        "Go",
		ForksCount:      10,
		StarsCount:      100,
		OpenIssuesCount: 5,
		WatchersCount:   20,
		CreatedAt:       time.Now().String(),
		UpdatedAt:       time.Now().String(),
	}

	mockRepo.On("GetRepo", ctx, owner, repo).Return(model.Repository{ID: uuid.New()}, nil)
	mockGitDetails.On("FetchRepo", ctx, owner, repo).Return(repoResp, int64(1), nil)
	mockRepo.On("CreateRepoRecord", ctx, mock.AnythingOfType("model.Repository")).Return(nil)

	_, err := gitService.FetchRepo(ctx, owner, repo)
	assert.NoError(t, err)
}
