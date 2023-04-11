package handler

import (
	"context"
	"encoding/json"
	"time"

	"afikrim_a.bitbucket.org/simple-go-queue/core/entity"
	"afikrim_a.bitbucket.org/simple-go-queue/core/service"
	"github.com/adjust/rmq/v5"
)

type PingHandler interface {
	Consume(delivery rmq.Delivery)

	RegisterHandler(queue rmq.Queue)
}

type pingHandler struct {
	pingSvc service.PingService
}

func NewPingHandler(pingSvc service.PingService) PingHandler {
	return &pingHandler{
		pingSvc: pingSvc,
	}
}

func (h *pingHandler) Consume(delivery rmq.Delivery) {
	req := &entity.PingRequestQueue{}
	err := json.Unmarshal([]byte(delivery.Payload()), req)
	if err != nil {
		delivery.Reject()
		return
	}

	err = h.pingSvc.PingWorker(context.Background(), req.CallbackCh)
	if err != nil {
		delivery.Reject()
		return
	}

	delivery.Ack()
}

func (h *pingHandler) RegisterHandler(queue rmq.Queue) {
	queue.StartConsuming(50, 1*time.Second)

	for i := 0; i < 10; i++ {
		queue.AddConsumer("ping", h)
	}
}
