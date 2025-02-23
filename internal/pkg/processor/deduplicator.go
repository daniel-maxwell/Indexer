package processor

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
)

// Defines the interface for duplicate checking.
type Deduper interface {
	IsDuplicate(signature string) bool
	StoreSignature(signature string)
}

// A naive in-memory implementation of Deduper.
type memoryDeduper struct {
	mu         sync.RWMutex
	signatures map[string]struct{}
}

// Creates a new Deduper instance backed by an in-memory map.
func NewDeduper() Deduper {
	return &memoryDeduper{
		signatures: make(map[string]struct{}),
	}
}

// Checks if a signature has already been stored.
func (d *memoryDeduper) IsDuplicate(signature string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	_, found := d.signatures[signature]
	return found
}

// Stores the given signature.
func (d *memoryDeduper) StoreSignature(signature string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.signatures[signature] = struct{}{}
}

// Creates a SHA-256 hash of the provided text.
func GenerateSignature(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}
