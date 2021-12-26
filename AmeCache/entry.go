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
	| EntryHeader | EmptyFlagSize | Timestamp |KeySize | Key | Value |
*/

const (
	EntryHeaderSize = 4 // int32 include all Entry Length
	EmptyFlagSize   = 1
	TimestampSize   = 8
	KeySizeInBytes  = 2 // int16 key length
)

func EncodeEntry(t uint64, key string, value []byte, buffer *[]byte) []byte {
	//Cal Entry Header Size
	keysize := len([]byte(key))
	entrySize := EntryHeaderSize + TimestampSize + KeySizeInBytes + uint32(keysize) + uint32(len(value)) + EmptyFlagSize

	if entrySize > uint32(len(*buffer)) {
		*buffer = make([]byte, entrySize)
	}
	bu := *buffer
	// Input Data
	binary.BigEndian.PutUint32(bu, uint32(entrySize))

	bu[EntryHeaderSize] = uint8(1)
	binary.BigEndian.PutUint64(bu[EntryHeaderSize+EmptyFlagSize:], t)
	binary.BigEndian.PutUint16(bu[EntryHeaderSize+EmptyFlagSize+TimestampSize:], uint16(keysize))
	copy(bu[EntryHeaderSize+EmptyFlagSize+TimestampSize+KeySizeInBytes:], []byte(key))
	copy(bu[EntryHeaderSize+EmptyFlagSize+TimestampSize+KeySizeInBytes+keysize:], value)
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

func readKeyFromEntry(entry []byte) string {
	keysize := binary.BigEndian.Uint16(entry[EntryHeaderSize+EmptyFlagSize+TimestampSize:])
	idx := EntryHeaderSize + EmptyFlagSize + TimestampSize + KeySizeInBytes
	return string(entry[idx : idx+int(keysize)])
}

// func readHashFromEntry(data []byte) uint64 {
// 	return 0
// }
func readValueFromEntry(entry []byte) []byte {
	l := readEntryheader(entry)
	keysize := binary.BigEndian.Uint16(entry[EntryHeaderSize+EmptyFlagSize+TimestampSize:])
	idx := EntryHeaderSize + EmptyFlagSize + TimestampSize + KeySizeInBytes + keysize
	return entry[idx:l]
}

// func re

func resetEntry(entry []byte) {
	entry[EntryHeaderSize] = 0
}
