package repository

import (
	"context"

	"afikrim_a.bitbucket.org/simple-go-queue/core/entity"
)

type PingRepository interface {
	PingEnqueue(ctx context.Context, callbackCh string) error
	PingSubscriber(ctx context.Context, channel string) (*entity.PingResponse, error)
	PingResponsePublish(ctx context.Context, channel string, res *entity.PingResponse) error
}
