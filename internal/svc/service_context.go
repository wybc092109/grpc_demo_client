package svc

import (
	"context"
	"grpc_demo_client/common/middleware"
	redisClient "grpc_demo_client/common/redis"
	"grpc_demo_client/internal/config"
	"grpc_demo_client/user/user"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	UserRPC         user.UserClient
	Config          config.Config
	CircuitBreakers map[string]*middleware.CircuitBreaker
}

func NewServiceContext(c config.Config) *ServiceContext {
	redisClient.RedisClient = redis.NewClient(&redis.Options{
		Addr:     c.RedisConfig.Host,
		Password: c.RedisConfig.Pass,
		DB:       int(c.RedisConfig.Db),
	})
	err := redisClient.RedisClient.Ping(context.Background()).Err()
	if err != nil {
		logx.Errorf("redisClient.RedisClient.Ping err:%v", err)
		panic(err)
	}
	// 初始化 user_grpc 客户端
	UserRPC := user.NewUserClient(zrpc.MustNewClient(c.UserRPC).Conn())

	// 初始化熔断器映射
	circuitBreakers := make(map[string]*middleware.CircuitBreaker)

	// 为每个服务创建独立的熔断器实例
	circuitBreakers["index_service"] = middleware.NewCircuitBreaker("user_service", redisClient.RedisClient,
		middleware.WithFailureThreshold(1),
		middleware.WithSuccessThreshold(2),
		middleware.WithHalfOpenTimeout(300),
	)

	// 可以在这里添加更多服务的熔断器
	// circuitBreakers["other_service"] = middleware.NewCircuitBreaker("other_service", redisClient.RedisClient, ...)

	return &ServiceContext{
		UserRPC:         UserRPC,
		Config:          c,
		CircuitBreakers: circuitBreakers,
	}
}
