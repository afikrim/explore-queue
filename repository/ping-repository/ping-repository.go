package pingrepository

import (
	"context"
	"encoding/json"

	"afikrim_a.bitbucket.org/simple-go-queue/core/entity"
	"afikrim_a.bitbucket.org/simple-go-queue/core/repository"
	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
)

type pingRepository struct {
	publisher  redis.Conn
	subscriber redis.Conn
	enqueuer   *work.Enqueuer
}

func NewPingRepository(publisher, subscriber redis.Conn, enqueuer *work.Enqueuer) repository.PingRepository {
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
	psc := redis.PubSubConn{Conn: r.subscriber}
	psc.Subscribe(channel)

	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			res := &entity.PingResponse{}
			err := json.Unmarshal(v.Data, res)
			if err != nil {
				return nil, err
			}

			return res, nil
		case redis.Subscription:
			if v.Count == 0 {
				return nil, nil
			}
		case error:
			return nil, v
		}
	}
}

func (r *pingRepository) PingResponsePublish(ctx context.Context, channel string, res *entity.PingResponse) error {
	resJson, err := json.Marshal(res)
	if err != nil {
		return err
	}

	_, err = r.publisher.Do("PUBLISH", channel, resJson)
	if err != nil {
		return err
	}

	return nil
}
