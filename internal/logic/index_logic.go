package logic

import (
	"context"

	"grpc_demo_client/internal/svc"
	"grpc_demo_client/internal/types"
	"grpc_demo_client/user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type IndexLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewIndexLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IndexLogic {
	return &IndexLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *IndexLogic) Index(req *types.Empty) (resp *types.IndexResp, err error) {
	var userGrpcResp *user.UserInfoResp

	// 直接调用RPC服务，熔断保护已经在中间件层实现
	userGrpcResp, err = l.svcCtx.UserRPC.UserInfo(l.ctx, &user.UserInfoReq{Name: "第一次测试链接"})
	if err != nil {
		logx.Errorf("UserInfo err: %v", err)
		return
	}

	resp = &types.IndexResp{
		Ping: userGrpcResp.Name,
	}
	return
}
