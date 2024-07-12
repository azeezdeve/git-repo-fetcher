package mock_data

import (
	"context"
	"github.com/project/pkg/object"
)

type MockGitDetails struct {
	FetchRepoFunc    func(ctx context.Context, owner, repo string) (*object.Repository, int64, error)
	FetchCommitsFunc func(ctx context.Context, owner, repo string) ([]object.Commit, int64, error)
}

func (m *MockGitDetails) FetchRepo(ctx context.Context, owner, repo string) (*object.Repository, int64, error) {
	return m.FetchRepoFunc(ctx, owner, repo)
}

func (m *MockGitDetails) FetchCommits(ctx context.Context, owner, repo string) ([]object.Commit, int64, error) {
	return m.FetchCommitsFunc(ctx, owner, repo)
}
