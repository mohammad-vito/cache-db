package server

import (
	"github.com/gin-gonic/gin"
	"sn/internal/server/handlers"
	"sn/internal/server/handlers/v1"
)

func APIMux(cfg *handlers.Config) *gin.Engine {
	r := gin.Default()
	v1.RegisterV1APIs(r, cfg)
	return r
}
