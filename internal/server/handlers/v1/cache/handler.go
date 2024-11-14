package cache

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"sn/internal/pkg/cache"
	"sn/internal/schemata"
	"time"
)

type Handler struct {
	log *zap.SugaredLogger
	s   *cache.Service
}

func NewHandler(core *cache.Core, log *zap.SugaredLogger) Handler {
	s := cache.NewService(core, log)
	return Handler{log: log, s: &s}
}

func (h Handler) Create(c *gin.Context) {
	var in schemata.SetCacheIn
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	expTime := time.Now().Add(time.Second * time.Duration(in.TTL))
	if err := h.s.Set(c, in.Key, in.Value, expTime.Unix()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}

func (h Handler) Get(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing key in query param"})
		return
	}
	v, err := h.s.Get(c, key)
	switch {
	case errors.Is(err, cache.ErrInvalidKey):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	default:
	}

	outputResp := schemata.GetCacheOut{
		Key:   key,
		Value: v,
	}
	c.JSON(http.StatusOK, outputResp)
}

func (h Handler) Delete(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing key in query param"})
		return
	}
	if err := h.s.Delete(c, key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}
