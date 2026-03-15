package errors

import "fmt"

// CodeMsg is a struct that contains a code and a message.
// It implements the error interface and is compatible with github.com/zeromicro/x/errors.
// Use it with go-zero httpx: type assert to *errors.CodeMsg to get Code and Msg for HTTP response.
//
// CodeMsg 包含 code 和 message，与 zeromicro/x/errors 兼容。
// 可与 go-zero httpx 配合：通过 *errors.CodeMsg 类型断言获取 Code、Msg 用于 HTTP 响应。
type CodeMsg struct {
	Code int
	Msg  string
}

// Error implements the error interface.
// Format matches zeromicro/x/errors: "code: %d, msg: %s".
func (c *CodeMsg) Error() string {
	return fmt.Sprintf("code: %d, msg: %s", c.Code, c.Msg)
}

// Unwrap returns nil so that *CodeMsg can be wrapped by Wrap/WithStack/WrapC.
func (c *CodeMsg) Unwrap() error {
	return nil
}

// NewCodeMsg creates a new CodeMsg error.
// Use this for zeromicro/go-zero compatibility (e.g. with httpx.SetErrorHandler).
// For drop-in compatibility with code that expects errors.New(code, msg), use NewCodeMsg.
//
// NewCodeMsg 创建一个带 code 和 msg 的错误，用于与 zeromicro/go-zero 兼容。
func NewCodeMsg(code int, msg string) error {
	return &CodeMsg{Code: code, Msg: msg}
}

// ToCodeMsg extracts *CodeMsg from the error chain, or builds one from *withCode.
// Returns (nil, false) if err is nil or the chain contains no CodeMsg or withCode.
// Use this to convert existing withCode errors to CodeMsg for HTTP layer (e.g. go-zero httpx).
//
// ToCodeMsg 从错误链中取出 *CodeMsg，或从 *withCode 构造。用于在 HTTP 层统一返回 CodeMsg。
func ToCodeMsg(err error) (*CodeMsg, bool) {
	for err != nil {
		if c, ok := err.(*CodeMsg); ok {
			return c, true
		}
		if w, ok := err.(*withCode); ok {
			coder := ParseCoder(w)
			msg := coder.String()
			if msg == "" {
				msg = w.err.Error()
			}
			return &CodeMsg{Code: w.code, Msg: msg}, true
		}
		err = Unwrap(err)
	}
	return nil, false
}
