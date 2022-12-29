package blog_repository

import (
	"context"

	"afikrim_a.bitbucket.org/simple-go-queue/core/entity"
	"afikrim_a.bitbucket.org/simple-go-queue/core/repository"
	"gorm.io/gorm"
)

type blogRepository struct {
	db *gorm.DB
}

func NewBlogRepository(db *gorm.DB) repository.BlogRepository {
	return &blogRepository{
		db: db,
	}
}

func (r *blogRepository) CreateBlog(ctx context.Context, blog *entity.Blog) (int64, error) {
	dto := BlogDto{}.
		FromEntity(blog).
		InitTimestamps()

	err := r.db.WithContext(ctx).Create(dto).Error
	if err != nil {
		return 0, err
	}

	return dto.ID, nil
}
