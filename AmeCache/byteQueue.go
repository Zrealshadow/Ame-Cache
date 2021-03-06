package amecache

const (
	DefaultMaxByteSize        int = 1024 * 4 * 1024 // 4MB
	DefaultInitializeByteSize int = 1024 * 64       // 64KB
	LeftMargin                    = 0
)

type byteQueue struct {
	head, tail  int
	rightMargin int
	queue       []byte
	capacity    int
	maxCapacity int
	usedByte    int
	count       int
}

func newByteQueue(init_size int, max_size int) *byteQueue {
	if init_size == 0 {
		init_size = DefaultInitializeByteSize
	}

	if max_size == 0 {
		max_size = DefaultMaxByteSize
	}
	bq := &byteQueue{
		head:        LeftMargin,
		tail:        LeftMargin,
		rightMargin: LeftMargin,
		queue:       make([]byte, init_size),
		capacity:    init_size,
		maxCapacity: max_size,
		usedByte:    0,
	}
	return bq
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
	bq.usedByte += len(entry)
	return index, nil
}

// get the oldest entry
func (bq *byteQueue) Peek() ([]byte, error) {
	if bq.count == 0 {
		return nil, ErrEmptyEntries
	}
	entry, err := bq.Get(uint32(bq.head))
	if err != nil {
		return nil, err
	}
	// for !checkEntryValid(entry) {
	// 	// Skip Empty Entry
	// 	bq.Pop()
	// 	entry, err = bq.Get(uint32(bq.head))
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }
	return entry, nil
}

// remove the oldest entry
func (bq *byteQueue) Pop() error {
	entry, err := bq.Get(uint32(bq.head))
	if err != nil {
		return err
	}

	bq.usedByte -= len(entry)
	bq.count--
	bq.head += int(readEntryheader(entry))

	// should consider the special situation
	if bq.head == bq.rightMargin {
		bq.head = LeftMargin
		if bq.tail == bq.rightMargin {
			bq.tail = LeftMargin
		}
		bq.rightMargin = bq.tail
	}

	return nil
}

func (bq *byteQueue) locateNewEntry(need int) error {
	if bq.head <= bq.tail {
		// ????????????
		// ??????????????? tail ?????????
		if bq.tail+need <= bq.capacity {
			// ?????????bq.tail ?????????
			return nil
		} else if LeftMargin+need < bq.head {
			// ?????????tail???????????????head?????????????????????
			bq.tail = LeftMargin
			return nil
		} else {
			//
			return bq.allocateExtraMemory(need)
		}

	} else {
		// tail ??? head ??????
		if bq.tail+need+EntryMinLen < bq.head {
			// ??????????????????EntryMinLen????????????????????????
			// ?????????bq.tail ?????? , ??????????????????!
			return nil
		} else {
			// ????????????, ?????????????????????????????? head = LeftMargin = 1
			return bq.allocateExtraMemory(need)
		}
	}
}

func (bq *byteQueue) allocateExtraMemory(need int) error {
	oldCapacity := bq.capacity
	if bq.capacity+need > bq.maxCapacity {
		return ErrOutOfMemory
	}
	// double increase
	bq.capacity *= 2
	for bq.capacity < oldCapacity+need {
		bq.capacity *= 2
	}

	if bq.capacity > bq.maxCapacity {
		bq.capacity = bq.maxCapacity
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
		copy(bq.queue, oldqueue[:bq.rightMargin])
		// fix with an Empty Entry
		valueLen := bq.head - bq.tail - EntryMinLen
		entry := GenerateEmptyEntry(valueLen)
		copy(bq.queue[bq.tail:], entry)
		bq.head = LeftMargin
		bq.tail = bq.rightMargin
		bq.count += 1
		bq.usedByte += len(entry)
	}
	return nil
}
