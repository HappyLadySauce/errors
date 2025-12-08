package errors

import (
	"fmt"
	"net/http"
	"sync"
)

var (
	// unknownCoder is the default error code for unknown errors.
	// unknownCoder 是未知错误的默认错误代码。
	unknownCoder defaultCoder = defaultCoder{1, http.StatusInternalServerError, "An internal server error occurred", "http://github.com/HappyLadySauce/errors/README.md"}
)

// Coder defines an interface for an error code detail information.
// Coder 定义了一个错误代码详细信息的接口。
type Coder interface {
	// HTTP status that should be used for the associated error code.
	// HTTP 状态码，用于关联错误代码。
	HTTPStatus() int

	// External (user) facing error text
	// 外部（用户）面向的错误文本。
	String() string

	// Reference returns the detail documents for user.
	// Reference 返回用户详细文档。
	Reference() string

	// Code returns the code of the coder
	// Code 返回错误代码的代码。
	Code() int
}

// defaultCoder implements the Coder interface.
// defaultCoder 是默认的错误代码详细信息。
type defaultCoder struct {
	// C refers to the integer code of the ErrCode.
	// C 指的是 ErrCode 的整数代码。
	C int

	// HTTP status that should be used for the associated error code.
	// HTTP 状态码，用于关联错误代码。
	HTTP int

	// External (user) facing error text
	// 外部（用户）面向的错误文本。
	Ext string

	// Ref specify the reference document
	// Ref 指定参考文档。
	Ref string
}

// Code returns the integer code of the coder.
// Code 返回错误代码的整数代码。
func (coder defaultCoder) Code() int {
	return coder.C

}

// String implements stringer. String returns the external error message,
// if any.
// String 实现 stringer 接口。返回外部错误消息，如果有的话。
func (coder defaultCoder) String() string {
	return coder.Ext
}

// HTTPStatus returns the associated HTTP status code, if any. Otherwise,
// returns 200.
// HTTPStatus 返回关联的 HTTP 状态码，如果有的话。否则，返回 200。
func (coder defaultCoder) HTTPStatus() int {
	if coder.HTTP == 0 {
		return 500
	}

	return coder.HTTP
}

// Reference returns the reference document.
// Reference 返回参考文档。
func (coder defaultCoder) Reference() string {
	return coder.Ref
}

// codes contains a map of error codes to metadata.
// codes 包含一个错误代码到元数据的映射。
var codes = map[int]Coder{}
var codeMux = &sync.Mutex{}

// Register register a user define error code.
// It will overrid the exist code.
// Register 注册一个用户定义的错误代码。
// 如果代码已经存在，将覆盖现有的代码。
func Register(coder Coder) {
	if coder.Code() == 0 {
		panic("code `0` is reserved by `github.com/HappyLadySauce/errors` as unknownCode error code")
	}

	codeMux.Lock()
	defer codeMux.Unlock()

	codes[coder.Code()] = coder
}

// MustRegister register a user define error code.
// It will panic when the same Code already exist.
// MustRegister 注册一个用户定义的错误代码。
// 如果代码已经存在，将panic。
func MustRegister(coder Coder) {
	if coder.Code() == 0 {
		panic("code '0' is reserved by 'github.com/HappyLadySauce/errors' as ErrUnknown error code")
	}

	codeMux.Lock()
	defer codeMux.Unlock()

	if _, ok := codes[coder.Code()]; ok {
		panic(fmt.Sprintf("code: %d already exist", coder.Code()))
	}

	codes[coder.Code()] = coder
}

// ParseCoder parse any error into *withCode.
// nil error will return nil direct.
// None withStack error will be parsed as ErrUnknown.
// ParseCoder 解析任何错误到 *withCode。
// 如果错误为 nil，则直接返回 nil。
// 如果没有 withStack 错误，则解析为 ErrUnknown。
func ParseCoder(err error) Coder {
	if err == nil {
		return nil
	}

	if v, ok := err.(*withCode); ok {
		if coder, ok := codes[v.code]; ok {
			return coder
		}
	}

	return unknownCoder
}

// IsCode reports whether any error in err's chain contains the given error code.
// IsCode 报告 err 的链中是否包含给定的错误代码。
func IsCode(err error, code int) bool {
	if v, ok := err.(*withCode); ok {
		if v.code == code {
			return true
		}

		if v.cause != nil {
			return IsCode(v.cause, code)
		}

		return false
	}

	return false
}

// init 初始化 unknownCoder。
func init() {
	codes[unknownCoder.Code()] = unknownCoder
}
