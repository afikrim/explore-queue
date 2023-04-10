package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"afikrim_a.bitbucket.org/simple-go-queue/core/entity"
	"afikrim_a.bitbucket.org/simple-go-queue/core/service"
	handlerApi "afikrim_a.bitbucket.org/simple-go-queue/handler/api"
	handlerWorker "afikrim_a.bitbucket.org/simple-go-queue/handler/worker"
	blog_repository "afikrim_a.bitbucket.org/simple-go-queue/repository/blog"
	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	"github.com/labstack/echo/v4"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type WorkerContext struct {
	entity.Blog
}

func main() {
	// init db connection using gorm
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// init db schema
	if err := db.AutoMigrate(&blog_repository.BlogDto{}); err != nil {
		panic(err)
	}

	redisPool := &redis.Pool{
		MaxActive: 5,
		MaxIdle:   5,
		Wait:      true,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", ":6379")
		},
	}
	enqueuer := work.NewEnqueuer("blog", redisPool)

	publisherPool := &redis.Pool{
		MaxActive: 5,
		MaxIdle:   5,
		Wait:      true,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", ":6379")
		},
	}
	publisher := publisherPool.Get()

	subscriberPool := &redis.Pool{
		MaxActive: 5,
		MaxIdle:   5,
		Wait:      true,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", ":6379")
		},
	}
	subscriber := subscriberPool.Get()

	// init repositories
	blogRepo := blog_repository.NewBlogRepository(db, publisher, subscriber, enqueuer)

	// init services
	blogService := service.NewBlogService(blogRepo)

	// init handlers
	blogHandlerApi := handlerApi.NewBlogHandler(blogService)
	blogHandlerWorker := handlerWorker.NewBlogHandler(blogService)

	// init worker
	workerPool := work.NewWorkerPool(WorkerContext{}, 10, "blog", redisPool)

	// register jobs
	blogHandlerWorker.RegisterHandler(workerPool)

	// init echo
	e := echo.New()

	// init groups
	apiV1 := e.Group("/api/v1")

	// register handlers
	blogHandlerApi.RegisterHandler(apiV1)

	// start echo
	go func() {
		if err := e.Start(":8080"); err != nil {
			panic(err)
		}
	}()

	// start worker
	go func() {
		workerPool.Start()
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		panic(err)
	}
	workerPool.Stop()
}
