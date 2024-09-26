package server

import (
	"github.com/gin-gonic/gin"
	"sn/internal/server/handler"
	"sn/internal/server/handler/mid"
	"time"
)

func APIMux(cfg *handler.Config) *gin.Engine {
	r := gin.Default()
	r.Use(mid.TimeoutMiddleware(time.Duration(cfg.TimeoutSec) * time.Second))
	handler.RegisterAPIs(r, cfg)
	return r
}
