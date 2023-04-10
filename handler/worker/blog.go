package worker

import (
	"context"

	"afikrim_a.bitbucket.org/simple-go-queue/core/entity"
	"afikrim_a.bitbucket.org/simple-go-queue/core/service"
	"github.com/gocraft/work"
)

type BlogHandler interface {
	CreateBlog(job *work.Job) error

	RegisterHandler(pool *work.WorkerPool)
}

type blogHandler struct {
	blogSvc service.BlogService
}

func NewBlogHandler(blogSvc service.BlogService) BlogHandler {
	return &blogHandler{
		blogSvc: blogSvc,
	}
}

func (h *blogHandler) CreateBlog(job *work.Job) error {
	req := &entity.CreateBlogRequest{
		Title: job.ArgString("title"),
		Body:  job.ArgString("body"),
	}
	callbackCh := job.ArgString("callback_ch")

	return h.blogSvc.CreateBlogWorker(context.Background(), callbackCh, req)
}

func (h *blogHandler) RegisterHandler(pool *work.WorkerPool) {
	pool.Job("create-blog", h.CreateBlog)
}
