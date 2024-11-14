package v1

import (
	"github.com/gin-gonic/gin"
	"sn/internal/server/handlers"
	hndlr "sn/internal/server/handlers/v1/cache"
)

func RegisterV1APIs(app *gin.Engine, cfg *handlers.Config) {
	h := hndlr.NewHandler(cfg.Cache, cfg.Log)

	app.POST("/set", h.Create)
	app.GET("/get", h.Get)
	app.DELETE("/del", h.Delete)

}
