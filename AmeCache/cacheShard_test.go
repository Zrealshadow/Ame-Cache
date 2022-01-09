package amecache

import (
	"fmt"
	"math/rand"
	"testing"
)

var posMap = make(map[string]uint32)

func TestCacheShardBasicOrder(t *testing.T) {
	evictedMap := make(map[string]string)
	inputMap := make(map[uint64]string)
	cs := newCacheShard(InitSize, MaxSize, func(key string, value interface{}) {
		evictedMap[key] = string(value.([]byte))
	})
	N := 20000
	FillUpCache(inputMap, cs, t, N, true)
	CheckCacheShardKV(inputMap, evictedMap, cs, t)
	// Del A Lot
}

func TestCacheShardBasicRandom(t *testing.T) {
	evictedMap := make(map[string]string)
	inputMap := make(map[uint64]string)
	cs := newCacheShard(InitSize, MaxSize, func(key string, value interface{}) {
		evictedMap[key] = string(value.([]byte))
	})
	N := 20000
	FillUpCache(inputMap, cs, t, N, false)
	CheckCacheShardKV(inputMap, evictedMap, cs, t)
	// Del A Lot
}

func GenerateRandomKVHash(usedKey map[uint64]string) (uint64, string, string) {
	khash := rand.Int63()
	_, ok := usedKey[uint64(khash)]
	for ok {
		khash := rand.Int63()
		_, ok = usedKey[uint64(khash)]
	}
	key := fmt.Sprintf("%d", khash)
	value := key
	usedKey[uint64(khash)] = value
	return uint64(khash), key, value
}

func FillUpCache(inputMap map[uint64]string, cs *cacheShard, t *testing.T, N int, ordered bool) {
	var kh uint64
	var k, v string
	for i := 0; i < N; i++ {
		if !ordered {
			kh, k, v = GenerateRandomKVHash(inputMap)
		} else {
			kh = uint64(i)
			k = fmt.Sprintf("%d", kh)
			v = k
			inputMap[uint64(i)] = v
		}

		err := cs.set(kh, k, []byte(v), true)
		//Fill up
		if err != nil {
			t.Fatalf("Push Err %s", err)
		}
		inputMap[kh] = v
	}

}

func CheckCacheShardKV(groundTrue map[uint64]string, evictedmap map[string]string, cs *cacheShard, t *testing.T) {
	fmt.Printf("Evicted Map %+v\n", evictedmap)
	for kh, v := range groundTrue {
		key := fmt.Sprintf("%d", kh)

		//DeBUG
		fmt.Printf("Check kh %d\n", kh)
		if _, ok := evictedmap[key]; ok {
			fmt.Printf("Khash exsit in evictedmap\n")
		}
		value, err := cs.get(kh, key)
		if err == nil {
			// Success Get
			if v != string(value) {
				t.Fatalf("khash %d Key %s Want Value %s Got Value %s", kh, key, v, string(value))
			}
			continue
		} else if err == ErrKeyNotExist {
			// find in evictedMap
			if value, ok := evictedmap[key]; ok {

				if v != string(value) {
					t.Fatalf("in EvictedMap khash %d Key %s Want Value %s Got Value %s", kh, key, v, string(value))
				}

				continue
			}
		}
		t.Fatalf("Missing Kh %d Err %s", kh, err)
	}
}
