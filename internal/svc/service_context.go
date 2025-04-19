package svc

import (
	"context"
	"encoding/json"
	"grpc_demo_client/common/middleware"
	redisClient "grpc_demo_client/common/redis"
	"grpc_demo_client/internal/config"
	"grpc_demo_client/user/user"
	"os"
	"strconv"

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
	// 加载 redis 环境变量
	redisHost := os.Getenv("REDIS_HOST")
	c.RedisConfig.Host = redisHost
	c.RedisConfig.Pass = os.Getenv("REDIS_PASS")
	redisDbStr := os.Getenv("REDIS_DB")
	if redisDbStr != "" {
		redisDb, err := strconv.Atoi(redisDbStr)
		if err != nil {
			logx.Errorf("redisDbStr:%v, err:%v", redisDbStr, err)
			panic(err)
		}
		c.RedisConfig.Db = int8(redisDb)
	}
	redisClient.RedisClient = redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: c.RedisConfig.Pass,
		DB:       int(c.RedisConfig.Db),
	})
	err := redisClient.RedisClient.Ping(context.Background()).Err()
	if err != nil {
		logx.Errorf("redisClient.RedisClient.Ping err:%v", err)
		panic(err)
	}
	// 读取 etc 配置
	ehost := os.Getenv("ETCD_HOST")
	if ehost != "" {
		ehosts := []string{}
		err := json.Unmarshal([]byte(ehost), &ehosts)
		if err != nil {
			logx.Error("etcd hosts unmarshal failed", err)
		} else {

			c.UserRPC.Etcd.Hosts = ehosts
		}
	}
	eKey := os.Getenv("ETCD_KEY")
	if eKey != "" {
		c.UserRPC.Etcd.Key = eKey
	}
	eUser := os.Getenv("ETCD_USER")
	if eUser != "" {
		c.UserRPC.Etcd.User = eUser
	}
	ePass := os.Getenv("ETCD_PASS")
	if ePass != "" {
		c.UserRPC.Etcd.Pass = ePass
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
