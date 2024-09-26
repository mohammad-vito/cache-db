package cache

import "context"

type SetCacheReq struct {
	Ctx        context.Context
	Name       string
	Value      interface{}
	Expiration int64
	DoneCh     chan struct{}
}

type GetValueReq struct {
	Ctx    context.Context
	Key    string
	RespCh chan interface{}
}

type DeleteValueReq struct {
	Ctx    context.Context
	Key    string
	DoneCh chan struct{}
}
