package cache

import (
	"go.uber.org/zap"
	"sync"
	"time"
)

type ExpirationData struct {
	Key string
	Exp int64
}

type Core struct {
	log                 *zap.SugaredLogger
	data                map[string]interface{}
	sortedKeyExpiration []ExpirationData
	mu                  sync.RWMutex
	processingCh        chan any

	SetCh    chan SetCacheReq
	GetCh    chan GetValueReq
	DeleteCh chan DeleteValueReq
}

func NewCore(logger *zap.SugaredLogger, numWrkrs int) *Core {
	core := Core{
		log:                 logger,
		data:                make(map[string]interface{}),
		sortedKeyExpiration: []ExpirationData{},
		mu:                  sync.RWMutex{},
		processingCh:        make(chan interface{}, 100),

		SetCh:    make(chan SetCacheReq, 100),
		GetCh:    make(chan GetValueReq, 100),
		DeleteCh: make(chan DeleteValueReq, 100),
	}
	go core.synchronizeRequests()
	go core.processTTLs()
	go core.processRequests(numWrkrs)
	return &core
}

func (c *Core) set(key string, val interface{}, ttl int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = val
	expData := ExpirationData{Key: key, Exp: ttl}

	sortedKeyExpInd := c.findExpDataIndex(c.sortedKeyExpiration, expData)
	c.sortedKeyExpiration = append(c.sortedKeyExpiration[:sortedKeyExpInd], append([]ExpirationData{expData}, c.sortedKeyExpiration[sortedKeyExpInd:]...)...)
	c.log.Infow("set", "key", key, "value", val, "time", time.Now())
}

func (c *Core) get(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data[key]
}

func (c *Core) delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
}

func (c *Core) synchronizeRequests() {
	for {
		var req interface{}
		var ok bool

		select {
		case req, ok = <-c.SetCh:
		case req, ok = <-c.GetCh:
		case req, ok = <-c.DeleteCh:
		}
		if !ok {
			continue
		}
		c.processingCh <- req
	}
}

func (c *Core) processRequests(numWrkrs int) {
	var wg sync.WaitGroup
	for i := 0; i < numWrkrs; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for req := range c.processingCh {
				c.processRequest(req)
			}
		}()
	}
	wg.Wait()
}

func (c *Core) processRequest(req any) {
	switch v := req.(type) {
	case SetCacheReq:
		select {
		case <-v.Ctx.Done():
			return
		default:
		}
		c.set(v.Name, v.Value, v.Expiration)
		v.DoneCh <- struct{}{}

	case DeleteValueReq:
		select {
		case <-v.Ctx.Done():
			return
		default:
		}
		c.delete(v.Key)
		v.DoneCh <- struct{}{}
	case GetValueReq:
		select {
		case <-v.Ctx.Done():
			return
		default:
		}
		v.RespCh <- c.get(v.Key)
	}
}

func (c *Core) processTTLs() {
	for {
		lenSortedExpKey := len(c.sortedKeyExpiration)
		if lenSortedExpKey == 0 {
			time.Sleep(time.Second)
			continue
		}
		c.mu.Lock()
		firstKeyExpiration := c.sortedKeyExpiration[lenSortedExpKey-1]
		if firstKeyExpiration.Exp > time.Now().Unix() {
			c.mu.Unlock()
			nextTTLTime := time.Unix(firstKeyExpiration.Exp, 0)
			time.Sleep(nextTTLTime.Sub(time.Now()))
			continue
		}
		c.sortedKeyExpiration = c.sortedKeyExpiration[:lenSortedExpKey-1]
		delete(c.data, firstKeyExpiration.Key)
		c.log.Infow("expired", "key", firstKeyExpiration.Key, "time", time.Now())
		c.mu.Unlock()
	}
}

func (c *Core) findExpDataIndex(slice []ExpirationData, data ExpirationData) int {
	low, high := 0, len(slice)-1

	for low <= high {
		mid := (low + high) / 2
		if slice[mid].Exp < data.Exp {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}
	return low
}
