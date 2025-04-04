package svc

import (
	"grpc_demo_client/internal/config"
	"grpc_demo_client/user/user"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	UserRPC user.UserClient
	Config  config.Config
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化 user_grpc 客户端
	UserRPC := user.NewUserClient(zrpc.MustNewClient(c.UserRPC).Conn())
	return &ServiceContext{
		UserRPC: UserRPC,
		Config: c,
	}
}
