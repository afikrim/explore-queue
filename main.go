package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"afikrim_a.bitbucket.org/simple-go-queue/core/service"
	handler "afikrim_a.bitbucket.org/simple-go-queue/handler/api"
	blog_repository "afikrim_a.bitbucket.org/simple-go-queue/repository/blog"
	"github.com/labstack/echo/v4"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

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

	// init repositories
	blogRepo := blog_repository.NewBlogRepository(db)

	// init services
	blogService := service.NewBlogService(blogRepo)

	// init handlers
	blogHandler := handler.NewBlogHandler(blogService)

	// init echo
	e := echo.New()

	// init groups
	apiV1 := e.Group("/api/v1")

	// register handlers
	blogHandler.RegisterHandler(apiV1)

	// start echo
	go func() {
		if err := e.Start(":8080"); err != nil {
			panic(err)
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		panic(err)
	}
}
