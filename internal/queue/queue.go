package queue

import (
	"container/heap"
	"sync"
	"time"
)

const MaxQueueSize = 1000 // or any limit you want

// RetryRequest represents a queued request
type RetryRequest struct {
	Timestamp time.Time
	Method    string
	URL       string
	Body      []byte
	Headers   map[string]string
	Priority  int
	Index     int
}

// PriorityQueue implements heap.Interface
type PriorityQueue []*RetryRequest

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	if pq[i].Priority == pq[j].Priority {
		return pq[i].Timestamp.Before(pq[j].Timestamp)
	}
	return pq[i].Priority > pq[j].Priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

func (pq *PriorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*RetryRequest)
	item.Index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

var (
	pq  PriorityQueue
	mux sync.Mutex
)

func Init() {
	heap.Init(&pq)
}

func EnqueueRetry(req RetryRequest) bool {
	mux.Lock()
	defer mux.Unlock()

	if pq.Len() >= MaxQueueSize {
		return false
	}

	heap.Push(&pq, &req)
	return true
}

func DequeueRetry() *RetryRequest {
	mux.Lock()
	defer mux.Unlock()
	if pq.Len() == 0 {
		return nil
	}
	return heap.Pop(&pq).(*RetryRequest)
}
