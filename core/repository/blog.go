package repository

import (
	"context"

	"afikrim_a.bitbucket.org/simple-go-queue/core/entity"
)

type BlogRepository interface {
	CreateBlog(ctx context.Context, blog *entity.Blog) (int64, error)
	CreateBlogEnqueue(ctx context.Context, callbackCh string, blog *entity.Blog) error
	CreateBlogSubscriber(ctx context.Context, channel string) (*entity.CreateBlogResponse, error)
	CreateBlogResponsePublish(ctx context.Context, channel string, res *entity.CreateBlogResponse) error
}
