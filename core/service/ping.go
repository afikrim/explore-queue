package service

import (
	"context"
	"fmt"

	"afikrim_a.bitbucket.org/simple-go-queue/core/entity"
	"afikrim_a.bitbucket.org/simple-go-queue/core/repository"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type PingService interface {
	Ping(ctx context.Context) (*entity.PingResponse, error)
	PingWorker(ctx context.Context, callbackCh string) error
}

type pingService struct {
	pingRepo repository.PingRepository
}

func NewPingService(pingRepo repository.PingRepository) PingService {
	return &pingService{
		pingRepo: pingRepo,
	}
}

func (s *pingService) Ping(ctx context.Context) (*entity.PingResponse, error) {
	trxId, err := gonanoid.New(32)
	if err != nil {
		return nil, err
	}

	if err := s.pingRepo.PingEnqueue(ctx, fmt.Sprintf("ping-%s", trxId)); err != nil {
		return nil, err
	}

	res, err := s.pingRepo.PingSubscriber(ctx, fmt.Sprintf("ping-%s", trxId))
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *pingService) PingWorker(ctx context.Context, callbackCh string) error {
	res := &entity.PingResponse{Message: "PONG!"}
	if err := s.pingRepo.PingResponsePublish(ctx, callbackCh, res); err != nil {
		return err
	}

	return nil
}
