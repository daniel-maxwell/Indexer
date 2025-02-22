package queue

import (
	"bytes"
	"testing"
)

// Tests creating a queue with a given capacity.
func TestCreateQueue(t *testing.T) {
	q, err := CreateQueue(3)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if q.capacity != 3 {
		t.Errorf("Expected queue size to be 3, got %d", q.capacity)
	}

	q, err = CreateQueue(1000000)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if q.capacity != 1000000 {
		t.Errorf("Expected queue size to be 1000000, got %d", q.capacity)
	}

	q, err = CreateQueue(0)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if q != nil {
		t.Errorf("Expected queue to be nil, got %v", q)
	}

	q, err = CreateQueue(-1)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if q != nil {
		t.Errorf("Expected queue to be nil, got %v", q)
	}
}

// Tests inserting elements into the queue.
func TestInsert(t *testing.T) {
	q, err := CreateQueue(3)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if q.Length() != 0 {
		t.Errorf("Expected queue length to be 0, got %d", q.Length())
	}

	err = q.Insert([]byte("a"))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if q.Length() != 1 {
		t.Errorf("Expected queue length to be 1, got %d", q.Length())
	}

	err = q.Insert([]byte("b"))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if q.Length() != 2 {
		t.Errorf("Expected queue length to be 2, got %d", q.Length())
	}

	err = q.Insert([]byte("c"))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if q.Length() != 3 {
		t.Errorf("Expected queue length to be 3, got %d", q.Length())
	}

	err = q.Insert([]byte("d"))
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if q.Length() != 3 {
		t.Errorf("Queue should be full, expected queue length to be 3, got %d", q.Length())
	}
}

// Tests removing elements from the queue.
func TestRemove(t *testing.T) {
	q, err := CreateQueue(3)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Insert byte slices
	q.Insert([]byte("a"))
	q.Insert([]byte("b"))
	q.Insert([]byte("c"))

	elem := q.Remove()
	if !bytes.Equal(elem, []byte("a")) {
		t.Errorf("Expected removed element to be 'a', got '%s'", elem)
	}
	if q.Length() != 2 {
		t.Errorf("Expected queue length to be 2, got %d", q.Length())
	}

	elem = q.Remove()
	if !bytes.Equal(elem, []byte("b")) {
		t.Errorf("Expected removed element to be 'b', got '%s'", elem)
	}
	if q.Length() != 1 {
		t.Errorf("Expected queue length to be 1, got %d", q.Length())
	}

	elem = q.Remove()
	if !bytes.Equal(elem, []byte("c")) {
		t.Errorf("Expected removed element to be 'c', got '%s'", elem)
	}
	if q.Length() != 0 {
		t.Errorf("Expected queue length to be 0, got %d", q.Length())
	}

	elem = q.Remove()
	if len(elem) != 0 {
		t.Errorf("Queue should be empty and return an empty slice, got '%s'", elem)
	}
	if q.Length() != 0 {
		t.Errorf("Expected queue length to be 0, got %d", q.Length())
	}
}

// Tests getting the length of the queue.
func TestLength(t *testing.T) {
	q, err := CreateQueue(3)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if q.Length() != 0 {
		t.Errorf("Expected queue length to be 0, got %d", q.Length())
	}

	err = q.Insert([]byte("a"))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if q.Length() != 1 {
		t.Errorf("Expected queue length to be 1, got %d", q.Length())
	}

	err = q.Insert([]byte("b"))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if q.Length() != 2 {
		t.Errorf("Expected queue length to be 2, got %d", q.Length())
	}

	err = q.Insert([]byte("c"))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if q.Length() != 3 {
		t.Errorf("Expected queue length to be 3, got %d", q.Length())
	}

	err = q.Insert([]byte("d"))
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if q.Length() != 3 {
		t.Errorf("Queue should be full, expected queue length to be 3, got %d", q.Length())
	}

	value := q.Remove()
	if !bytes.Equal(value, []byte("a")) {
		t.Errorf("Expected removed element to be 'a', got '%s'", value)
	}
	if q.Length() != 2 {
		t.Errorf("Expected queue length to be 2, got %d", q.Length())
	}

	value = q.Remove()
	if !bytes.Equal(value, []byte("b")) {
		t.Errorf("Expected removed element to be 'b', got '%s'", value)
	}
	if q.Length() != 1 {
		t.Errorf("Expected queue length to be 1, got %d", q.Length())
	}

	value = q.Remove()
	if !bytes.Equal(value, []byte("c")) {
		t.Errorf("Expected removed element to be 'c', got '%s'", value)
	}
	if q.Length() != 0 {
		t.Errorf("Expected queue length to be 0, got %d", q.Length())
	}

	value = q.Remove()
	if len(value) != 0 {
		t.Errorf("Queue should be empty and return an empty slice, got '%s'", value)
	}
}

// Tests checking if the queue is empty.
func TestIsEmpty(t *testing.T) {
	q, err := CreateQueue(3)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !q.IsEmpty() {
		t.Errorf("Expected queue to be empty")
	}
	q.Insert([]byte("a"))
	if q.IsEmpty() {
		t.Errorf("Expected queue to not be empty")
	}
	q.Remove()
	if !q.IsEmpty() {
		t.Errorf("Expected queue to be empty again")
	}
}
