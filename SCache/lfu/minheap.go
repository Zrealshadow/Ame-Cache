package lfu

import "container/heap"

type pQueue []*entry

func (q pQueue) Len() int {
	return len(q)
}

func (pq pQueue) Less(i, j int) bool {
	return pq[i].weight < pq[j].weight
}

func (pq pQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *pQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*entry)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *pQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

func (pq *pQueue) update(item *entry, value interface{}, weight int) {
	item.value = value
	item.weight = weight
	heap.Fix(pq, item.index)
}
