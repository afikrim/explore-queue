package handler

import (
	"afikrim_a.bitbucket.org/simple-go-queue/core/service"
	"github.com/labstack/echo/v4"
)

type PingHandler interface {
	Ping(ctx echo.Context) error

	RegisterHandler(e *echo.Group)
}

type pingHandler struct {
	pingSvc service.PingService
}

func NewPingHandler(pingSvc service.PingService) PingHandler {
	return &pingHandler{
		pingSvc: pingSvc,
	}
}

func (h *pingHandler) Ping(ctx echo.Context) error {
	res, err := h.pingSvc.Ping(ctx.Request().Context())
	if err != nil {
		return err
	}

	return ctx.JSON(200, res)
}

func (h *pingHandler) RegisterHandler(e *echo.Group) {
	e.GET("/ping", h.Ping)
}
