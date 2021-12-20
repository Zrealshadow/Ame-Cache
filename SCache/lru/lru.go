package lru

import (
	"container/list"

	scache "github.com/lingze/localCache/SCache"
)

type lru struct {
	maxBytes int
	// 缓存最大容量

	onEvicted func(key string, value interface{})

	usedBytes int

	ll *list.List

	cache map[string]*list.Element
}

func New(maxbytes int, onevicted func(key string, value interface{})) scache.Cache {
	c := &lru{
		maxBytes:  maxbytes,
		onEvicted: onevicted,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
	}
	return c
}

// -------------- Interface --------------- //
type entry struct {
	key   string
	value interface{}
}

func (e *entry) Len() int {
	return scache.CalcLen(e.value)
}

func (c *lru) Set(key string, value interface{}) {
	if e, ok := c.cache[key]; ok {
		c.ll.MoveToBack(e)
		en := e.Value.(*entry)

		//update size
		c.usedBytes = c.usedBytes - scache.CalcLen(en.value) + scache.CalcLen(value)
		en.value = value
		return
	}

	// non-hit
	en := &entry{key, value}
	e := c.ll.PushBack(en)
	c.cache[key] = e

	c.usedBytes += en.Len()
	if c.usedBytes > c.maxBytes && c.maxBytes > 0 {
		c.DelOldest()
		// Pop another record
	}
}

func (c *lru) Get(key string) interface{} {
	if e, ok := c.cache[key]; ok {
		c.ll.MoveToBack(e)
		return e.Value.(*entry).value
	}
	return nil
}

func (c *lru) Del(key string) {
	if e, ok := c.cache[key]; ok {
		// the key exist
		c.remove(e)
	}
}

func (c *lru) DelOldest() {
	if c.ll.Len() == 0 {
		return
	}
	c.remove(c.ll.Front())
}

func (c *lru) remove(e *list.Element) {
	if e == nil {
		return
	}

	c.ll.Remove(e)

	en := e.Value.(*entry)
	c.usedBytes = c.usedBytes - en.Len()
	delete(c.cache, en.key)

	if c.onEvicted != nil {
		c.onEvicted(en.key, en.value)
	}
}

func (c *lru) Len() int {
	return c.ll.Len()
}
