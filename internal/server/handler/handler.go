package handler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"os"
	"sn/internal/business/cache"
	hndlr "sn/internal/server/handler/v1/cache"
)

type Config struct {
	Shutdown   chan os.Signal
	Log        *zap.SugaredLogger
	Cache      *cache.Core
	TimeoutSec int
}

func RegisterAPIs(app *gin.Engine, cfg *Config) {
	h := hndlr.NewHandler(cfg.Cache, cfg.Log)

	app.POST("/set", h.Create)
	app.GET("/get", h.Get)
	app.DELETE("/del", h.Delete)

}
