// Package errcode 定义统一错误码与带码错误类型。
// 分段约定（LLM_DEV_GUIDE.md §5.3）：0 成功；1xxx 参数/鉴权；2xxx 业务；5xxx 服务端。
package errcode

import "fmt"

// Code 错误码。
type Code int

// 预定义错误码（W1 基础集，业务码随功能逐步补充）。
const (
	OK              Code = 0
	ErrInvalidParam Code = 1001 // 参数校验失败
	ErrUnauthorized Code = 1002 // 未登录/凭证无效
	ErrInternal     Code = 5001 // 服务端内部错误
)

// 错误码分段。
const (
	segOK       = 0
	segParam    = 1000 // 1xxx 参数/鉴权
	segBusiness = 2000 // 2xxx 业务
	segServer   = 5000 // 5xxx 服务端
)

// Error 带 code 的业务错误；service 层返回本类型，api 层统一转换为响应包裹。
type Error struct {
	Code Code
	Msg  string
}

func (e *Error) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Msg)
}

// New 构造带码错误。code 必须落在合法段位（0/1xxx/2xxx/5xxx）内，
// 否则 panic——段位错误属于编程缺陷，应在编码期暴露而非运行期兜底。
func New(code Code, msg string) *Error {
	if !validSegment(code) {
		panic(fmt.Sprintf("errcode: invalid code segment: %d (allowed: 0, 1xxx, 2xxx, 5xxx)", code))
	}
	return &Error{Code: code, Msg: msg}
}

func validSegment(c Code) bool {
	switch {
	case c == segOK:
		return true
	case c >= 1000 && c <= 1999:
		return true
	case c >= 2000 && c <= 2999:
		return true
	case c >= 5000 && c <= 5999:
		return true
	default:
		return false
	}
}

// SegmentOf 返回错误码所属段位基数（0/1000/2000/5000）；非法段位返回 -1。
func SegmentOf(c Code) int {
	switch {
	case c == segOK:
		return segOK
	case c >= 1000 && c <= 1999:
		return segParam
	case c >= 2000 && c <= 2999:
		return segBusiness
	case c >= 5000 && c <= 5999:
		return segServer
	default:
		return -1
	}
}
