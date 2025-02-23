package processor

import (
    "crypto/sha256"
    "encoding/hex"
    "sync"
)

// Deduper describes how we check for near-duplicates.
type Deduper interface {
    IsDuplicate(signature string) bool
    StoreSignature(signature string)
}

// memoryDeduper is a naive in-memory store of signatures.
type memoryDeduper struct {
    mu         sync.RWMutex
    signatures map[string]struct{}
}

// NewDeduper returns a new Deduper instance backed by an in-memory map.
func NewDeduper() Deduper {
    return &memoryDeduper{
        signatures: make(map[string]struct{}),
    }
}

func (d *memoryDeduper) IsDuplicate(signature string) bool {
    d.mu.RLock()
    defer d.mu.RUnlock()

    _, found := d.signatures[signature]
    return found
}

func (d *memoryDeduper) StoreSignature(signature string) {
    d.mu.Lock()
    defer d.mu.Unlock()

    d.signatures[signature] = struct{}{}
}

// GenerateSignature is a helper to create a hash from text.
// Here, we do a simple SHA-256. 
func GenerateSignature(text string) string {
    sum := sha256.Sum256([]byte(text))
    return hex.EncodeToString(sum[:])
}
