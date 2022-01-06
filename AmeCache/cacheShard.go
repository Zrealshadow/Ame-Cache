package amecache

import (
	"sync"
	"time"
)

type cacheShard struct {
	mu        sync.RWMutex
	queue     byteQueue
	index     map[uint64]uint32
	onEvicted func(key string, value interface{})
	// lifewindows time.Duration
	EntryBuffer []byte

	// Debug
	// usedbyte int
	// count    int
}

func newCacheShard(init_size int, max_size int, onEvicted func(key string, value interface{})) *cacheShard {
	cs := &cacheShard{
		queue:       *newByteQueue(init_size, max_size),
		onEvicted:   onEvicted,
		index:       make(map[uint64]uint32),
		EntryBuffer: make([]byte, 0),
	}
	return cs
}

func (cs *cacheShard) set(khash uint64, key string, value []byte, force bool) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	t := time.Now().Unix()

	if offset, ok := cs.index[khash]; ok {
		// key exist
		entry, err := cs.queue.Get(offset)
		if err != nil {
			return err
		}
		if !force {
			ekey := readKeyFromEntry(entry)
			// hash collision
			if ekey != key {
				return ErrHashCollision
			}
		}
		cs.queue.Reset(offset)
	}

	entry := EncodeEntry(uint64(t), khash, key, value, &cs.EntryBuffer)

	for {
		if index, err := cs.queue.Push(entry); err == nil {
			// insert success
			cs.index[khash] = uint32(index)
			// cs.usedbyte += len(entry)
			// cs.count++
			// fmt.Printf("push key %s, EntrySize %d, (Head %d Tail %d, RightMargin %d), cs.count %d, usedbyte %d\n", key, len(entry), cs.queue.head, cs.queue.tail, cs.queue.rightMargin, cs.count, cs.usedbyte)
			return nil
		}
		// pop out the oldest for capacity
		if err := cs.delOldest(); err != nil {
			return ErrEntrySizeOverflow
		}
	}
}

func (cs *cacheShard) get(khash uint64, key string) ([]byte, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	if offset, ok := cs.index[khash]; ok {
		entry, err := cs.queue.Get(offset)
		if err != nil {
			return nil, err
		}
		ekey := readKeyFromEntry(entry)
		if ekey != key {
			return nil, ErrKeyNotExist
		}
		return readValueFromEntry(entry), nil
	}
	return nil, ErrKeyNotExist
}

func (cs *cacheShard) del(khash uint64, key string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if offset, ok := cs.index[khash]; ok {
		entry, err := cs.queue.Get(offset)
		if err != nil {
			return err
		}
		ekey := readKeyFromEntry(entry)
		if ekey != key {
			return nil
		}

		//Check the  Validation
		if checkEntryValid(entry) {
			return nil
		}

		cs.remove(khash, entry)
		resetEntry(entry)
	}
	return nil
}

func (cs *cacheShard) delOldest() error {
	// We have to consider empty entry
	entry, err := cs.queue.Peek()
	if err != nil {
		return err
	}
	// valid entry
	if checkEntryValid(entry) {
		khash := readHashKeyFromEntry(entry)
		cs.remove(khash, entry)
	}
	// empty entry
	// err = cs.queue.Pop()
	// if err == nil {
	// 	cs.count--
	// 	cs.usedbyte -= len(entry)
	// 	key := readKeyFromEntry(entry)
	// 	fmt.Printf("Pop key %s, EntrySize %d (Head %d Tail %d, rightmargin %d), cs.count %d, usedbyte %d |", key, len(entry), cs.queue.head, cs.queue.tail, cs.queue.rightMargin, cs.count, cs.usedbyte)
	// }
	// return err
	return cs.queue.Pop()
}

func (cs *cacheShard) remove(khash uint64, entry []byte) error {
	// pos := cs.index[khash]
	delete(cs.index, khash)
	if cs.onEvicted != nil {
		key := readKeyFromEntry(entry)
		value := readValueFromEntry(entry)
		cs.onEvicted(key, value)
	}
	return nil
}
