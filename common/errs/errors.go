package errs

import (
	"errors"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// 公共状态码
const (
	PARAM_ERR    uint32 = 100 //参数异常
	IP_ILLEGAL   uint32 = 102 //非法IP
	OVERTIME     uint32 = 103 //请求超时
	LIMITED_FLOW uint32 = 104 //请求限流
	UNKNOWN_ERR  uint32 = 110 //未知错误
	StatusOK     uint32 = 200 //成功
)

type GErr interface {
	GRPCStatus() *status.Status
}

func (e *CodeError) GRPCStatus() *status.Status {
	return e.code
}

type CodeError struct {
	base error
	code *status.Status
}

// GetErrCode 返回错误码
func (e *CodeError) GetErrCode() uint32 {
	return uint32(e.code.Code())
}

// GetErrInfo 返回错误信息
func (e *CodeError) GetErrInfo() string {
	return e.code.Message()
}

// NewErrCodeInfo 自定义错误码，自定义错误信息
func NewErrCodeInfo(errCode uint32, errInfo string) *CodeError {
	return &CodeError{
		base: errors.New(errInfo),
		code: status.New(codes.Code(errCode), errInfo),
	}
}

// NewErrCode 自定义错误码，通用错误信息
func NewErrCode(errCode uint32) *CodeError {
	return &CodeError{
		base: errors.New("未知错误,请重试"),
		code: status.New(codes.Code(errCode), "未知错误,请重试"),
	}
}

// NewInfo 通用错误码，自定义错误信息
func NewInfo(errInfo string) *CodeError {
	return &CodeError{
		base: errors.New(errInfo),
		code: status.New(codes.Code(UNKNOWN_ERR), errInfo),
	}
}

// NewErr 通用错误码,通用错误信息
func NewErr(err error) *CodeError {
	return &CodeError{
		base: err,
		code: status.New(codes.Code(UNKNOWN_ERR), "未知错误,请重试"),
	}
}

// IsCodeError 判断
func IsCodeError(err error) bool {
	_, yes := err.(*CodeError)
	return yes
}

// Error 格式化输出
func (e *CodeError) Error() string {
	return fmt.Sprintf("ErrCode:%d，ErrInfo:%s", e.GetErrCode(), e.GetErrInfo())
}
