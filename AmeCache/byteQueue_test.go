package amecache

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

var buffer = make([]byte, 1000)

const InitSize = 1000
const MaxSize = 4000

func TestByteQueueBasicAPIWithoutAllocation(t *testing.T) {
	m := make(map[string]uint32)
	kv := make(map[string]string)
	bq := newByteQueue(InitSize, MaxSize)
	N := 9
	for i := 0; i < N; i++ {
		key := strconv.Itoa(int(i))
		value := make([]byte, 50)
		copy(value, []byte(key))
		entry := EncodeEntry(uint64(time.Now().Unix()), key, value, &buffer)
		pos, err := bq.Push(entry)
		if err != nil {
			panic(fmt.Sprintf("push err : %s", err.Error()))
		}
		m[key] = uint32(pos)
		kv[key] = string(value)
	}

	//Get
	for i := 0; i < N; i++ {
		want_key := strconv.Itoa(i)
		pos := m[want_key]
		entry, err := bq.Get(pos)
		if err != nil {
			panic(fmt.Sprintf("Get err : %s", err.Error()))
		}
		key := readKeyFromEntry(entry)
		value := readValueFromEntry(entry)
		if key != want_key || string(value) != kv[want_key] {
			t.Fatalf("Want Key %s  Got Key %s\nWant Value %s Got Value %s", want_key, key, kv[want_key], string(value))
		}
	}
	//Pop and Peek
	size := bq.count
	for i := 0; i < N-1; i++ {
		//Got Oldest
		want_key := strconv.Itoa(i)
		entry, err := bq.Peek()
		if err != nil {
			panic(fmt.Sprintf("Peek Error %s", err.Error()))
		}
		key := readKeyFromEntry(entry)
		value := readValueFromEntry(entry)
		if key != want_key || string(value) != kv[want_key] {
			t.Fatalf("Want Key %s  Got Key %s\nWant Value %s Got Value %s", want_key, key, kv[want_key], string(value))
		}
		err = bq.Pop()
		if err != nil {
			panic(fmt.Sprintf("Pop Error %s", err.Error()))
		}

		size--
		if size != bq.count {
			t.Fatal("No Pop entry")
		}

		// Get i + 1 Key
		want_key = strconv.Itoa(i + 1)
		entry, err = bq.Get(m[want_key])
		if err != nil {
			panic(fmt.Sprintf("Get Error %s", err.Error()))
		}
		key = readKeyFromEntry(entry)
		value = readValueFromEntry(entry)
		if key != want_key || string(value) != kv[want_key] {
			t.Fatalf("Want Key %s  Got Key %s\nWant Value %s Got Value %s", want_key, key, kv[want_key], string(value))
		}
	}

	// CheckAllKV(kv, m, bq, t)
}

func TestByteQueueBasicAllocation(t *testing.T) {
	m := make(map[string]uint32)
	kv := make(map[string]string)
	bq := newByteQueue(InitSize, MaxSize)
	// InitSize 1000 MaxSize 4000
	// fill up ByteQueue
	N := 10
	for i := 0; i < N; i++ {
		key := strconv.Itoa(i)
		entry, value := GenerateFixSizeEntry(100, key)
		kv[key] = value
		pos, err := bq.Push(entry)

		if err != nil {
			panic(fmt.Sprintf("push err : %s", err.Error()))
		}

		m[key] = uint32(pos)
	}
	//Check Capacity
	if InitSize != bq.capacity {
		t.Fatalf("Should Not Allocate")
	}
	//Push An Entry and Allocate New Memory
	key := strconv.Itoa(N)
	entry, value := GenerateFixSizeEntry(100, key)
	kv[key] = value
	pos, err := bq.Push(entry)

	if err != nil {
		panic(fmt.Sprintf("push err : %s", err.Error()))
	}
	m[key] = uint32(pos)

	if InitSize*2 != bq.capacity {
		t.Fatalf("Allocate Extra Memory Failed")
	}

	// Allocate Max Entry
	for i := N + 1; i < 40; i++ {
		key := strconv.Itoa(i)
		entry, value := GenerateFixSizeEntry(100, key)
		kv[key] = value
		pos, err := bq.Push(entry)

		if err != nil {
			panic(fmt.Sprintf("push err : %s", err.Error()))
		}

		m[key] = uint32(pos)
	}
	if bq.capacity != MaxSize {
		t.Fatal("Allocate Memory Error")
	}
	// CheckAllKV(kv, m, bq, t)
	//Random Insert a Entry, but Get ErrOutOfMemory
	entry, _ = GenerateFixSizeEntry(100, "123")
	_, err = bq.Push(entry)
	if err != ErrOutOfMemory {
		t.Fatal("No Error throw, due to out of max memory")
	}

}

func TestByteQueueBasicAPITailBeforeHead(t *testing.T) {
	m := make(map[string]uint32)
	kv := make(map[string]string)
	bq := newByteQueue(InitSize, MaxSize)
	// fill up ByteQueue
	N := 10
	for i := 0; i < N; i++ {
		key := strconv.Itoa(i)
		entry, value := GenerateFixSizeEntry(100, key)

		pos, err := bq.Push(entry)

		if err != nil {
			panic(fmt.Sprintf("push err : %s", err.Error()))
		}

		if i < N/2 {
			continue
		}

		kv[key] = value
		m[key] = uint32(pos)
	}

	// Pop 0-4
	for i := 0; i < N/2; i++ {
		err := bq.Pop()
		if err != nil {
			panic(fmt.Sprintf("Pop err : %s", err.Error()))
		}
	}
	// fmt.Printf("%+v\n", *bq)
	// Push
	for i := 0; i < N/2-1; i++ {
		key := strconv.Itoa(i + N)
		entry, value := GenerateFixSizeEntry(100, key)

		pos, err := bq.Push(entry)

		if err != nil {
			panic(fmt.Sprintf("push err : %s", err.Error()))
		}
		kv[key] = value
		m[key] = uint32(pos)
	}

	if bq.capacity != InitSize {
		t.Fatalf("byteQueue allocate extra memory, %d", bq.capacity)
	}

	// Allocate More Memory and redistribute the entry position
	for i := 0; i < N; i++ {
		key := strconv.Itoa(i + N)
		entry, value := GenerateFixSizeEntry(100, key)
		pos, err := bq.Push(entry)
		if err != nil {
			panic(fmt.Sprintf("push err : %s", err.Error()))
		}
		kv[key] = value
		m[key] = uint32(pos)
	}

	CheckAllKV(kv, m, bq, t)
}

func TestByteQueueCheckEmptyEntry(t *testing.T) {
	m := make(map[string]uint32)
	kv := make(map[string]string)
	bq := newByteQueue(400, MaxSize)
	N := 4
	// fill up byteQueue
	//  | 0:100 | 1:100 | 2:100 | 3:100|
	for i := 0; i < N; i++ {
		key := strconv.Itoa(i)
		entry, value := GenerateFixSizeEntry(100, key)

		pos, err := bq.Push(entry)

		if err != nil {
			panic(fmt.Sprintf("push err : %s", err.Error()))
		}

		if i < N/2 {
			continue
		}

		kv[key] = value
		m[key] = uint32(pos)
	}

	// remove Oldest
	// | | |2:100| 3 :100|
	for i := 0; i < N/2; i++ {
		err := bq.Pop()
		if err != nil {
			panic(fmt.Sprintf("pop err : %s", err.Error()))
		}
	}

	// Push to trail
	// |4 :100| EmptyEntry | 2 :100| 3:100| 5: 100|
	for i := N; i < N+2; i++ {
		key := strconv.Itoa(i)
		entry, value := GenerateFixSizeEntry(100, key)

		pos, err := bq.Push(entry)

		if err != nil {
			panic(fmt.Sprintf("push err : %s", err.Error()))
		}

		if i < N/2 {
			continue
		}
		kv[key] = value
		m[key] = uint32(pos)
	}

	key_order := []int{4, -1, 2, 3, 5}
	CheckOrder(key_order, bq, t)
}

func GenerateFixSizeEntry(l int, key string) ([]byte, string) {
	// EntryLen := 4 + 1 + 8 + 2 + keysize + valueSize
	keysize := len([]byte(key))
	if l < 15+keysize {
		return make([]byte, 0), ""
	}
	value := make([]byte, l-15-keysize)
	copy(value, []byte(key))
	entry := EncodeEntry(uint64(time.Now().Unix()), key, value, &buffer)
	return entry, string(value)
}

func CheckAllKV(kv map[string]string, m map[string]uint32, bq *byteQueue, t *testing.T) {
	for key, value := range kv {
		pos := m[key]
		entry, err := bq.Get(pos)
		if err != nil {
			t.Fatalf("Get Error : %s", err.Error())
		}

		got_key := readKeyFromEntry(entry)
		got_value := readValueFromEntry(entry)

		if key != got_key || value != string(got_value) {
			t.Fatalf("Want Key %s  Got Key %s\nWant Value %s Got Value %s", key, got_key, value, string(got_value))
		}
	}
}

// It will pop element
func CheckOrder(order []int, bq *byteQueue, t *testing.T) {
	for idx, k := range order {
		key := strconv.Itoa(k)
		entry, err := bq.Peek()

		if err != nil {
			panic(err)
		}

		if checkEntryValid(entry) {
			key_got := readKeyFromEntry(entry)
			if key_got != key {
				t.Fatalf("[idx %d] Want Key %s, Got Key %s", idx, key, key_got)
			}
			bq.Pop()
			continue
		}

		// Entry Invalid
		if k != -1 {
			t.Fatalf("[idx %d] Missing Empty Entry", idx)
		}
		bq.Pop()
	}
}
