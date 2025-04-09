package middleware

import "errors"

var (
	// ErrCircuitBreakerOpen 熔断器开启错误
	ErrCircuitBreakerOpen = errors.New("circuit breaker is open")
)
