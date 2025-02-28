package deduper

import (
    "context"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "strings"
    "time"
    "indexer/internal/pkg/config"
    "indexer/internal/pkg/logger"
    "github.com/redis/go-redis/v9"
    "go.uber.org/zap"
)

// Defines the interface for duplicate checking.
type Deduper interface {
	IsDuplicate(signature string) bool
	StoreSignature(signature string)
}

// Implements the Deduper interface with Redis as the backing store.
type redisDeduper struct {
    client       *redis.Client
    redisKeyPrefix string
}

// Creates a new instance of redisDeduper.
// We store dedup signatures in a Redis SET, e.g. "deduper_signatures".
func NewRedisDeduper(config *config.Config) (Deduper, error) {
    rdb := redis.NewClient(&redis.Options{
        Addr:     fmt.Sprintf("%s:%s", config.RedisHost, config.RedisPort),
        Password: config.RedisPassword, // "" if no auth
        DB:       config.RedisDB,
    })

    // Test connection
    context, cancel := context.WithTimeout(context.Background(), 2 * time.Second)
    defer cancel()
    if err := rdb.Ping(context).Err(); err != nil {
        logger.Log.Error("Failed to connect to Redis", zap.Error(err))
        return nil, err
    }

    logger.Log.Info("Connected to Redis successfully",
        zap.String("host", config.RedisHost),
        zap.String("port", config.RedisPort),
    )

    return &redisDeduper{
        client:         rdb,
        redisKeyPrefix: "deduper_signatures", // could be configurable
    }, nil
}

// IsDuplicate checks if signature is in Redis.
func (redisDeduper *redisDeduper) IsDuplicate(signature string) bool {
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()

    exists, err := redisDeduper.client.SIsMember(ctx, redisDeduper.redisKeyPrefix, signature).Result()
    if err != nil {
        // If there's an error, assume not duplicate so we don't block indexing. 
        logger.Log.Error("Redis IsDuplicate check failed", zap.Error(err))
        return false
    }
    return exists
}

// Adds the signature to the Redis SET.
func (redisDeduper *redisDeduper) StoreSignature(signature string) {
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()
    if err := redisDeduper.client.SAdd(ctx, redisDeduper.redisKeyPrefix, signature).Err(); err != nil {
        logger.Log.Error("Failed to store signature in Redis", zap.Error(err))
    }
}

// Creates a SHA-256 hash of the text.
func GenerateSignature(text string) string {
    // A simple SHA-256 hash of the text
    sum := sha256.Sum256([]byte(strings.TrimSpace(text)))
    return hex.EncodeToString(sum[:])
}
