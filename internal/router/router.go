package router

import (
	"github.com/gin-gonic/gin"

	"swiftwallet/internal/handler"
	"swiftwallet/internal/service"
)

func New(s service.Service) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	h := handler.New(s)

	api := r.Group("/api/v1")
	{
		api.POST("/wallet", h.Operate)
		api.GET("/wallets/:id", h.Balance)
	}
	return r
}
