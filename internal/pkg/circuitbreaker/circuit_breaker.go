package circuitbreaker

import (
    "errors"
    "sync"
    "time"
    
    "go.uber.org/zap"
    
    "indexer/internal/pkg/logger"
    "indexer/internal/pkg/metrics"
)

var (
    ErrCircuitOpen = errors.New("circuit breaker is open")
)

type CircuitBreaker struct {
    mutex            sync.Mutex
    failureCount     int
    lastFailure      time.Time
    resetTimeout     time.Duration
    failureThreshold int
    serviceName      string
    state            string // "closed", "open", "half-open"
}

func NewCircuitBreaker(serviceName string, failureThreshold int, resetTimeout time.Duration) *CircuitBreaker {
    cb := &CircuitBreaker{
        serviceName:      serviceName,
        failureThreshold: failureThreshold,
        resetTimeout:     resetTimeout,
        state:            "closed",
    }
    
    // Initialize metric with closed state (0)
    metrics.CircuitBreakerState.WithLabelValues(serviceName).Set(0)
    
    return cb
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
    cb.mutex.Lock()
    
    if cb.state == "open" {
        // Check if we should retry (half-open)
        if time.Since(cb.lastFailure) > cb.resetTimeout {
            cb.state = "half-open"
            metrics.CircuitBreakerState.WithLabelValues(cb.serviceName).Set(1)
            logger.Log.Info("Circuit half-open, allowing test request", 
                zap.String("service", cb.serviceName))
        } else {
            cb.mutex.Unlock()
            return ErrCircuitOpen
        }
    }
    
    cb.mutex.Unlock()
    
    // Execute the function
    err := fn()
    
    cb.mutex.Lock()
    defer cb.mutex.Unlock()
    
    if err != nil {
        cb.failureCount++
        cb.lastFailure = time.Now()
        
        if cb.state == "half-open" || cb.failureCount >= cb.failureThreshold {
            cb.state = "open"
            metrics.CircuitBreakerState.WithLabelValues(cb.serviceName).Set(2)
            logger.Log.Warn("Circuit opened due to failures", 
                zap.String("service", cb.serviceName),
                zap.Int("failures", cb.failureCount),
                zap.Time("until", cb.lastFailure.Add(cb.resetTimeout)))
        }
        
        return err
    }
    
    // Success - reset if we were in half-open state
    if cb.state == "half-open" {
        cb.state = "closed"
        cb.failureCount = 0
        metrics.CircuitBreakerState.WithLabelValues(cb.serviceName).Set(0)
        logger.Log.Info("Circuit closed after successful test", 
            zap.String("service", cb.serviceName))
    }
    
    return nil
}

func (cb *CircuitBreaker) State() string {
    cb.mutex.Lock()
    defer cb.mutex.Unlock()
    return cb.state
}