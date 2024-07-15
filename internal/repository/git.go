package repository

import (
	"context"
	"errors"
	"github.com/project/internal/model"
	"gorm.io/gorm"
)

type IGitRepo interface {
	CreateRepoRecord(context.Context, model.Repository) error
	UpdateRepoRecord(context.Context, model.Repository) error
	CreateCommitRecord(context.Context, []model.Commit) error
	GetRepo(context.Context, string, string) (*model.Repository, error)
	GetRepos(context.Context, int, int) ([]model.Repository, int64, error)
	GetReposByLanguage(context.Context, string) ([]model.Repository, error)
	GetTopNRepoByStarCount(context.Context, int) ([]model.Repository, error)
}

type gitRepo struct {
	db *gorm.DB
}

func NewGitDBRepo(db *gorm.DB) IGitRepo {
	return gitRepo{
		db: db,
	}
}

func (g gitRepo) CreateRepoRecord(ctx context.Context, repository model.Repository) error {
	return g.db.WithContext(ctx).Create(&repository).Error
}

func (g gitRepo) UpdateRepoRecord(ctx context.Context, repository model.Repository) error {
	return g.db.WithContext(ctx).Where("id = ?", repository.ID).Updates(&repository).Error
}

func (g gitRepo) CreateCommitRecord(ctx context.Context, commit []model.Commit) error {
	return g.db.WithContext(ctx).Create(&commit).Error
}

func (g gitRepo) GetRepo(ctx context.Context, owner, name string) (*model.Repository, error) {
	var resp model.Repository
	if err := g.db.WithContext(ctx).Where("owner = ? AND name = ?", owner, name).First(&resp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &resp, nil
}

func (g gitRepo) GetRepos(ctx context.Context, size, page int) ([]model.Repository, int64, error) {
	offset := size * (page - 1)
	var (
		resp  []model.Repository
		total int64
	)

	if err := g.db.WithContext(ctx).Model(&model.Repository{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	total = total / int64(size)
	if err := g.db.WithContext(ctx).Model(&model.Repository{}).Offset(offset).Limit(size).Find(&resp).Error; err != nil {
		return nil, 0, err
	}

	return resp, total, nil
}

func (g gitRepo) GetReposByLanguage(ctx context.Context, language string) ([]model.Repository, error) {
	var resp []model.Repository

	if err := g.db.WithContext(ctx).Where("language = ?", language).Find(&resp).Error; err != nil {
		return nil, err
	}

	return resp, nil
}

func (g gitRepo) GetTopNRepoByStarCount(ctx context.Context, n int) ([]model.Repository, error) {
	var repos []model.Repository

	if err := g.db.WithContext(ctx).Order("stars_count desc").Limit(n).Find(&repos).Error; err != nil {
		return nil, err
	}

	return repos, nil
}
