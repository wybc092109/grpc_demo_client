package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	UserRPC zrpc.RpcClientConf
	rest.RestConf
	RateLimit   RateLimitConfig 
	RedisConfig struct {
		Host string 
		Pass string 
		Db   int8   
	} 
}

type RateLimitConfig struct {
	Rate     int64 // 令牌生成速率（每秒）
	Capacity int64 // 令牌桶容量
}

var C *Config
