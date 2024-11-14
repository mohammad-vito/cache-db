package server

import (
	"github.com/gin-gonic/gin"
	"sn/internal/server/handlers"
	"sn/internal/server/handlers/mid"
	"sn/internal/server/handlers/v1"
	"time"
)

func APIMux(cfg *handlers.Config) *gin.Engine {
	r := gin.Default()
	r.Use(mid.TimeoutMiddleware(time.Duration(cfg.TimeoutSec) * time.Second))
	v1.RegisterV1APIs(r, cfg)
	return r
}
