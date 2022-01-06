package amecache

import (
	"encoding/binary"
)

// type entry struct {
// 	timestamp time.Time
// 	key       uint64
// 	value     []byte
// 	empty bool
// }

// func () encode

// try to not to claim new object

// Entry Structure

/*
	| EntryHeader | EmptyFlagSize | Timestamp | HashKey |KeySize | Key | Value |
*/

const (
	EntryHeaderSize = 4 // int32 include all Entry Length
	EmptyFlagSize   = 1
	TimestampSize   = 8
	HashKeySize     = 8 //64/8 = 8
	KeySizeInBytes  = 2 // int16 key length
	EntryMinLen     = EntryHeaderSize + EmptyFlagSize
)

func EncodeEntry(t uint64, khash uint64, key string, value []byte, buffer *[]byte) []byte {
	//Cal Entry Header Size
	keysize := len([]byte(key))
	entrySize := EntryHeaderSize + TimestampSize + HashKeySize + KeySizeInBytes + uint32(keysize) + uint32(len(value)) + EmptyFlagSize

	if entrySize > uint32(len(*buffer)) {
		*buffer = make([]byte, entrySize)
	}
	bu := *buffer
	// Input Data
	L := EntryHeaderSize
	binary.BigEndian.PutUint32(bu, uint32(entrySize))
	bu[EntryHeaderSize] = uint8(1)
	L += EmptyFlagSize
	binary.BigEndian.PutUint64(bu[L:], t)
	L += TimestampSize
	binary.BigEndian.PutUint64(bu[L:], khash)
	L += HashKeySize
	binary.BigEndian.PutUint16(bu[L:], uint16(keysize))
	L += KeySizeInBytes
	copy(bu[L:], []byte(key))
	L += keysize
	copy(bu[L:], value)

	// DEBUG
	// fmt.Printf("Entry Header : %v\n", bu[:EntryHeaderSize])
	// fmt.Printf("Entry Empty Flag : %v\n", bu[EntryHeaderSize])
	// fmt.Printf("Timestamp : %v\n", bu[EntryHeaderSize+EmptyFlagSize:EntryHeaderSize+EmptyFlagSize+TimestampSize])
	// fmt.Printf("KeySize  : %v\n", bu[EntryHeaderSize+EmptyFlagSize+TimestampSize:EntryHeaderSize+EmptyFlagSize+TimestampSize+KeySizeInBytes])
	// fmt.Printf("\n\n")
	return bu[:entrySize]
}

func readEntryheader(data []byte) uint32 {
	return binary.BigEndian.Uint32(data)
}

func checkEntryValid(entry []byte) bool {
	b := entry[EntryHeaderSize]
	return b == 1
}

func readTimestampFromEntry(entry []byte) uint64 {
	return binary.BigEndian.Uint64(entry[EntryHeaderSize+EmptyFlagSize:])
}

func readHashKeyFromEntry(entry []byte) uint64 {
	return binary.BigEndian.Uint64(entry[EntryHeaderSize+EmptyFlagSize+TimestampSize:])
}

func readKeyFromEntry(entry []byte) string {
	keysize := binary.BigEndian.Uint16(entry[EntryHeaderSize+EmptyFlagSize+TimestampSize+HashKeySize:])
	idx := EntryHeaderSize + EmptyFlagSize + TimestampSize + KeySizeInBytes + HashKeySize
	return string(entry[idx : idx+int(keysize)])
}

// func readHashFromEntry(data []byte) uint64 {
// 	return 0
// }
func readValueFromEntry(entry []byte) []byte {
	l := readEntryheader(entry)
	keysize := binary.BigEndian.Uint16(entry[EntryHeaderSize+EmptyFlagSize+TimestampSize+HashKeySize:])
	idx := EntryHeaderSize + EmptyFlagSize + TimestampSize + KeySizeInBytes + keysize + HashKeySize
	return entry[idx:l]
}

// func re

func resetEntry(entry []byte) {
	entry[EntryHeaderSize] = 0
}

func GenerateEmptyEntry(valueSize int) []byte {
	l := EntryMinLen + valueSize
	bu := make([]byte, l)
	binary.BigEndian.PutUint32(bu, uint32(l))
	bu[EntryHeaderSize] = uint8(0)
	return bu
}
