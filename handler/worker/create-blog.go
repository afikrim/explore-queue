package handler

import (
	"context"
	"encoding/json"
	"time"

	"afikrim_a.bitbucket.org/simple-go-queue/core/entity"
	"afikrim_a.bitbucket.org/simple-go-queue/core/service"
	"github.com/adjust/rmq/v5"
)

type CreateBlogHandler interface {
	Consume(delivery rmq.Delivery)

	RegisterHandler(queue rmq.Queue)
}

type createBlogHandler struct {
	blogSvc service.BlogService
}

func NewCreateBlogHandler(blogSvc service.BlogService) CreateBlogHandler {
	return &createBlogHandler{
		blogSvc: blogSvc,
	}
}

func (h *createBlogHandler) Consume(delivery rmq.Delivery) {
	req := &entity.CreateBlogRequestQueue{}
	err := json.Unmarshal([]byte(delivery.Payload()), req)
	if err != nil {
		delivery.Reject()
		return
	}

	err = h.blogSvc.CreateBlogWorker(context.Background(), req.CallbackCh, &req.CreateBlogRequest)
	if err != nil {
		delivery.Reject()
		return
	}

	delivery.Ack()
}

func (h *createBlogHandler) RegisterHandler(queue rmq.Queue) {
	queue.StartConsuming(10, 1*time.Second)

	for i := 0; i < 10; i++ {
		queue.AddConsumer("create-blog", h)
	}
}
