package pingrepository

import (
	"context"
	"encoding/json"

	"afikrim_a.bitbucket.org/simple-go-queue/core/entity"
	"afikrim_a.bitbucket.org/simple-go-queue/core/repository"
	"github.com/adjust/rmq/v5"
	"github.com/go-redis/redis/v8"
)

type pingRepository struct {
	publisherClient  *redis.Client
	subscriberClient *redis.Client
	queue            rmq.Queue
}

func NewPingRepository(publisherClient, subscriberClient *redis.Client, queue rmq.Queue) repository.PingRepository {
	return &pingRepository{
		publisherClient:  publisherClient,
		subscriberClient: subscriberClient,
		queue:            queue,
	}
}

func (r *pingRepository) PingEnqueue(ctx context.Context, callbackCh string) error {
	// _, err := r.enqueuer.Enqueue("ping", work.Q{"callback_ch": callbackCh})
	rawMsg, err := json.Marshal(&map[string]interface{}{"callback_ch": callbackCh})
	if err != nil {
		return err
	}

	if err := r.queue.Publish(string(rawMsg)); err != nil {
		return err
	}

	return nil
}

func (r *pingRepository) PingSubscriber(ctx context.Context, channel string) (*entity.PingResponse, error) {
	psc := r.subscriberClient.Subscribe(ctx, channel)
	defer psc.Close()

	for {
		msg, err := psc.ReceiveMessage(ctx)
		if err != nil {
			return nil, err
		}

		if msg != nil {
			res := &entity.PingResponse{}
			err = json.Unmarshal([]byte(msg.Payload), res)
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

	_, err = r.publisherClient.Publish(ctx, channel, resJson).Result()
	if err != nil {
		return err
	}

	return nil
}
