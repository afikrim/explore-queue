package pingrepository

import (
	"context"
	"encoding/json"

	"afikrim_a.bitbucket.org/simple-go-queue/core/entity"
	"afikrim_a.bitbucket.org/simple-go-queue/core/repository"
	"github.com/gocraft/work"
	goredis "github.com/redis/go-redis/v9"
)

type pingRepository struct {
	publisher  *goredis.Client
	subscriber *goredis.Client
	enqueuer   *work.Enqueuer
}

func NewPingRepository(publisher, subscriber *goredis.Client, enqueuer *work.Enqueuer) repository.PingRepository {
	return &pingRepository{
		publisher:  publisher,
		subscriber: subscriber,
		enqueuer:   enqueuer,
	}
}

func (r *pingRepository) PingEnqueue(ctx context.Context, callbackCh string) error {
	_, err := r.enqueuer.Enqueue("ping", work.Q{"callback_ch": callbackCh})
	if err != nil {
		return err
	}

	return nil
}

func (r *pingRepository) PingSubscriber(ctx context.Context, channel string) (*entity.PingResponse, error) {
	psc := r.subscriber.Subscribe(ctx, channel)

	for {
		msg, err := psc.ReceiveMessage(ctx)
		if err != nil {
			return nil, err
		}

		if msg != nil {
			res := &entity.PingResponse{}
			err = json.Unmarshal([]byte(msg.Payload), &res)
			if err != nil {
				return nil, err
			}

			return res, nil
		}
	}
}

func (r *pingRepository) PingResponsePublish(ctx context.Context, channel string, res *entity.PingResponse) error {
	resJson, err := json.Marshal(res)
	if err != nil {
		return err
	}

	_, err = r.publisher.Publish(ctx, channel, resJson).Result()
	if err != nil {
		return err
	}

	return nil
}
