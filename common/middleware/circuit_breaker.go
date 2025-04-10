package middleware

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

// CircuitBreakerState 熔断器状态
type CircuitBreakerState int

const (
	StateClosed   CircuitBreakerState = iota // 关闭状态（允许请求）
	StateHalfOpen                            // 半开状态（允许部分请求）
	StateOpen                                // 开启状态（阻止请求）
)

const (
	DefaultFailureThreshold  = 5                  // 默认失败阈值
	DefaultHalfOpenTimeout   = 30 * time.Second   // 默认半开状态超时时间
	DefaultWindowSize        = 10 * time.Second   // 默认统计窗口大小
	DefaultSuccessThreshold  = 2                  // 默认成功阈值
	DefaultRedisKeyPrefix    = "circuit_breaker:" // 默认Redis key前缀
	DefaultRedisLockDuration = 10 * time.Second   // 默认Redis锁持续时间
)

// CircuitBreaker 熔断器结构体
type CircuitBreaker struct {
	name             string
	state            CircuitBreakerState
	failureThreshold int
	successThreshold int
	halfOpenTimeout  time.Duration
	windowSize       time.Duration
	redisClient      *redis.Client
	mutex            sync.RWMutex
	lastStateChange  time.Time
}

// NewCircuitBreaker 创建新的熔断器实例
func NewCircuitBreaker(name string, redisClient *redis.Client, options ...Option) *CircuitBreaker {
	cb := &CircuitBreaker{
		name:             name,
		state:            StateClosed,
		failureThreshold: DefaultFailureThreshold,
		successThreshold: DefaultSuccessThreshold,
		halfOpenTimeout:  DefaultHalfOpenTimeout,
		windowSize:       DefaultWindowSize,
		redisClient:      redisClient,
		lastStateChange:  time.Now(),
	}

	// 应用自定义选项
	for _, option := range options {
		option(cb)
	}

	return cb
}

// Option 熔断器配置选项
type Option func(*CircuitBreaker)

// WithFailureThreshold 设置失败阈值
func WithFailureThreshold(threshold int) Option {
	return func(cb *CircuitBreaker) {
		cb.failureThreshold = threshold
	}
}

// WithSuccessThreshold 设置成功阈值
func WithSuccessThreshold(threshold int) Option {
	return func(cb *CircuitBreaker) {
		cb.successThreshold = threshold
	}
}

// WithHalfOpenTimeout 设置半开状态超时时间
func WithHalfOpenTimeout(timeout time.Duration) Option {
	return func(cb *CircuitBreaker) {
		cb.halfOpenTimeout = timeout
	}
}

// WithWindowSize 设置统计窗口大小
func WithWindowSize(size time.Duration) Option {
	return func(cb *CircuitBreaker) {
		cb.windowSize = size
	}
}

// Execute 执行被保护的函数
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	if !cb.allowRequest() {
		return fmt.Errorf("circuit breaker is open")
	}

	err := fn()
	cb.recordResult(err == nil)
	return err
}

// allowRequest 判断是否允许请求通过
func (cb *CircuitBreaker) allowRequest() bool {
	cb.mutex.RLock()
	state := cb.state
	lastChange := cb.lastStateChange
	cb.mutex.RUnlock()

	switch state {
	case StateClosed:
		return true
	case StateOpen:
		lt := time.Since(lastChange)
		if lt > cb.halfOpenTimeout {
			cb.mutex.Lock()
			defer cb.mutex.Unlock()
			// 重新检查状态，因为可能在获取锁的过程中状态已经改变
			if cb.state == StateOpen {
				cb.setState(StateHalfOpen)
				return true
			}
		}
		return false
	case StateHalfOpen:
		// TODO 熔断逻辑
		return true
	default:
		return false
	}
}

// recordResult 记录请求结果
func (cb *CircuitBreaker) recordResult(success bool) {
	key := cb.getRedisKey()
	ctx := context.Background()

	// 增加总请求计数
	cb.redisClient.Incr(ctx, key+":total")

	if success {
		// 记录成功
		cb.redisClient.Incr(ctx, key+":success")
		if cb.state == StateHalfOpen {
			successCount, _ := cb.redisClient.Get(ctx, key+":success").Int()
			if successCount >= cb.successThreshold {
				cb.setState(StateClosed)
			}
		}
	} else {
		// 记录失败
		cb.redisClient.Incr(ctx, key+":failure")
		if cb.state == StateClosed {
			// 获取当前窗口内的总请求数和失败数
			totalCount, _ := cb.redisClient.Get(ctx, key+":total").Int()
			failureCount, _ := cb.redisClient.Get(ctx, key+":failure").Int()

			// 计算失败率
			if totalCount > 0 && float64(failureCount)/float64(totalCount) >= 0.5 && failureCount >= cb.failureThreshold {
				cb.setState(StateOpen)
			}
		}
	}

	// 设置过期时间
	cb.redisClient.Expire(ctx, key+":total", cb.windowSize)
	cb.redisClient.Expire(ctx, key+":success", cb.windowSize)
	cb.redisClient.Expire(ctx, key+":failure", cb.windowSize)
}

// setState 设置熔断器状态
func (cb *CircuitBreaker) setState(newState CircuitBreakerState) {
	// 注意：调用此方法前必须已经持有写锁
	if cb.state != newState {
		cb.state = newState
		cb.lastStateChange = time.Now()
		// 记录状态变更
		logx.Infof("Circuit breaker %s state changed to %v", cb.name, newState)
		// 重置计数器
		key := cb.getRedisKey()
		ctx := context.Background()
		cb.redisClient.Del(ctx, key+":success", key+":failure")
	}
}

// getRedisKey 获取Redis键名
func (cb *CircuitBreaker) getRedisKey() string {
	return DefaultRedisKeyPrefix + cb.name
}
