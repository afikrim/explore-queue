package handler

import (
	"afikrim_a.bitbucket.org/simple-go-queue/core/entity"
	"afikrim_a.bitbucket.org/simple-go-queue/core/service"
	"github.com/labstack/echo/v4"
)

type BlogHandler interface {
	CreateBlog(ctx echo.Context) error

	RegisterHandler(e *echo.Group)
}

type blogHandler struct {
	blogSvc service.BlogService
}

func NewBlogHandler(blogSvc service.BlogService) BlogHandler {
	return &blogHandler{
		blogSvc: blogSvc,
	}
}

func (h *blogHandler) CreateBlog(ctx echo.Context) error {
	var req entity.CreateBlogRequest
	if err := ctx.Bind(&req); err != nil {
		return err
	}

	res, err := h.blogSvc.CreateBlog(ctx.Request().Context(), &req)
	if err != nil {
		return err
	}

	return ctx.JSON(200, res)
}

func (h *blogHandler) RegisterHandler(e *echo.Group) {
	e.POST("/blogs", h.CreateBlog)
}
