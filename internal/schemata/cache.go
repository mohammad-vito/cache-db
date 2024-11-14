package schemata

type SetCacheIn struct {
	Key   string      `json:"key" binding:"required"`
	Value interface{} `json:"value" binding:"required"`
	TTL   int         `json:"ttl" binding:"required"`
}

type GetCacheOut struct {
	Key   string      `json:"key" binding:"required"`
	Value interface{} `json:"value" binding:"required"`
}
