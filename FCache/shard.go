package fcache

import (
	"container/list"
	"sync"

	lc "github.com/lingze/localCache"
)

type cacheShard struct {
	mu            sync.RWMutex
	maxEntrySize  int
	usedEntrySize int
	onEvicted     func(key string, value interface{})
	ll            *list.List
	cache         map[string]*list.Element
}

func newCacheSahrd(maxEntriesSize int, onEvicted func(key string, value interface{})) *cacheShard {
	return &cacheShard{
		maxEntrySize: maxEntriesSize,
		onEvicted:    onEvicted,
		ll:           list.New(),
		cache:        make(map[string]*list.Element),
	}
}

func (c *cacheShard) get(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if e, ok := c.cache[key]; ok {
		c.ll.MoveToBack(e)
		return e.Value.(*entry).value
	}

	return nil
}

func (c *cacheShard) set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if e, ok := c.cache[key]; ok {
		e.Value.(*entry).value = value
		c.ll.MoveToBack(e)
	}

	en := &entry{key: key, value: value}

	for c.usedEntrySize+en.Len() > c.maxEntrySize && c.maxEntrySize > 0 {
		c.delOldest()
	}

	e := c.ll.PushBack(en)
	c.cache[key] = e

}

func (c *cacheShard) del(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if e, ok := c.cache[key]; ok {
		c.remove(e)
	}
}

func (c *cacheShard) delOldest() {
	if c.ll.Len() == 0 {
		return
	}
	c.remove(c.ll.Front())
}

func (c *cacheShard) remove(e *list.Element) {
	if e == nil {
		return
	}
	c.ll.Remove(e)
	en := e.Value.(*entry)
	delete(c.cache, en.key)

	c.usedEntrySize -= en.Len()

	if c.onEvicted != nil {
		c.onEvicted(en.key, en.value)
	}
}
func (c *cacheShard) Len() int {
	return c.ll.Len()
}

type entry struct {
	key   string
	value interface{}
}

func (e *entry) Len() int {
	return lc.CalcLen(e)
}
