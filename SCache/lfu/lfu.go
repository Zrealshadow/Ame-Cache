package lfu

import (
	"container/heap"

	scache "github.com/lingze/localCache/SCache"
)

type lfu struct {
	maxBytes int

	onEvicted func(key string, value interface{})

	usedBytes int

	pqueue *pQueue

	cache map[string]*entry
}

type entry struct {
	key    string
	value  interface{}
	weight int
	index  int
}

func (e *entry) Len() int {
	return scache.CalcLen(e.value)
}

func New(maxBytes int, onEvicted func(key string, value interface{})) scache.Cache {
	q := make(pQueue, 0, 1024)
	c := &lfu{
		maxBytes:  maxBytes,
		onEvicted: onEvicted,
		cache:     make(map[string]*entry),
		pqueue:    &q,
	}

	return c
}

func (c *lfu) Set(key string, value interface{}) {
	if e, ok := c.cache[key]; ok {
		// exist
		c.usedBytes = c.usedBytes - e.Len() + scache.CalcLen(value)
		c.pqueue.update(e, value, e.weight+1)
		return
	}

	en := &entry{
		key:   key,
		value: value,
	}

	heap.Push(c.pqueue, en)

	c.cache[key] = en

	c.usedBytes += en.Len()

	if c.usedBytes > c.maxBytes && c.maxBytes > 0 {
		c.DelOldest()
	}

}

func (c *lfu) Get(key string) interface{} {
	if e, ok := c.cache[key]; ok {
		return e.value
	}
	return nil
}

func (c *lfu) Del(key string) {
	if e, ok := c.cache[key]; ok {
		heap.Remove(c.pqueue, e.index)
		c.remove(e)
	}
	return
}

func (c *lfu) DelOldest() {
	if c.pqueue.Len() == 0 {
		return
	}
	e := heap.Pop(c.pqueue)
	c.remove(e.(*entry))
}

func (c *lfu) remove(e *entry) {
	if e == nil {
		return
	}

	delete(c.cache, e.key)
	c.usedBytes = c.usedBytes - e.Len()

	if c.onEvicted != nil {
		c.onEvicted(e.key, e.value)
	}
}

func (c *lfu) Len() int {
	return c.pqueue.Len()
}
