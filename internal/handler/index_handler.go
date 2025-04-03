package handler

import (
	"net/http"

	"grpc_demo_client/common/response"
	"grpc_demo_client/internal/logic"
	"grpc_demo_client/internal/svc"
	"grpc_demo_client/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func indexHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.Empty
		if err := httpx.Parse(r, &req); err != nil {
			//参数错误处理
			response.ParamErrorResult(r, w, err)
			return
		}

		l := logic.NewIndexLogic(r.Context(), svcCtx)
		resp, err := l.Index(&req)
		response.HttpResponse(r, w, resp, err)
	}
}
