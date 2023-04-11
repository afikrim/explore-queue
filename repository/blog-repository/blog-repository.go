package blogrepository

import (
	"context"
	"encoding/json"

	"afikrim_a.bitbucket.org/simple-go-queue/core/entity"
	"afikrim_a.bitbucket.org/simple-go-queue/core/repository"
	"github.com/adjust/rmq/v5"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type blogRepository struct {
	db               *gorm.DB
	publisherClient  *redis.Client
	subscriberClient *redis.Client
	queue            rmq.Queue
}

func NewBlogRepository(db *gorm.DB, publisherClient, subscriberClient *redis.Client, queue rmq.Queue) repository.BlogRepository {
	return &blogRepository{
		db:               db,
		publisherClient:  publisherClient,
		subscriberClient: subscriberClient,
		queue:            queue,
	}
}

func (r *blogRepository) CreateBlog(ctx context.Context, blog *entity.Blog) (int64, error) {
	dto := BlogDto{}.
		FromEntity(blog).
		InitTimestamps()

	err := r.db.WithContext(ctx).Create(dto).Error
	if err != nil {
		return 0, err
	}

	return dto.ID, nil
}

func (r *blogRepository) CreateBlogEnqueue(ctx context.Context, callbackCh string, req *entity.Blog) error {
	rawMsg, err := json.Marshal(&map[string]interface{}{"title": req.Title, "body": req.Body, "callback_ch": callbackCh})
	if err != nil {
		return err
	}

	if err := r.queue.Publish(string(rawMsg)); err != nil {
		return err
	}

	return nil
}

func (r *blogRepository) CreateBlogResponsePublish(ctx context.Context, channel string, res *entity.CreateBlogResponse) error {
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

func (r *blogRepository) CreateBlogSubscriber(ctx context.Context, channel string) (*entity.CreateBlogResponse, error) {
	psc := r.subscriberClient.Subscribe(ctx, channel)
	defer psc.Close()

	for {
		msg, err := psc.ReceiveMessage(ctx)
		if err != nil {
			return nil, err
		}

		if msg != nil {
			res := &entity.CreateBlogResponse{}
			err = json.Unmarshal([]byte(msg.Payload), res)
			if err != nil {
				return nil, err
			}

			return res, nil
		}
	}
}
