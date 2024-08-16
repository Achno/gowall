package mediancut

import (
	"container/heap"
	// "github.com/pkg/errors"
	"errors"
)

var ErrEmpty = errors.New("Empty")

// An queueItem is something we manage in a priority queue.
type queueItem struct {
	Value    ColorCube // The value of the item; arbitrary.
	Priority int       // The priority of the item in the queue.
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.
}

// A heapQueue implements heap.Interface and holds Items.
type heapQueue []*queueItem

func (hq heapQueue) Len() int { return len(hq) }

func (hq heapQueue) Less(i, j int) bool {
	return hq[i].Priority > hq[j].Priority
}

func (hq heapQueue) Swap(i, j int) {
	hq[i], hq[j] = hq[j], hq[i]
	hq[i].index = i
	hq[j].index = j
}

func (hq *heapQueue) Push(x interface{}) {
	n := len(*hq)
	item := x.(*queueItem)
	item.index = n
	*hq = append(*hq, item)
}

func (hq *heapQueue) Pop() interface{} {
	old := *hq
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	*hq = old[0 : n-1]
	return item
}

// update modifies the priority and value of an queueItem in the queue.
func (hq *heapQueue) update(item *queueItem, value ColorCube, priority int) {
	item.Value = value
	item.Priority = priority
	heap.Fix(hq, item.index)
}

type priorityQueue struct {
	queue heapQueue
}

// NewPriorityQueue return a maximum heap
func NewPriorityQueue(maxSize int) *priorityQueue {
	q := &priorityQueue{
		queue: make(heapQueue, 0, maxSize),
	}
	return q
}

func (pq *priorityQueue) Push(v ColorCube, priority int) {
	heap.Push(&pq.queue, &queueItem{
		Value:    v,
		Priority: priority,
	})
}

func (pq *priorityQueue) Pop() (ColorCube, int, error) {
	if pq.Empty() {
		return ColorCube{}, 0, ErrEmpty
	}
	item := heap.Pop(&pq.queue).(*queueItem)
	return item.Value, item.Priority, nil
}

func (pq *priorityQueue) Len() int {
	return pq.queue.Len()
}

func (pq *priorityQueue) Empty() bool {
	return pq.queue.Len() == 0
}

// Less use color count to sort
func (pq *priorityQueue) Less(i, j int) bool {
	return pq.queue[i].Value.Count > pq.queue[j].Value.Count
}

func (pq *priorityQueue) Swap(i, j int) {
	pq.queue[i], pq.queue[j] = pq.queue[j], pq.queue[i]
}
