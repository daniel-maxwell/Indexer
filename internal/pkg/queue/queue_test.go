package queue

import (
	"reflect"
	"testing"
	"indexer/internal/pkg/models"
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

	err = q.Insert(models.PageData{URL: "a"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if q.Length() != 1 {
		t.Errorf("Expected queue length to be 1, got %d", q.Length())
	}

	err = q.Insert(models.PageData{URL: "b"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if q.Length() != 2 {
		t.Errorf("Expected queue length to be 2, got %d", q.Length())
	}

	err = q.Insert(models.PageData{URL: "c"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if q.Length() != 3 {
		t.Errorf("Expected queue length to be 3, got %d", q.Length())
	}

	err = q.Insert(models.PageData{URL: "d"})
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

	// Insert PageData values
	if err := q.Insert(models.PageData{URL: "a"}); err != nil {
		t.Errorf("Insert error: %v", err)
	}
	if err := q.Insert(models.PageData{URL: "b"}); err != nil {
		t.Errorf("Insert error: %v", err)
	}
	if err := q.Insert(models.PageData{URL: "c"}); err != nil {
		t.Errorf("Insert error: %v", err)
	}

	elem, err := q.Remove()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if elem.URL != "a" {
		t.Errorf("Expected removed element URL to be 'a', got '%s'", elem.URL)
	}
	if q.Length() != 2 {
		t.Errorf("Expected queue length to be 2, got %d", q.Length())
	}

	elem, err = q.Remove()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if elem.URL != "b" {
		t.Errorf("Expected removed element URL to be 'b', got '%s'", elem.URL)
	}
	if q.Length() != 1 {
		t.Errorf("Expected queue length to be 1, got %d", q.Length())
	}

	elem, err = q.Remove()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if elem.URL != "c" {
		t.Errorf("Expected removed element URL to be 'c', got '%s'", elem.URL)
	}
	if q.Length() != 0 {
		t.Errorf("Expected queue length to be 0, got %d", q.Length())
	}

	elem, err = q.Remove()
	if err == nil {
		t.Errorf("Expected error when removing from empty queue, got nil")
	}
	if !reflect.DeepEqual(elem, models.PageData{}) {
		t.Errorf("Expected removed element to be zero value, got %v", elem)
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

	err = q.Insert(models.PageData{URL: "a"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if q.Length() != 1 {
		t.Errorf("Expected queue length to be 1, got %d", q.Length())
	}

	err = q.Insert(models.PageData{URL: "b"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if q.Length() != 2 {
		t.Errorf("Expected queue length to be 2, got %d", q.Length())
	}

	err = q.Insert(models.PageData{URL: "c"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if q.Length() != 3 {
		t.Errorf("Expected queue length to be 3, got %d", q.Length())
	}

	err = q.Insert(models.PageData{URL: "d"})
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if q.Length() != 3 {
		t.Errorf("Queue should be full, expected queue length to be 3, got %d", q.Length())
	}

	value, err := q.Remove()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if value.URL != "a" {
		t.Errorf("Expected removed element URL to be 'a', got '%s'", value.URL)
	}
	if q.Length() != 2 {
		t.Errorf("Expected queue length to be 2, got %d", q.Length())
	}

	value, err = q.Remove()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if value.URL != "b" {
		t.Errorf("Expected removed element URL to be 'b', got '%s'", value.URL)
	}
	if q.Length() != 1 {
		t.Errorf("Expected queue length to be 1, got %d", q.Length())
	}

	value, err = q.Remove()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if value.URL != "c" {
		t.Errorf("Expected removed element URL to be 'c', got '%s'", value.URL)
	}
	if q.Length() != 0 {
		t.Errorf("Expected queue length to be 0, got %d", q.Length())
	}

	value, err = q.Remove()
	if err == nil {
		t.Errorf("Expected error when removing from empty queue, got nil")
	}
	if !reflect.DeepEqual(value, models.PageData{}) {
		t.Errorf("Expected removed element to be zero value, got %v", value)
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
	q.Insert(models.PageData{URL: "a"})
	if q.IsEmpty() {
		t.Errorf("Expected queue to not be empty")
	}
	q.Remove()
	if !q.IsEmpty() {
		t.Errorf("Expected queue to be empty again")
	}
}
