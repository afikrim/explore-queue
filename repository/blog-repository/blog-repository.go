package blogrepository

import (
	"context"
	"encoding/json"

	"afikrim_a.bitbucket.org/simple-go-queue/core/entity"
	"afikrim_a.bitbucket.org/simple-go-queue/core/repository"
	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	"gorm.io/gorm"
)

type blogRepository struct {
	db             *gorm.DB
	publisherPool  *redis.Pool
	subscriberPool *redis.Pool
	enqueuer       *work.Enqueuer
}

func NewBlogRepository(db *gorm.DB, publisherPool, subscriberPool *redis.Pool, enqueuer *work.Enqueuer) repository.BlogRepository {
	return &blogRepository{
		db:             db,
		publisherPool:  publisherPool,
		subscriberPool: subscriberPool,
		enqueuer:       enqueuer,
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
	conn, err := r.publisherPool.GetContext(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	resJson, err := json.Marshal(res)
	if err != nil {
		return err
	}

	_, err = conn.Do("PUBLISH", channel, resJson)
	if err != nil {
		return err
	}

	return nil
}

func (r *blogRepository) CreateBlogSubscriber(ctx context.Context, channel string) (*entity.CreateBlogResponse, error) {
	conn, err := r.subscriberPool.GetContext(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	psc := redis.PubSubConn{Conn: conn}
	psc.Subscribe(channel)

	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			res := &entity.CreateBlogResponse{}
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
