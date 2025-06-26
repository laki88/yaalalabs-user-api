package utils

import (
	"container/heap"
	"user-ws-api/models"
)

type orderHeap struct {
	orders   []models.Order
	lessFunc func(a, b models.Order) bool
}

type OrderHeapQueue struct {
	h *orderHeap
}

func NewOrderHeapQueue(lessFunc func(a, b models.Order) bool) *OrderHeapQueue {
	h := &orderHeap{
		orders:   []models.Order{},
		lessFunc: lessFunc,
	}
	heap.Init(h)
	return &OrderHeapQueue{h: h}
}

// OrderHeapQueue API

func (q *OrderHeapQueue) Push(order models.Order) {
	heap.Push(q.h, order)
}

func (q *OrderHeapQueue) Pop() models.Order {
	return heap.Pop(q.h).(models.Order)
}

func (q *OrderHeapQueue) Peek() (models.Order, bool) {
	if len(q.h.orders) == 0 {
		return models.Order{}, false
	}
	return q.h.orders[0], true
}

func (q *OrderHeapQueue) Len() int {
	return len(q.h.orders)
}

// orderHeap (heap.Interface)

func (h *orderHeap) Len() int { return len(h.orders) }

func (h *orderHeap) Less(i, j int) bool {
	return h.lessFunc(h.orders[i], h.orders[j])
}

func (h *orderHeap) Swap(i, j int) {
	h.orders[i], h.orders[j] = h.orders[j], h.orders[i]
}

func (h *orderHeap) Push(x any) {
	h.orders = append(h.orders, x.(models.Order))
}

func (h *orderHeap) Pop() any {
	n := len(h.orders)
	item := h.orders[n-1]
	h.orders = h.orders[0 : n-1]
	return item
}
