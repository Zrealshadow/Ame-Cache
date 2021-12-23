package scache

import (
	"sync"

	lc "github.com/lingze/localCache"
)

const DefaultMaxBytes = 1 << 29

type SimpleCache struct {
	m          sync.RWMutex
	cache      lc.Cache
	nhit, nget int
}

func NewSimpleCache(cache lc.Cache) *SimpleCache {
	return &SimpleCache{
		cache: cache,
	}
}

func (sc *SimpleCache) Set(key string, value interface{}) {
	sc.m.Lock()
	defer sc.m.Unlock()
	sc.cache.Set(key, value)
}

func (sc *SimpleCache) Get(key string) interface{} {
	sc.m.RLock()
	defer sc.m.RUnlock()
	sc.nget++

	if sc.cache == nil {
		return nil
	}

	v := sc.cache.Get(key)

	if v != nil {
		sc.nhit++
	}
	return v
}

type Stat struct {
	NHit, NGet int
}

func (sc *SimpleCache) stat() *Stat {
	sc.m.RLock()
	defer sc.m.RUnlock()
	return &Stat{
		NHit: sc.nhit,
		NGet: sc.nget,
	}
}

type Getter interface {
	Get(key string) interface{}
}

// 和hanlderFunc 相同
// 可以传入一个GetFunc func(key string) interface{}， 也可以使用Get进行读取数据源数据
// 当然也可以穿入一个完成了Getter接口的struct
type GetFunc func(key string) interface{}

func (f GetFunc) Get(key string) interface{} {
	return f(key)
}

type SCache struct {
	mainCache *SimpleCache
	getter    Getter
}

func NewSCache(getter Getter, cache lc.Cache) *SCache {
	return &SCache{
		mainCache: &SimpleCache{cache: cache},
		getter:    getter,
	}
}

func (sc *SCache) Get(key string) interface{} {
	val := sc.mainCache.Get(key)

	if val != nil {
		return val
	}

	if sc.getter != nil {
		val = sc.getter.Get(key)
		if val == nil {
			return nil
		}

		sc.mainCache.Set(key, val)
		return val
	}
	return nil
}

func (sc *SCache) Stat() *Stat {
	return sc.mainCache.stat()
}

func (sc *SCache) Set(key string, value interface{}) {
	if value == nil {
		return
	}

	sc.mainCache.Set(key, value)
}
