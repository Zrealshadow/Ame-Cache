package fcache_test

import (
	"testing"

	fcache "github.com/lingze/localCache/FCache"
	"github.com/matryer/is"
)

func TestFastCacheGet(t *testing.T) {
	db := map[string]string{
		"key1": "val1",
		"key2": "val2",
		"key3": "val3",
		"key4": "val4",
	}

	cache := fcache.NewFastCache(0, 64, nil)

	is := is.New(t)

	// var wg sync.WaitGroup
	// for k, v := range db {
	// 	wg.Add(1)
	// 	go func(k, v string) {
	// 		defer wg.Done()

	// 	}(k, v)
	// }
	for k, v := range db {
		cache.Set(k, v)
	}
	for k, v := range db {
		is.Equal(cache.Get(k).(string), v)
	}

}

func TestFastCacheLRU(t *testing.T) {

	is := is.New(t)

	EvictedKeys := make([]string, 0)

	onEvicted := func(key string, value interface{}) {
		EvictedKeys = append(EvictedKeys, key)
	}
	cache := fcache.NewFastCache(16, 1, onEvicted)

	cache.Set("Key1", 1)
	cache.Set("Key2", 2)
	is.Equal(cache.Get("Key1"), 1)
	cache.Set("Key3", 3)
	is.Equal(cache.Get("Key2"), nil)
	is.Equal(EvictedKeys[0], "Key2")
	cache.Set("Key4", 4)
	is.Equal(EvictedKeys[1], "Key1")
}
