package deduper

import (
	"context"
	"testing"
	"time"
	"go.uber.org/zap"
	"indexer/internal/pkg/config"
	"indexer/internal/pkg/logger"
)

func init() {
	logger.Log = zap.NewNop() // Set up a no-op logger to avoid nil pointer dereferences in tests.
}

// Validates that a new deduper instance connects to Redis,
// can store a signature, and then correctly identifies it as a duplicate.
func TestRedisDeduper(t *testing.T) {
	// Create a test configuration for Redis.
	config := &config.Config{
		RedisHost:     "localhost",
		RedisPort:     "6379",
		RedisPassword: "",
		RedisDB:       0,
	}

	// Create a new redisDeduper.
	deduper, err := NewRedisDeduper(config)
	if err != nil {
		t.Fatalf("Failed to create Redis deduper: %v", err)
	}

	// Clear the Redis set used for deduplication before testing.
	context, cancel := context.WithTimeout(context.Background(), 2 * time.Second)
	defer cancel()
	redisDeduper, ok := deduper.(*redisDeduper)
	if !ok {
		t.Fatal("Type assertion to *redisDeduper failed")
	}
	if err := redisDeduper.client.Del(context, redisDeduper.redisKeyPrefix).Err(); err != nil {
		t.Fatalf("Failed to clear Redis set: %v", err)
	}

	signature := "testsignature"

	// Initially, the signature should not be detected as duplicate.
	if deduper.IsDuplicate(signature) {
		t.Error("Expected signature not to be duplicate initially")
	}

	// Store the signature.
	deduper.StoreSignature(signature)

	// Give Redis a moment to persist the signature.
	time.Sleep(100 * time.Millisecond)

	// Now the signature should be detected as duplicate.
	if !deduper.IsDuplicate(signature) {
		t.Error("Expected signature to be detected as duplicate after storing")
	}
}
