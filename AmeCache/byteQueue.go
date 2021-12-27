package amecache

const (
	DefaultMaxByteSize        int = 1024 * 4 * 1024 // 4MB
	DefaultInitializeByteSize int = 1024 * 64       // 64KB
	LeftMargin                    = 1
)

type byteQueue struct {
	head, tail  int
	rightMargin int
	queue       []byte
	capacity    int
	usedByte    int
	count       int
	full        bool
}

func (bq *byteQueue) Get(offset uint32) ([]byte, error) {
	if offset < 0 || offset > uint32(bq.capacity) {
		return nil, ErrOutOfIndex
	}
	entrysize := readEntryheader(bq.queue[offset:])
	return bq.queue[offset : uint32(offset)+entrysize], nil
}

func (bq *byteQueue) Reset(offset uint32) error {
	if offset < 0 || offset > uint32(bq.capacity) {
		return ErrOutOfIndex
	}
	resetEntry(bq.queue[offset:])
	return nil
}

func (bq *byteQueue) Pop() ([]byte, error) {
	return nil, nil
}

func (bq *byteQueue) Push(entry []byte) (int, error) {
	entrySize := len(entry)

	if err := bq.locateNewEntry(entrySize); err != nil {
		return 0, err
	}
	index := bq.tail

	bq.tail += copy(bq.queue[bq.tail:], entry)

	// update margin
	if bq.tail > bq.head {
		bq.rightMargin = bq.tail
	}

	bq.count++

	return index, nil

}

func (bq *byteQueue) push(entry []byte) {

}

func (bq *byteQueue) locateNewEntry(need int) error {
	if bq.head <= bq.tail {
		// 正常排序
		// 检查能否从 tail 处插入
		if bq.tail+need <= bq.capacity {
			// 正常从bq.tail 处插入
			return nil
		} else if LeftMargin+need <= bq.head {
			// 无法从tail处插入，在head处找寻可插入的
			bq.tail = LeftMargin
			return nil
		} else {
			//
			return bq.allocateExtraMemory(need)
		}

	} else {
		// tail 在 head 之前
		if bq.tail+need < bq.head {
			//从正常bq.tail 插入 , 这里不能相等!
			return nil
		} else {
			// 寻求扩容, 扩容后会回到正序，且 head = LeftMargin = 1
			return bq.allocateExtraMemory(need)
		}
	}
}

func (bq *byteQueue) allocateExtraMemory(need int) error {
	bq.capacity += need
	if bq.capacity > DefaultMaxByteSize {
		return ErrOutOfMemory
	}
	// double increase
	bq.capacity *= 2
	if bq.capacity > DefaultMaxByteSize {
		bq.capacity = DefaultMaxByteSize
	}

	oldqueue := bq.queue
	bq.queue = make([]byte, bq.capacity)

	if bq.count == 0 {
		return nil
	}
	// re-allocate the position
	if bq.head < bq.tail {
		copy(bq.queue, oldqueue[:bq.rightMargin])
	} else {
		// tail < head
		copy(bq.queue, oldqueue[bq.head:bq.rightMargin])
		copy(bq.queue[bq.rightMargin-bq.head:], oldqueue[:bq.tail])
		bq.head = LeftMargin
		bq.tail = bq.usedByte + 1
	}
	return nil
}
