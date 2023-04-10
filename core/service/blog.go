package service

import (
	"context"
	"fmt"

	"afikrim_a.bitbucket.org/simple-go-queue/core/entity"
	"afikrim_a.bitbucket.org/simple-go-queue/core/repository"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type BlogService interface {
	CreateBlog(ctx context.Context, req *entity.CreateBlogRequest) (*entity.CreateBlogResponse, error)
	CreateBlogWorker(ctx context.Context, callbackCh string, req *entity.CreateBlogRequest) error
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
	trxId, err := gonanoid.New(32)
	if err != nil {
		return nil, err
	}

	blog := entity.NewBlog(req.Title, req.Body)
	if err := s.blogRepo.CreateBlogEnqueue(ctx, fmt.Sprintf("create-blog-%s", trxId), blog); err != nil {
		return nil, err
	}

	res, err := s.blogRepo.CreateBlogSubscriber(ctx, fmt.Sprintf("create-blog-%s", trxId))
	if err != nil {
		return nil, err
	}

	return &entity.CreateBlogResponse{ID: res.ID}, nil
}

func (s *blogService) CreateBlogWorker(ctx context.Context, callbackCh string, req *entity.CreateBlogRequest) error {
	blog := entity.NewBlog(req.Title, req.Body)
	id, err := s.blogRepo.CreateBlog(ctx, blog)
	if err != nil {
		return err
	}

	res := &entity.CreateBlogResponse{ID: id}
	if err := s.blogRepo.CreateBlogResponsePublish(ctx, callbackCh, res); err != nil {
		return err
	}

	return nil
}
