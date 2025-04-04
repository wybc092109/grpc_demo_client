package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	UserRPC zrpc.RpcClientConf
	rest.RestConf
}

var C *Config