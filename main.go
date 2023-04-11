package main

import (
	"context"
	"errors"
	"log"
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
	"github.com/adjust/rmq/v5"
	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
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

	errChan := make(chan error, 10)
	go logErrors(errChan)

	workerClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	publisherClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	subscriberClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// init queue
	queueConn, err := rmq.OpenConnectionWithRedisClient("blog", workerClient, errChan)
	if err != nil {
		panic(err)
	}

	blogQueue, err := queueConn.OpenQueue("blog")
	if err != nil {
		panic(err)
	}
	pingQueue, err := queueConn.OpenQueue("ping")
	if err != nil {
		panic(err)
	}

	// init repositories
	blogRepo := blogrepository.NewBlogRepository(db, publisherClient, subscriberClient, blogQueue)
	pingRepo := pingrepository.NewPingRepository(publisherClient, subscriberClient, pingQueue)

	// init services
	blogService := service.NewBlogService(blogRepo)
	pingService := service.NewPingService(pingRepo)

	// init handlers
	blogHandlerApi := handlerApi.NewBlogHandler(blogService)
	pingHandlerApi := handlerApi.NewPingHandler(pingService)

	createBlogHandlerWorker := handlerWorker.NewCreateBlogHandler(blogService)
	pingHandlerWorker := handlerWorker.NewPingHandler(pingService)

	// init consumer
	createBlogHandlerWorker.RegisterHandler(blogQueue)
	pingHandlerWorker.RegisterHandler(pingQueue)

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

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		panic(err)
	}
}

func logErrors(errChan <-chan error) {
	for err := range errChan {
		switch err := err.(type) {
		case *rmq.HeartbeatError:
			if err.Count == rmq.HeartbeatErrorLimit {
				log.Print("heartbeat error (limit): ", err)
			} else {
				log.Print("heartbeat error: ", err)
			}
		case *rmq.ConsumeError:
			log.Print("consume error: ", err)
		case *rmq.DeliveryError:
			log.Print("delivery error: ", err.Delivery, err)
		default:
			log.Print("other error: ", err)
		}
	}
}
