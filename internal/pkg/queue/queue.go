package queue

import (
	"errors"
	"indexer/internal/pkg/models"
	"sync"
)

type Queue struct {
    q        []models.PageData
    capacity int
    closed   bool
    mu       sync.Mutex
}

// First in, first out queue 
type FifoQueue interface {
    Insert()
    Remove()
}

// Creates an empty queue with a specified capacity
func CreateQueue(capacity int) (*Queue, error) {
    if capacity <= 0 {
        return nil, errors.New("capacity should be greater than 0")
    }
    return &Queue{
        q:        make([]models.PageData, 0, capacity),
        capacity: capacity,
        closed:   false,
    }, nil
}

// Inserts an item into the queue
func (q *Queue) Insert(item models.PageData) error {
    q.mu.Lock()
    defer q.mu.Unlock()
    if q.closed {
        return errors.New("queue is closed")
    }
    if len(q.q) < int(q.capacity) {
        q.q = append(q.q, item)
        return nil
    }
    return errors.New("queue is full")
}

// Removes the oldest element from the queue
func (q *Queue) Remove() (models.PageData, error) {
    q.mu.Lock()
    defer q.mu.Unlock()
    if len(q.q) > 0 {
        item := q.q[0]
        q.q = q.q[1:]
        return item, nil
    }
    return models.PageData{}, errors.New("Queue is empty")
}

// Returns the number of elements in the queue
func (q *Queue) Length() int {
    return len(q.q)
}

// Returns true if the queue is empty
func (q *Queue) IsEmpty() bool {
    return len(q.q) == 0
}

// Closes the queue, preventing further insertions
func (q *Queue) Close() {
    q.mu.Lock()
    defer q.mu.Unlock()
    q.closed = true
}
