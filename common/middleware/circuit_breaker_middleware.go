package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// getServiceName 根据请求路径获取服务名称
func getServiceName(path string) string {
	// 从路径中提取服务名称，这里使用简单的路径前缀作为服务名称
	// 例如：/user/info -> user_service
	if len(path) > 1 {
		// 去除开头的斜杠，获取第一段路径
		parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
		if len(parts) > 0 {
			return parts[0] + "_service"
		}
	}
	return "default_service"
}

// CircuitBreakerMiddleware 熔断器中间件
func CircuitBreakerMiddleware(cbs map[string]*CircuitBreaker) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 根据请求路径获取服务名称
			serviceName := getServiceName(r.URL.Path)

			// 获取对应的熔断器
			cb, ok := cbs[serviceName]
			if !ok {
				// 如果没有找到对应的熔断器，使用默认熔断器
				cb = cbs["default_service"]
				if cb == nil {
					// 如果默认熔断器也不存在，直接放行请求
					next(w, r)
					return
				}
			}

			// 检查熔断器状态
			if !cb.allowRequest() {
				httpx.Error(w, ErrCircuitBreakerOpen)
				return
			}

			// 包装ResponseWriter以捕获状态码和响应内容
			wrapper := &responseWriterWrapper{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				body:           make([]byte, 0),
			}

			// 执行后续处理
			next(wrapper, r)

			// 检查响应内容中的错误码
			success := true
			if wrapper.statusCode != http.StatusOK {
				success = false
			} else if len(wrapper.body) > 0 {
				// 检查响应体中的错误码
				responseData := &struct {
					Status uint32 `json:"status"`
				}{}
				if err := json.Unmarshal(wrapper.body, responseData); err == nil {
					// 如果错误码不是200（成功），则认为请求失败
					if responseData.Status != 200 {
						success = false
					}
				}
			}
			cb.recordResult(success)
		}
	}
}

// responseWriterWrapper 包装http.ResponseWriter以捕获状态码和响应内容
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriterWrapper) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return w.ResponseWriter.Write(b)
}