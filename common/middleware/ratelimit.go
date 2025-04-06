package middleware

import (
	"context"
	"sync"
	"time"

	"net/http"

	"grpc_demo_client/common/errs"
	redisClient "grpc_demo_client/common/redis"
	"grpc_demo_client/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

// TokenBucket 令牌桶结构体
type TokenBucket struct {
	rate     int64  // 令牌生成速率
	capacity int64  // 桶的容量
	key      string // Redis键名
	mutex    sync.Mutex
}

// NewTokenBucket 创建一个新的令牌桶
func NewTokenBucket(rate int64, capacity int64) *TokenBucket {
	return &TokenBucket{
		rate:     rate,
		capacity: capacity,
		key:      "token_bucket",
		mutex:    sync.Mutex{},
	}
}

// getTokens 获取当前可用的令牌数
func (tb *TokenBucket) getTokens() int64 {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	ctx := context.Background()
	// 获取当前令牌数和上次更新时间
	tokenKey := tb.key + ":tokens"
	tokens, err := redisClient.RedisClient.Get(ctx, tokenKey).Int64()
	if err != nil {
		tokens = tb.capacity
	}

	lastTimeUnix, err := redisClient.RedisClient.Get(ctx, tb.key+":last_time").Int64()
	if err != nil {
		lastTimeUnix = time.Now().Unix()
	}

	now := time.Now()
	nowUnix := now.Unix()
	elapsed := nowUnix - lastTimeUnix
	// 根据时间间隔生成新的令牌
	tokens = tokens + elapsed*tb.rate

	// 确保不超过桶的容量
	if tokens > tb.capacity {
		tokens = tb.capacity
	}

	// 更新Redis中的值
	err = redisClient.RedisClient.Set(ctx, tb.key+":tokens", tokens, 24*time.Hour).Err()
	if err != nil {
		logx.Error("redis set err:", err)
	}
	err = redisClient.RedisClient.Set(ctx, tb.key+":last_time", nowUnix, 24*time.Hour).Err()
	if err != nil {
		logx.Error("redis set err:", err)
	}

	return tokens
}

// Allow 判断是否允许请求通过
func (tb *TokenBucket) Allow() bool {
	tokens := tb.getTokens()
	if tokens >= 1 {
		tb.mutex.Lock()
		defer tb.mutex.Unlock()

		ctx := context.Background()
		// 使用Redis的DECR原子操作减少令牌
		err := redisClient.RedisClient.DecrBy(ctx, tb.key+":tokens", 1).Err()
		if err!= nil {
			logx.Error("redis decr err:", err)
			return false	
		}
		return true
	}
	return false
}

// Handler 实现中间件接口
func (tb *TokenBucket) Handler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !tb.Allow() {
			logx.Error("rate limit exceeded")

			response.HttpResponse(r, w, nil, errs.NewErrCodeInfo(110, "rate limit exceeded"))
			return
		}
		next(w, r)
	}
}
