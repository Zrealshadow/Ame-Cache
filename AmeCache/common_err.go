package amecache

type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrHashCollision     = Error("Hash Collision")
	ErrEntrySizeOverflow = Error("Entry Size is bigger than whole cacheShard size")
	ErrKeyNotExist       = Error("Key is not exist")
	ErrOutOfIndex        = Error("Out of Index")
	ErrOutOfMemory       = Error("Out of Memory")
	ErrEmptyEntries      = Error("Empty Entries")
)
