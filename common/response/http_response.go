package response

import (
	"google.golang.org/grpc/status"
	"grpc_demo_client/common/errs"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// HttpResponse 通用接口返回 成功失败均200
func HttpResponse(r *http.Request, w http.ResponseWriter, resp interface{}, err error) {
	RawHttpResult(r, w, resp, err, http.StatusOK)
}

// AuthHttpResult 认证接口返回 失败返回401
func AuthHttpResult(r *http.Request, w http.ResponseWriter, resp interface{}, err error) {
	RawHttpResult(r, w, resp, err, http.StatusUnauthorized)
}

// GrantHttpRequest 授权接口返回 失败返回403
func GrantHttpRequest(r *http.Request, w http.ResponseWriter, resp interface{}, err error) {
	RawHttpResult(r, w, resp, err, http.StatusForbidden)
}

// RawHttpResult 基础接口返回
func RawHttpResult(r *http.Request, w http.ResponseWriter, resp interface{}, err error, defaultErrorStatusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err == nil {
		// 成功
		r := Success(resp)
		httpx.WriteJson(w, http.StatusOK, r)
	} else {
		logx.WithContext(r.Context()).Errorf("ERR: %+v ", err)
		// 如果是自定义的CodeError
		errCode := errs.UNKNOWN_ERR
		errMsg := "未知错误，请重试"
		if errs.IsCodeError(err) {
			errx := err.(*errs.CodeError)
			errCode = errx.GetErrCode()
			if errCode < 17 {
				// grpc内部错误，不应该将msg传到前端
			} else {
				errMsg = errx.GetErrInfo()
			}
		} else if s, ok := status.FromError(err); ok {
			// rpc返回的错误会被包装成status.Status
			if s.Code() < 17 {
				// grpc内部错误，不应该将msg传到前端
			} else {
				errCode = uint32(s.Code())
				errMsg = s.Message()
			}
		}
		httpx.WriteJson(w, defaultErrorStatusCode, Error(errCode, errMsg))
	}
}

// ParamErrorResult http 参数错误返回
func ParamErrorResult(r *http.Request, w http.ResponseWriter, err error) {
	httpx.WriteJson(w, http.StatusOK, Error(errs.PARAM_ERR, " 参数有误: "+err.Error()))
}
