package fcache

type Hasher interface {
	Sum64(key string) uint64
}

type FastCache struct {
	shards    []*cacheShard
	hash      Hasher
	shardMask uint64
}

func NewFastCache(maxEntriesSize int, shardsNum int, onEvicted func(key string, value interface{})) *FastCache {
	fc := &FastCache{
		hash:      newDefaultHasher(),
		shards:    make([]*cacheShard, shardsNum),
		shardMask: uint64(shardsNum - 1),
	}

	for i := 0; i < shardsNum; i++ {
		fc.shards[i] = newCacheShard(maxEntriesSize, onEvicted)
	}

	return fc
}

func (fc *FastCache) getShard(key string) *cacheShard {
	hid := fc.hash.Sum64(key)
	return fc.shards[hid&fc.shardMask]
}

func (fc *FastCache) Get(key string) interface{} {
	shard := fc.getShard(key)
	return shard.get(key)
}

func (fc *FastCache) Set(key string, value interface{}) {
	shard := fc.getShard(key)
	shard.set(key, value)
}

func (fc *FastCache) Del(key string) {
	shard := fc.getShard(key)
	shard.del(key)
}

func (fc *FastCache) Len() int {
	l := 0
	for _, shard := range fc.shards {
		l += shard.Len()
	}
	return l
}
