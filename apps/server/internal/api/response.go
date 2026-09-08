package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/lindaailabs/yuyan/server/internal/pkg/errcode"
)

// Response REST 统一响应包裹：{"code":0,"msg":"ok","data":...}（LLM_DEV_GUIDE.md §5.3）。
type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

// OK 成功响应。
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{Code: int(errcode.OK), Msg: "ok", Data: data})
}

// FailErr 失败响应：errcode.Error 按段位映射 HTTP 状态，未知错误归入 5xxx。
func FailErr(c *gin.Context, err error) {
	var ec *errcode.Error
	if !errors.As(err, &ec) {
		ec = errcode.New(errcode.ErrInternal, "internal error")
	}
	c.JSON(httpStatusOf(ec.Code), Response{Code: int(ec.Code), Msg: ec.Msg, Data: nil})
}

func httpStatusOf(code errcode.Code) int {
	switch {
	case code == errcode.ErrUnauthorized:
		return http.StatusUnauthorized // 1002 → 401，客户端拦截器据此触发 refresh
	case errcode.SegmentOf(code) == 1000:
		return http.StatusBadRequest // 其余 1xxx 参数错误 → 400
	case errcode.SegmentOf(code) == 2000:
		return http.StatusOK // 2xxx 业务错误：HTTP 层成功、业务码表达失败
	default:
		return http.StatusInternalServerError
	}
}
