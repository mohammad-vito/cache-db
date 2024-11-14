package handlers

import (
	"go.uber.org/zap"
	"sn/internal/pkg/cache"
)

type Config struct {
	Log        *zap.SugaredLogger
	Cache      *cache.Core
	TimeoutSec int
}
