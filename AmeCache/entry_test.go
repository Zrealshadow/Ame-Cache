package amecache

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	"github.com/matryer/is"
)

func TestEntryAPI(t *testing.T) {
	is := is.New(t)
	N := 2
	units := make([]*Unit, N)
	buffer := make([]byte, 10)
	entrySizeLists := make([]uint32, N)
	entries := make([][]byte, N)
	for i := 0; i < N; i++ {
		units[i] = &Unit{t: uint64(i), key: strconv.Itoa(i), value: []byte(strconv.Itoa(i))}
		buSlice := EncodeEntry(units[i].t, units[i].t, units[i].key, units[i].value, &buffer)
		entries[i] = make([]byte, len(buSlice))
		copy(entries[i], buSlice)
		entrySizeLists[i] = uint32(len(buSlice))
	}
	// check
	for i := 0; i < N; i++ {
		is.Equal(entrySizeLists[i], readEntryheader(entries[i]))
		is.Equal(true, checkEntryValid(entries[i]))
		is.Equal(units[i].t, readTimestampFromEntry(entries[i]))                 // Timestamp
		is.Equal(units[i].key, readKeyFromEntry(entries[i]))                     // Key
		is.Equal(string(units[i].value), string(readValueFromEntry(entries[i]))) // Value
	}

	// Random Data
	for i := 0; i < N; i++ {
		units[i] = RandomGenerate()
		buSlice := EncodeEntry(units[i].t, units[i].t, units[i].key, units[i].value, &buffer)
		entries[i] = make([]byte, len(buSlice))
		copy(entries[i], buSlice)
		resetEntry(entries[i])
		entrySizeLists[i] = uint32(len(buSlice))
	}

	// check
	for i := 0; i < N; i++ {
		is.Equal(entrySizeLists[i], readEntryheader(entries[i]))
		is.Equal(false, checkEntryValid(entries[i]))
		is.Equal(units[i].t, readTimestampFromEntry(entries[i])) // Timestamp
		is.Equal(units[i].t, readHashKeyFromEntry(entries[i]))
		is.Equal(units[i].key, readKeyFromEntry(entries[i]))                     // Key
		is.Equal(string(units[i].value), string(readValueFromEntry(entries[i]))) // Value
	}
}

type Unit struct {
	t     uint64
	key   string
	value []byte
}

func RandomGenerate() *Unit {
	return &Unit{
		t:     rand.Uint64(),
		key:   strconv.Itoa(rand.Int()),
		value: []byte(fmt.Sprintf("Value : %f", rand.ExpFloat64())),
	}
}
