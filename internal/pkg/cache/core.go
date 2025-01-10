package cache

import (
	"go.uber.org/zap"
	"sync"
	"time"
)

type Core struct {
	log                 *zap.SugaredLogger
	data                map[string]interface{}
	sortedKeyExpiration []ExpirationData
	mu                  sync.RWMutex

	SetCh    chan SetCacheReq
	GetCh    chan GetValueReq
	DeleteCh chan DeleteValueReq
}

func NewCore(logger *zap.SugaredLogger) *Core {
	core := Core{
		log:                 logger,
		data:                make(map[string]interface{}),
		sortedKeyExpiration: []ExpirationData{},
		mu:                  sync.RWMutex{},

		SetCh:    make(chan SetCacheReq),
		GetCh:    make(chan GetValueReq),
		DeleteCh: make(chan DeleteValueReq),
	}
	go core.processTTLQueue()
	go core.processRequests()
	return &core
}

func (c *Core) set(key string, val interface{}, expUnix int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = val
	expData := ExpirationData{Key: key, Exp: expUnix}

	sortedKeyExpInd := findExpDataInsertingIndex(c.sortedKeyExpiration, expData)
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

func (c *Core) processRequests() {
	for {
		select {
		case req := <-c.SetCh:
			c.set(req.Name, req.Value, req.Expiration)
			req.DoneCh <- struct{}{}
		case req := <-c.GetCh:
			req.RespCh <- c.get(req.Key)
		case req := <-c.DeleteCh:
			c.delete(req.Key)
			req.DoneCh <- struct{}{}
		}
	}
}

func (c *Core) processTTLQueue() {
	for {
		c.processTTL()
	}
}

func (c *Core) processTTL() {
	c.mu.Lock()
	lenSortedExpKey := len(c.sortedKeyExpiration)
	if lenSortedExpKey == 0 {
		c.mu.Unlock()
		time.Sleep(time.Second)
		return
	}
	firstKeyExpiration := c.sortedKeyExpiration[lenSortedExpKey-1]
	if firstKeyExpiration.Exp > time.Now().Unix() {
		c.mu.Unlock()
		nextTTLTime := time.Unix(firstKeyExpiration.Exp, 0)
		time.Sleep(nextTTLTime.Sub(time.Now()))
		return
	}

	c.sortedKeyExpiration = c.sortedKeyExpiration[:lenSortedExpKey-1]
	delete(c.data, firstKeyExpiration.Key)
	c.log.Infow("expired", "key", firstKeyExpiration.Key, "time", time.Now())
	c.mu.Unlock()
}

func findExpDataInsertingIndex(slice []ExpirationData, data ExpirationData) int {
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
