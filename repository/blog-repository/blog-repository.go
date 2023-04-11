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
	db         *gorm.DB
	publisher  redis.Conn
	subscriber redis.Conn
	enqueuer   *work.Enqueuer
}

func NewBlogRepository(db *gorm.DB, publisher, subscriber redis.Conn, enqueuer *work.Enqueuer) repository.BlogRepository {
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

	_, err = r.publisher.Do("PUBLISH", channel, resJson)
	if err != nil {
		return err
	}

	return nil
}

func (r *blogRepository) CreateBlogSubscriber(ctx context.Context, channel string) (*entity.CreateBlogResponse, error) {
	psc := redis.PubSubConn{Conn: r.subscriber}
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
