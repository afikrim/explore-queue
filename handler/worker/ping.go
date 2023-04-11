package handler

import (
	"context"

	"afikrim_a.bitbucket.org/simple-go-queue/core/service"
	"github.com/gocraft/work"
)

type PingHandler interface {
	Ping(job *work.Job) error

	RegisterHandler(pool *work.WorkerPool)
}

type pingHandler struct {
	pingSvc service.PingService
}

func NewPingHandler(pingSvc service.PingService) PingHandler {
	return &pingHandler{
		pingSvc: pingSvc,
	}
}

func (h *pingHandler) Ping(job *work.Job) error {
	callbackCh := job.ArgString("callback_ch")

	return h.pingSvc.PingWorker(context.Background(), callbackCh)
}

func (h *pingHandler) RegisterHandler(pool *work.WorkerPool) {
	pool.PeriodicallyEnqueue("* * * * *", "ping")
	pool.JobWithOptions("ping", work.JobOptions{MaxFails: 1, MaxConcurrency: 10}, h.Ping)
}
