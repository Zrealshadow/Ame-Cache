package amecache_test

import (
	"fmt"
	"math/rand"
	"testing"

	acache "github.com/lingze/localCache/AmeCache"
)

func TestAmeCache(t *testing.T) {

	usedKey := make(map[string]string) // key -> value
	EvictedKey := make(map[string]int) // key -> Value
	onEvict := func(key string, value interface{}) {
		EvictedKey[key] = 1
	}

	N := 1000000 // Input K-V Count
	opts := []acache.AmeCacheOption{
		acache.AddOnEvictedOption(onEvict),
	}
	ac := acache.NewAmeCache(opts...)

	for i := 0; i < N; i++ {
		key, value := GenerateRandomKV(usedKey)
		err := ac.Set(key, value)
		if err == nil {
			// No Key Hash Collision
			usedKey[key] = string(value)
		}

	}

	for k, v := range usedKey {
		// ac.Get()
		v_get, err := ac.Get(k)
		if err == nil {
			if string(v_get) != v {
				t.Fatalf("Err in AmeCache, Key %s Want Value %s Got Value %s", k, v, v_get)
			}
		} else if err == acache.ErrKeyNotExist {
			if _, ok := EvictedKey[k]; !ok {
				t.Fatalf("Missing Key %s", k)
			}
		} else {
			t.Fatalf("throw unexpected error %s", err.Error())
		}
	}

}

func GenerateRandomKV(usedKey map[string]string) (string, []byte) {
	key := fmt.Sprint(rand.Int63())
	_, ok := usedKey[key]
	for ok {
		key = fmt.Sprint(rand.Int63())
		_, ok = usedKey[key]
	}
	size := rand.Int()%100 + len([]byte(key))
	value := make([]byte, size)
	copy(value, []byte(key))
	return key, value
}
