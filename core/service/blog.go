package service

import (
	"context"

	"afikrim_a.bitbucket.org/simple-go-queue/core/entity"
	"afikrim_a.bitbucket.org/simple-go-queue/core/repository"
)

type BlogService interface {
	CreateBlog(ctx context.Context, req *entity.CreateBlogRequest) (*entity.CreateBlogResponse, error)
}

type blogService struct {
	blogRepo repository.BlogRepository
}

func NewBlogService(blogRepo repository.BlogRepository) BlogService {
	return &blogService{
		blogRepo: blogRepo,
	}
}

func (s *blogService) CreateBlog(ctx context.Context, req *entity.CreateBlogRequest) (*entity.CreateBlogResponse, error) {
	blog := entity.NewBlog(req.Title, req.Body)
	id, err := s.blogRepo.CreateBlog(ctx, blog)
	if err != nil {
		return nil, err
	}

	return &entity.CreateBlogResponse{ID: id}, nil
}
