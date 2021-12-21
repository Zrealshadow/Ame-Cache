package scache

import (
	"fmt"
	"runtime"
)

type Cache interface {
	Set(key string, value interface{})
	Get(key string) interface{}
	Del(key string)
	DelOldest()
	Len() int
}

// 计算内存的
func CalcLen(value interface{}) int {
	var n int
	switch v := value.(type) {
	case Value:
		n = v.Len()
	case string:
		if runtime.GOARCH == "amd64" {
			n = 16 + len(v)
		} else {
			n = 8 + len(v)
		}
	case bool, uint8, int8:
		n = 1
	case int16, uint16:
		n = 2
	case int32, uint32, float32:
		n = 4
	case int64, uint64, float64:
		n = 8
	case int, uint:
		if runtime.GOARCH == "amd64" {
			n = 8
		} else {
			n = 4
		}
	case complex64:
		n = 8
	case complex128:
		n = 16
	case []uint8:
		n = len(v)
	default:
		panic(fmt.Sprintf("%T is not implement cache.Value", value))
	}
	return n
}

type Value interface {
	Len() int
}
