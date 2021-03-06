package benchmark__test

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/allegro/bigcache/v2"
	amecache "github.com/lingze/localCache/AmeCache"
	fcache "github.com/lingze/localCache/FCache"
	scache "github.com/lingze/localCache/SCache"
	"github.com/lingze/localCache/SCache/lru"
)

const maxEntrySize = 256

// -------- Helper Tool --------- //

func key(i int) string {
	return fmt.Sprintf("key-%010d", i)
}

func value() []byte {
	return V(make([]byte, 100))
}

type V []byte

func (v V) Len() int {
	return len(v)
}

func parallelKey(threadID int, counter int) string {
	return fmt.Sprintf("key-%04d-%06d", threadID, counter)
}

func initBigCache(entriesInWindow int) *bigcache.BigCache {
	cache, _ := bigcache.NewBigCache(bigcache.Config{
		Shards:             256,
		LifeWindow:         10 * time.Minute,
		MaxEntriesInWindow: entriesInWindow,
		MaxEntrySize:       maxEntrySize,
		Verbose:            true,
	})
	return cache
}

func initAmeCache(maxEntryCnt int) *amecache.AmeCache {
	maxSize, initSize := 1024*1024, 1024*64
	if maxEntryCnt*100 > maxSize {
		maxSize = maxEntryCnt * 100
	}

	if maxEntryCnt*25 > initSize {
		initSize = maxEntryCnt
	}
	opts := []amecache.AmeCacheOption{
		amecache.ShardsNumOption(256),
		amecache.ShardInitByteSizeOption(initSize),
		amecache.ShardMaxByteSizeOption(maxSize),
	}
	return amecache.NewAmeCache(opts...)
}

func initFastCache(maxEntriesSize int) *fcache.FastCache {
	return fcache.NewFastCache(maxEntrySize, 256, nil)
}

// --------------------   Set Performance Test ------------------//
func BenchmarkMapSet(b *testing.B) {
	m := make(map[string][]byte, b.N)
	for i := 0; i < b.N; i++ {
		m[key(i)] = value()
	}
}

func BenchmarkSimpleCacheSet(b *testing.B) {
	cache := scache.NewSCache(nil, lru.New(b.N*100, nil))
	for i := 0; i < b.N; i++ {
		cache.Set(key(i), value())
	}
}

func BenchmarkFastCacheSet(b *testing.B) {
	cache := initFastCache(b.N * 100)
	for i := 0; i < b.N; i++ {
		cache.Set(key(i), value())
	}
}

func BenchmarkConcurrentMapSet(b *testing.B) {
	var m sync.Map
	for i := 0; i < b.N; i++ {
		m.Store(key(i), value())
	}
}

func BenchmarkBigCacheSet(b *testing.B) {
	cache := initBigCache(b.N)
	for i := 0; i < b.N; i++ {
		_ = cache.Set(key(i), value())
	}
}

func BenchmarkAmeCacheSet(b *testing.B) {
	cache := initAmeCache(b.N)
	for i := 0; i < b.N; i++ {
		// fmt.Printf("QAQ\n")
		_ = cache.Set(key(i), value())
	}
}

// ------------------ Get Performance Test ---------------- //

func BenchmarkMapGet(b *testing.B) {
	b.StopTimer()
	m := make(map[string][]byte)
	for i := 0; i < b.N; i++ {
		m[key(i)] = value()
	}

	b.StartTimer()
	hitCount := 0
	for i := 0; i < b.N; i++ {
		if m[key(i)] != nil {
			hitCount++
		}
	}
}

func BenchmarkSimpleCacheGet(b *testing.B) {
	b.StopTimer()
	cache := scache.NewSCache(nil, lru.New(b.N*100, nil))
	for i := 0; i < b.N; i++ {
		cache.Set(key(i), value())
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(key(i))
	}
}

func BenchmarkFastCacheGet(b *testing.B) {
	b.StopTimer()
	cache := initFastCache(b.N * 100)
	for i := 0; i < b.N; i++ {
		cache.Set(key(i), value())
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(key(i))
	}
}

func BenchmarkConcurrentMapGet(b *testing.B) {
	b.StopTimer()
	var m sync.Map
	for i := 0; i < b.N; i++ {
		m.Store(key(i), value())
	}

	b.StartTimer()
	hitCounter := 0
	for i := 0; i < b.N; i++ {
		_, ok := m.Load(key(i))
		if ok {
			hitCounter++
		}
	}
}

func BenchmarkBigCacheGet(b *testing.B) {
	b.StopTimer()
	cache := initBigCache(b.N)
	for i := 0; i < b.N; i++ {
		_ = cache.Set(key(i), value())
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cache.Get(key(i))
	}
}

func BenchmarkAmeCacheGet(b *testing.B) {
	b.StopTimer()
	cache := initAmeCache(b.N)
	for i := 0; i < b.N; i++ {
		_ = cache.Set(key(i), value())
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cache.Get(key(i))
	}
}

// --------------- Parallel Set Performance Test ------------//

func BenchmarkSimpleCacheSetParallel(b *testing.B) {
	cache := scache.NewSCache(nil, lru.New(b.N*100, nil))
	rand.Seed(time.Now().Unix())

	b.RunParallel(func(pb *testing.PB) {
		id := rand.Intn(1000)
		counter := 0
		for pb.Next() {
			cache.Set(parallelKey(id, counter), value())
			counter = counter + 1
		}
	})
}

func BenchmarkFastCacheSetParallel(b *testing.B) {
	cache := initFastCache(b.N * 100)
	rand.Seed(time.Now().Unix())

	b.RunParallel(func(pb *testing.PB) {
		id := rand.Intn(1000)
		counter := 0
		for pb.Next() {
			cache.Set(parallelKey(id, counter), value())
			counter = counter + 1
		}
	})
}

func BenchmarkConcurrentMapSetParallel(b *testing.B) {
	var cache sync.Map
	rand.Seed(time.Now().Unix())
	b.RunParallel(func(pb *testing.PB) {
		tid := rand.Intn(1000)
		counter := 0
		for pb.Next() {
			cache.Store(parallelKey(tid, counter), value())
		}
	})
}

func BenchmarkBigCacheSetParallel(b *testing.B) {
	cache := initBigCache(b.N)
	rand.Seed(time.Now().Unix())

	b.RunParallel(func(pb *testing.PB) {
		id := rand.Intn(1000)
		counter := 0
		for pb.Next() {
			_ = cache.Set(parallelKey(id, counter), value())
			counter = counter + 1
		}
	})
}

func BenchmarkAmeCacheSetParallel(b *testing.B) {
	cache := initAmeCache(b.N)
	rand.Seed(time.Now().Unix())

	b.RunParallel(func(pb *testing.PB) {
		id := rand.Intn(1000)
		counter := 0
		for pb.Next() {
			_ = cache.Set(parallelKey(id, counter), value())
			counter = counter + 1
		}
	})
}

// --------------- Parallel Get Performance Test --------------- //

func BenchmarkSimpleCacheGetParallel(b *testing.B) {
	b.StopTimer()
	cache := scache.NewSCache(nil, lru.New(b.N*100, nil))

	for i := 0; i < b.N; i++ {
		cache.Set(key(i), value())
	}

	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			cache.Get(key(counter))
			counter = counter + 1
		}
	})
}

func BenchmarkFastCacheGetParallel(b *testing.B) {
	b.StopTimer()
	cache := initFastCache(b.N * 100)

	for i := 0; i < b.N; i++ {
		cache.Set(key(i), value())
	}

	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			cache.Get(key(counter))
			counter = counter + 1
		}
	})
}

func BenchmarkConcurrentMapGetParallel(b *testing.B) {
	b.StopTimer()
	var cache sync.Map

	for i := 0; i < b.N; i++ {
		cache.Store(key(i), value())
	}

	b.StartTimer()
	b.RunParallel(func(p *testing.PB) {
		counter := 0
		for p.Next() {
			cache.Load(key(counter))
			counter += 1
		}
	})
}

func BenchmarkBigCacheGetParallel(b *testing.B) {
	b.StopTimer()
	cache := initBigCache(b.N)
	for i := 0; i < b.N; i++ {
		_ = cache.Set(key(i), value())
	}

	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			_, _ = cache.Get(key(counter))
			counter = counter + 1
		}
	})
}

func BenchmarkAmeCacheGetParallel(b *testing.B) {
	b.StopTimer()
	cache := initAmeCache(b.N)
	for i := 0; i < b.N; i++ {
		_ = cache.Set(key(i), value())
	}

	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			_, _ = cache.Get(key(counter))
			counter = counter + 1
		}
	})
}

// --------------- Parallel Set Get Performance Test ----------- //
func BenchmarkSimpleCacheSetGetParallel(b *testing.B) {
	cache := scache.NewSCache(nil, lru.New(b.N*100, nil))
	tids := make([]int, 0, 10)
	rand.Seed(time.Now().Unix())
	b.RunParallel(func(pb *testing.PB) {
		tid := rand.Intn(1000)
		tids = append(tids, tid)
		counter := 0
		for pb.Next() {
			cache.Set(parallelKey(tid, counter), value())
			counter = counter + 1
		}
	})

	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		tid := tids[rand.Intn(len(tids))]
		for pb.Next() {
			cache.Get(parallelKey(tid, counter))
			counter = counter + 1
		}
	})
}

func BenchmarkFastCacheSetGetParallel(b *testing.B) {
	cache := initFastCache(b.N * 100)
	tids := make([]int, 0, 10)
	rand.Seed(time.Now().Unix())
	b.RunParallel(func(pb *testing.PB) {
		tid := rand.Intn(1000)
		tids = append(tids, tid)
		counter := 0
		for pb.Next() {
			cache.Set(parallelKey(tid, counter), value())
			counter = counter + 1
		}
	})

	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		tid := tids[rand.Intn(len(tids))]
		for pb.Next() {
			cache.Get(parallelKey(tid, counter))
			counter = counter + 1
		}
	})
}

func BenchmarkConcurrentSetGetParallel(b *testing.B) {
	var cache sync.Map
	tids := make([]int, 0, 10)
	rand.Seed(time.Now().Unix())
	b.RunParallel(func(pb *testing.PB) {
		tid := rand.Intn(1000)
		tids = append(tids, tid)
		counter := 0
		for pb.Next() {
			cache.Store(parallelKey(tid, counter), value())
			counter = counter + 1
		}
	})

	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		tid := tids[rand.Intn(len(tids))]
		for pb.Next() {
			cache.Load(parallelKey(tid, counter))
			counter = counter + 1
		}
	})
}

func BenchmarkBigCacheSetGetParallel(b *testing.B) {
	cache := initBigCache(b.N)
	tids := make([]int, 0, 10)
	rand.Seed(time.Now().Unix())
	b.RunParallel(func(pb *testing.PB) {
		tid := rand.Intn(1000)
		tids = append(tids, tid)
		counter := 0
		for pb.Next() {
			cache.Set(parallelKey(tid, counter), value())
			counter = counter + 1
		}
	})

	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		tid := tids[rand.Intn(len(tids))]
		for pb.Next() {
			_, _ = cache.Get(parallelKey(tid, counter))
			counter = counter + 1
		}
	})
}

func BenchmarkAmeCacheSetGetParallel(b *testing.B) {
	cache := initAmeCache(b.N)
	tids := make([]int, 0, 10)
	rand.Seed(time.Now().Unix())
	b.RunParallel(func(pb *testing.PB) {
		tid := rand.Intn(1000)
		tids = append(tids, tid)
		counter := 0
		for pb.Next() {
			cache.Set(parallelKey(tid, counter), value())
			counter = counter + 1
		}
	})

	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		tid := tids[rand.Intn(len(tids))]
		for pb.Next() {
			cache.Get(parallelKey(tid, counter))
			counter = counter + 1
		}
	})
}
