package amecache

type AmeCache struct {
	hash              Hasher
	shards            []*cacheShard
	shardMask         uint64 // shardNum - 1
	force             bool
	onEvicted         func(key string, value interface{})
	shardNum          int
	shardMaxByteSize  int
	shardInitByteSize int
}

const (
	defaultShardNum           int = 1024
	defaultMaxByteSize        int = 1024 * 1024 // 1MB
	defaultInitializeByteSize int = 1024 * 64   // 64KB
)

type Hasher interface {
	Sum64(key string) uint64
}

func NewAmeCache(options ...AmeCacheOption) *AmeCache {
	c := &AmeCache{
		shardNum:          defaultShardNum,
		shardMask:         uint64(defaultShardNum) - 1,
		shardMaxByteSize:  defaultMaxByteSize,
		shardInitByteSize: defaultInitializeByteSize,
		onEvicted:         nil,
		force:             false,
	}

	for _, opt := range options {
		opt(c)
	}

	c.shards = make([]*cacheShard, c.shardNum)
	for i := 0; i < c.shardNum; i++ {
		c.shards[i] = newCacheShard(c.shardInitByteSize, c.shardMaxByteSize, c.onEvicted)
	}
	return c
}

type AmeCacheOption func(*AmeCache)

func ShardsNum(shardNum int) AmeCacheOption {
	return func(ac *AmeCache) {
		ac.shardNum = shardNum
		ac.shardMask = uint64(shardNum - 1)
	}
}

func ShardMaxByteSize(maxByteSize int) AmeCacheOption {
	return func(ac *AmeCache) {
		ac.shardMaxByteSize = maxByteSize
	}
}

func ShardInitByteSize(initByteSize int) AmeCacheOption {
	return func(ac *AmeCache) {
		ac.shardInitByteSize = initByteSize
	}
}

func AddOnEvicted(onEvicted func(key string, value interface{})) AmeCacheOption {
	return func(ac *AmeCache) {
		ac.onEvicted = onEvicted
	}
}

func (c *AmeCache) ForceReplace() AmeCacheOption {
	return func(ac *AmeCache) {
		ac.force = true
	}
}

func (c *AmeCache) getShard(hkey uint64) *cacheShard {
	shardIdx := hkey & c.shardMask
	return c.shards[shardIdx]
}

func (c *AmeCache) Set(key string, value []byte) error {
	hkey := c.hash.Sum64(key)
	return c.getShard(hkey).set(hkey, key, value, c.force)
}

func (c *AmeCache) Get(key string) ([]byte, error) {
	hkey := c.hash.Sum64(key)
	return c.getShard(hkey).get(hkey, key)
}

func (c *AmeCache) Del(key string) error {
	hkey := c.hash.Sum64(key)
	return c.getShard(hkey).del(hkey, key)
}
