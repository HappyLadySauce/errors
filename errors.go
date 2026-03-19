package errors

import (
	stderrors "errors"
	"fmt"
	"io"
)

// New returns an error with the supplied message.
// New also records the stack trace at the point it was called.
// New 返回一个带有指定消息的错误。
// New 还会记录调用点的堆栈跟踪。
func New(message string) error {
	return &fundamental{
		msg: message,
		stack: callers(),
	}
}

// Errorf formats according to a format specifier and returns the string
// as a value that satisfies error.
// Errorf also records the stack trace at the point it was called.
// Errorf 根据格式说明符格式化并返回一个满足 error 的值。
// Errorf 还会记录调用点的堆栈跟踪。
func Errorf(format string, args ...interface{}) error {
	return &fundamental{
		msg: fmt.Sprintf(format, args...),
		stack: callers(),
	}
}

// fundamental is an error that has a message and a stack, but no caller.
// fundamental 是一个有消息和堆栈但没有调用者的错误。
type fundamental struct {
	msg string
	*stack
}

// Error returns the message of the error.
// Error 返回错误的消息。
func (f *fundamental) Error() string { return f.msg }

// Format formats the error according to the fmt.Formatter interface.
//
//    %s    message
//    %v    message and stack trace
//    %+v   message, stack trace, and caller
//    %q    quoted message
// Format 接受改变某些动词打印的标志，如下所示：
//
//    %s    消息
//    %v    消息和堆栈跟踪
//    %+v   消息、堆栈跟踪和调用者
//    %q    引用的消息
func (f *fundamental) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			io.WriteString(s, f.msg)
			f.stack.Format(s, verb)
			return
		}
		fallthrough
	case 's':
		io.WriteString(s, f.msg)
	case 'q':
		fmt.Fprintf(s, "%q", f.msg)
	}
}

// WithStack annotates err with a stack trace at the point WithStack was called.
// If err is nil, WithStack returns nil.
// WithStack 在 WithStack 被调用的点注释 err 与堆栈跟踪。
// 如果 err 为 nil，WithStack 返回 nil。
func WithStack(err error) error {
	if err == nil {
		return nil
	}

	if e, ok := err.(*withCode); ok {
		return &withCode{
			err:   e.err,
			code:  e.code,
			cause: err,
			stack: callers(),
		}
	}

	return &withStack{
		err,
		callers(),
	}
}

// withStack is an error that has a message, a stack, and a cause.
// withStack 是一个有消息、堆栈和根本原因的错误。
type withStack struct {
	error
	*stack
}

// Cause returns the original error.
// Cause 返回原始错误。
func (w *withStack) Cause() error { return w.error }

// Unwrap provides compatibility for Go 1.13 error chains.
// Unwrap 提供与 Go 1.13 错误链的兼容性。
func (w *withStack) Unwrap() error {
	if e, ok := w.error.(interface{ Unwrap() error }); ok {
		return e.Unwrap()
	}

	return w.error
}

// Format formats the error according to the fmt.Formatter interface.
// Format 根据 fmt.Formatter 接口格式化错误。
//
//    %s    message
//    %v    message and stack trace
//    %+v   message, stack trace, and caller
//    %q    quoted message
// Format 接受改变某些动词打印的标志，如下所示：
//
//    %s    消息
//    %v    消息和堆栈跟踪
//    %+v   消息、堆栈跟踪和调用者
//    %q    引用的消息
func (w *withStack) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v", w.Cause())
			w.stack.Format(s, verb)
			return
		}
		fallthrough
	case 's':
		io.WriteString(s, w.Error())
	case 'q':
		fmt.Fprintf(s, "%q", w.Error())
	}
}

// Wrap returns an error annotating err with a stack trace
// at the point Wrap is called, and the supplied message.
// If err is nil, Wrap returns nil.
// Wrap 返回一个错误，注释 err 与堆栈跟踪
// 在 Wrap 被调用的点，以及提供的消息。
// 如果 err 为 nil，Wrap 返回 nil。
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*withCode); ok {
		return &withCode{
			err:   fmt.Errorf(message),
			code:  e.code,
			cause: err,
			stack: callers(),
		}
	}

	err = &withMessage{
		cause: err,
		msg:   message,
	}
	return &withStack{
		err,
		callers(),
	}
}

// Wrapf returns an error annotating err with a stack trace
// at the point Wrapf is called, and the format specifier.
// If err is nil, Wrapf returns nil.
// Wrapf 返回一个错误，注释 err 与堆栈跟踪
// 在 Wrapf 被调用的点，以及格式说明符。
// 如果 err 为 nil，Wrapf 返回 nil。
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	if e, ok := err.(*withCode); ok {
		return &withCode{
			err:   fmt.Errorf(format, args...),
			code:  e.code,
			cause: err,
			stack: callers(),
		}
	}

	err = &withMessage{
		cause: err,
		msg:   fmt.Sprintf(format, args...),
	}
	return &withStack{
		err,
		callers(),
	}
}

// WithMessage annotates err with a new message.
// If err is nil, WithMessage returns nil.
// WithMessage 注释 err 与新消息。
// 如果 err 为 nil，WithMessage 返回 nil。
func WithMessage(err error, message string) error {
	if err == nil {
		return nil
	}
	return &withMessage{
		cause: err,
		msg:   message,
	}
}

// WithMessagef annotates err with the format specifier.
// If err is nil, WithMessagef returns nil.
// WithMessagef 注释 err 与格式说明符。
// 如果 err 为 nil，WithMessagef 返回 nil。
func WithMessagef(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return &withMessage{
		cause: err,
		msg:   fmt.Sprintf(format, args...),
	}
}

type withMessage struct {
	cause error
	msg   string
}

func (w *withMessage) Error() string { return w.msg }
func (w *withMessage) Cause() error  { return w.cause }

// Unwrap provides compatibility for Go 1.13 error chains.
func (w *withMessage) Unwrap() error { return w.cause }

func (w *withMessage) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v\n", w.Cause())
			io.WriteString(s, w.msg)
			return
		}
		fallthrough
	case 's', 'q':
		io.WriteString(s, w.Error())
	}
}

type withCode struct {
	err   error
	code  int
	cause error
	*stack
}

// WithCode returns an error with the supplied code and format specifier.
// WithCode also records the stack trace at the point it was called.
// WithCode 返回一个带有指定代码和格式说明符的错误。
// WithCode 还会记录调用点的堆栈跟踪。
func WithCode(code int, format string, args ...interface{}) error {
	return &withCode{
		err:   fmt.Errorf(format, args...),
		code:  code,
		stack: callers(),
	}
}

// WrapC returns an error with the supplied code and format specifier.
// WrapC also records the stack trace at the point it was called.
// WrapC 返回一个带有指定代码和格式说明符的错误。
// WrapC 还会记录调用点的堆栈跟踪。
func WrapC(err error, code int, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	return &withCode{
		err:   fmt.Errorf(format, args...),
		code:  code,
		cause: err,
		stack: callers(),
	}
}

// Error return the externally-safe error message.
// Error 返回外部安全的错误消息。
func (w *withCode) Error() string { return fmt.Sprintf("%v", w) }

// Cause return the cause of the withCode error.
// Cause 返回 withCode 错误的根本原因。
func (w *withCode) Cause() error { return w.cause }

// Unwrap provides compatibility for Go 1.13 error chains.
// Unwrap 提供与 Go 1.13 错误链的兼容性。
func (w *withCode) Unwrap() error { return w.cause }

// Cause returns the underlying cause of the error, if possible.
// An error value has a cause if it implements the following
// interface:
//
//     type causer interface {
//            Cause() error
//     }
//
// If the error does not implement Cause, the original error will
// be returned. If the error is nil, nil will be returned without further
// investigation.
// Cause 返回错误的根本原因，如果可能。
// 如果错误实现了以下接口，则错误有根本原因：
//
//     type causer interface {
//            Cause() error
//     }
//
// 如果错误不实现 Cause，则返回原始错误。如果错误为 nil，则返回 nil 而不进行进一步调查。
func Cause(err error) error {
	type causer interface {
		Cause() error
	}

	for err != nil {
		cause, ok := err.(causer)
		if !ok {
			break
		}

		if cause.Cause() == nil {
			break
		}

		err = cause.Cause()
	}
	return err
}

// Join returns an error that wraps the given errors.
// Any nil error values are discarded.
// Join returns nil if every input is nil.
// The error formats as the concatenation of the strings obtained
// by calling the Error method of each element of errs, with a newline
// between each string.
//
// A returned error wraps all non-nil errors and supports
// errors.Is/As/Unwrap for each.
//
// Join 返回一个包装了给定错误的错误。
// 任何 nil 错误值都会被丢弃。
// 如果所有输入都是 nil，Join 返回 nil。
// 错误格式化为调用每个 errs 元素的 Error 方法获得的字符串的连接，每个字符串之间有一个换行符。
//
// 返回的错误包装所有非 nil 错误，并支持对每个错误使用 errors.Is/As/Unwrap。
// 这是一个对 Go 1.20+ 标准库 errors.Join 的包装。
func Join(errs ...error) error {
	return stderrors.Join(errs...)
}
