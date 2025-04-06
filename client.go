package main

import (
	"flag"
	"fmt"

	"grpc_demo_client/common/middleware"
	"grpc_demo_client/internal/config"
	"grpc_demo_client/internal/handler"
	"grpc_demo_client/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/client.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)
	// 初始化限流中间件
	rateLimit := middleware.NewTokenBucket(c.RateLimit.Rate, c.RateLimit.Capacity)
	server.Use(rateLimit.Handler)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
