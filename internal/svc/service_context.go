package svc

import (
	"context"
	redisClient "grpc_demo_client/common/redis"
	"grpc_demo_client/internal/config"
	"grpc_demo_client/user/user"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	UserRPC user.UserClient
	Config  config.Config
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
	return &ServiceContext{
		UserRPC: UserRPC,
		Config:  c,
	}
}
