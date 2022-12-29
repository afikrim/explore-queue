package repository

import (
	"context"

	"afikrim_a.bitbucket.org/simple-go-queue/core/entity"
)

type BlogRepository interface {
	CreateBlog(ctx context.Context, blog *entity.Blog) (int64, error)
}
