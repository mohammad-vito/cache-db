package cache

import (
	"context"
	"errors"
	"go.uber.org/zap"
)

type Service struct {
	log   *zap.SugaredLogger
	cache *Core
}

var ErrInvalidKey = errors.New("invalid key")
var ErrCtxCancelled = errors.New("context cancelled")

func NewService(c *Core, logger *zap.SugaredLogger) Service {
	return Service{cache: c, log: logger}
}

func (s *Service) Set(ctx context.Context, key string, value interface{}, expirationUnix int64) error {
	doneCh := make(chan struct{}, 1)
	select {
	case s.cache.SetCh <- SetCacheReq{
		Ctx:        ctx,
		Name:       key,
		Value:      value,
		Expiration: expirationUnix,
		DoneCh:     doneCh}:

		<-doneCh
		return nil

	case <-ctx.Done():
		return ErrCtxCancelled
	}
}

func (s *Service) Get(ctx context.Context, key string) (interface{}, error) {
	respCh := make(chan interface{}, 1)
	select {
	case s.cache.GetCh <- GetValueReq{
		Ctx:    ctx,
		Key:    key,
		RespCh: respCh,
	}:
		v := <-respCh
		if v == nil {
			return nil, ErrInvalidKey
		}
		return v, nil

	case <-ctx.Done():
		return nil, ErrCtxCancelled
	}
}

func (s *Service) Delete(ctx context.Context, name string) error {
	respCh := make(chan struct{}, 1)
	select {
	case s.cache.DeleteCh <- DeleteValueReq{
		Ctx:    ctx,
		Key:    name,
		DoneCh: respCh,
	}:
		<-respCh
		return nil
	case <-ctx.Done():
		return ErrCtxCancelled
	}
}
