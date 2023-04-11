package blogrepository

import (
	"context"
	"encoding/json"

	"afikrim_a.bitbucket.org/simple-go-queue/core/entity"
	"afikrim_a.bitbucket.org/simple-go-queue/core/repository"
	"github.com/gocraft/work"
	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type blogRepository struct {
	db         *gorm.DB
	publisher  *goredis.Client
	subscriber *goredis.Client
	enqueuer   *work.Enqueuer
}

func NewBlogRepository(db *gorm.DB, publisher, subscriber *goredis.Client, enqueuer *work.Enqueuer) repository.BlogRepository {
	return &blogRepository{
		db:         db,
		publisher:  publisher,
		subscriber: subscriber,
		enqueuer:   enqueuer,
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
	_, err := r.enqueuer.Enqueue("create-blog", work.Q{"title": req.Title, "body": req.Body, "callback_ch": callbackCh})
	if err != nil {
		return err
	}

	return nil
}

func (r *blogRepository) CreateBlogResponsePublish(ctx context.Context, channel string, res *entity.CreateBlogResponse) error {
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

func (r *blogRepository) CreateBlogSubscriber(ctx context.Context, channel string) (*entity.CreateBlogResponse, error) {
	psc := r.subscriber.Subscribe(ctx, channel)

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
