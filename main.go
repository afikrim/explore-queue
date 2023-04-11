package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"afikrim_a.bitbucket.org/simple-go-queue/core/service"
	handlerApi "afikrim_a.bitbucket.org/simple-go-queue/handler/api"
	handlerWorker "afikrim_a.bitbucket.org/simple-go-queue/handler/worker"
	blogrepository "afikrim_a.bitbucket.org/simple-go-queue/repository/blog-repository"
	pingrepository "afikrim_a.bitbucket.org/simple-go-queue/repository/ping-repository"
	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	"github.com/labstack/echo/v4"
	goredis "github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type WorkerContext struct{}

func main() {
	// init db connection using gorm
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// init db schema
	if err := db.AutoMigrate(&blogrepository.BlogDto{}); err != nil {
		panic(err)
	}

	redisPool := &redis.Pool{
		MaxActive: 25,
		MaxIdle:   10,
		Wait:      true,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", ":6379")
		},
	}
	enqueuer := work.NewEnqueuer("blog", redisPool)

	publisher := goredis.NewClient(&goredis.Options{
		Addr:     ":6379",
		Password: "",
		DB:       0,
	})

	subscriber := goredis.NewClient(&goredis.Options{
		Addr:     ":6379",
		Password: "",
		DB:       0,
	})

	// init repositories
	blogRepo := blogrepository.NewBlogRepository(db, publisher, subscriber, enqueuer)
	pingRepo := pingrepository.NewPingRepository(publisher, subscriber, enqueuer)

	// init services
	blogService := service.NewBlogService(blogRepo)
	pingService := service.NewPingService(pingRepo)

	// init handlers
	blogHandlerApi := handlerApi.NewBlogHandler(blogService)
	pingHandlerApi := handlerApi.NewPingHandler(pingService)

	blogHandlerWorker := handlerWorker.NewBlogHandler(blogService)
	pingHandlerWorker := handlerWorker.NewPingHandler(pingService)

	// init worker
	workerPool := work.NewWorkerPool(WorkerContext{}, 10, "blog", redisPool)

	// register jobs
	blogHandlerWorker.RegisterHandler(workerPool)
	pingHandlerWorker.RegisterHandler(workerPool)

	// init echo
	e := echo.New()

	// init groups
	apiV1 := e.Group("/api/v1")

	// register handlers
	blogHandlerApi.RegisterHandler(apiV1)
	pingHandlerApi.RegisterHandler(apiV1)

	// start echo
	go func() {
		if err := e.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
