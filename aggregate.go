package errors

import (
	"errors"
	"fmt"
)

// MessageCountMap contains occurrence for each error message.
// MessageCountMap 包含每个错误消息的出现次数。
type MessageCountMap map[string]int

// Aggregate represents an object that contains multiple errors, but does not
// necessarily have singular semantic meaning.
// The aggregate can be used with `errors.Is()` to check for the occurrence of
// a specific error type.
// Errors.As() is not supported, because the caller presumably cares about a
// specific error of potentially multiple that match the given type.
// Aggregate 表示一个包含多个错误但不一定是单一语义的对象。
// Aggregate 可以与 `errors.Is()` 一起使用，以检查特定错误类型的出现。
// Errors.As() 不支持，因为调用者可能关心的是与给定类型匹配的潜在多个错误中的特定错误。
type Aggregate interface {
	error
	Errors() []error
	Is(error) bool
}

// NewAggregate converts a slice of errors into an Aggregate interface, which
// is itself an implementation of the error interface.  If the slice is empty,
// this returns nil.
// It will check if any of the element of input error list is nil, to avoid
// nil pointer panic when call Error().
// NewAggregate 将一个错误列表转换为 Aggregate 接口，
// 本身是一个错误接口的实现。如果列表为空，则返回 nil。
// 它会检查输入错误列表中的任何元素是否为 nil，以避免在调用 Error() 时出现空指针 panic。
func NewAggregate(errlist []error) Aggregate {
	if len(errlist) == 0 {
		return nil
	}
	// In case of input error list contains nil
	var errs []error
	for _, e := range errlist {
		if e != nil {
			errs = append(errs, e)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return aggregate(errs)
}

// This helper implements the error and Errors interfaces.  Keeping it private
// prevents people from making an aggregate of 0 errors, which is not
// an error, but does satisfy the error interface.
// 这个辅助函数实现了错误和 Errors 接口。保持它私有，防止人们创建一个空的聚合，这不是一个错误，但确实满足错误接口。
type aggregate []error

// Error is part of the error interface.
// Error 是错误接口的一部分。
func (agg aggregate) Error() string {
	if len(agg) == 0 {
		// This should never happen, really.
		return ""
	}
	if len(agg) == 1 {
		return agg[0].Error()
	}
	seenerrs := NewString()
	result := ""
	agg.visit(func(err error) bool {
		msg := err.Error()
		if seenerrs.Has(msg) {
			return false
		}
		seenerrs.Insert(msg)
		if len(seenerrs) > 1 {
			result += ", "
		}
		result += msg
		return false
	})
	if len(seenerrs) == 1 {
		return result
	}
	return "[" + result + "]"
}

// Is 是 Aggregate 接口的一部分。
// Is 检查错误是否与给定的错误匹配。
func (agg aggregate) Is(target error) bool {
	return agg.visit(func(err error) bool {
		return errors.Is(err, target)
	})
}

// visit 是 aggregate 的一部分。
// visit 遍历错误列表，并调用给定的函数。
func (agg aggregate) visit(f func(err error) bool) bool {
	for _, err := range agg {
		switch err := err.(type) {
		case aggregate:
			if match := err.visit(f); match {
				return match
			}
		case Aggregate:
			for _, nestedErr := range err.Errors() {
				if match := f(nestedErr); match {
					return match
				}
			}
		default:
			if match := f(err); match {
				return match
			}
		}
	}

	return false
}

// Errors is part of the Aggregate interface.
// Errors 是 Aggregate 接口的一部分。
// Errors 返回错误列表。
func (agg aggregate) Errors() []error {
	return []error(agg)
}

// Matcher is used to match errors.  Returns true if the error matches.
// Matcher 用于匹配错误。如果错误匹配，则返回 true。
type Matcher func(error) bool

// FilterOut removes all errors that match any of the matchers from the input
// error.  If the input is a singular error, only that error is tested.  If the
// input implements the Aggregate interface, the list of errors will be
// processed recursively.
//
// This can be used, for example, to remove known-OK errors (such as io.EOF or
// os.PathNotFound) from a list of errors.
// FilterOut 删除所有匹配给定匹配器的错误。如果输入是一个单一的错误，则只测试该错误。如果输入实现 Aggregate 接口，则错误列表将递归处理。
// 这可以用于例如从错误列表中删除已知的 OK 错误（例如 io.EOF 或 os.PathNotFound）。
func FilterOut(err error, fns ...Matcher) error {
	if err == nil {
		return nil
	}
	if agg, ok := err.(Aggregate); ok {
		return NewAggregate(filterErrors(agg.Errors(), fns...))
	}
	if !matchesError(err, fns...) {
		return err
	}
	return nil
}

// matchesError returns true if any Matcher returns true
// matchesError 返回 true 如果任何 Matcher 返回 true
func matchesError(err error, fns ...Matcher) bool {
	for _, fn := range fns {
		if fn(err) {
			return true
		}
	}
	return false
}

// filterErrors returns any errors (or nested errors, if the list contains
// nested Errors) for which all fns return false. If no errors
// remain a nil list is returned. The resulting silec will have all
// nested slices flattened as a side effect.
// filterErrors 返回任何错误（或嵌套错误，如果列表包含嵌套 Errors），
// 对于所有 fns 返回 false。如果没有错误，则返回 nil。结果切片将具有所有嵌套切片展平作为副作用。
func filterErrors(list []error, fns ...Matcher) []error {
	result := []error{}
	for _, err := range list {
		r := FilterOut(err, fns...)
		if r != nil {
			result = append(result, r)
		}
	}
	return result
}

// Flatten takes an Aggregate, which may hold other Aggregates in arbitrary
// nesting, and flattens them all into a single Aggregate, recursively.
// Flatten 接受一个 Aggregate，可能包含其他 Aggregate 在任意嵌套中，并将其全部展平为一个 Aggregate，递归地。
func Flatten(agg Aggregate) Aggregate {
	result := []error{}
	if agg == nil {
		return nil
	}
	for _, err := range agg.Errors() {
		if a, ok := err.(Aggregate); ok {
			r := Flatten(a)
			if r != nil {
				result = append(result, r.Errors()...)
			}
		} else {
			if err != nil {
				result = append(result, err)
			}
		}
	}
	return NewAggregate(result)
}

// CreateAggregateFromMessageCountMap converts MessageCountMap Aggregate
// CreateAggregateFromMessageCountMap 将 MessageCountMap 转换为 Aggregate。
func CreateAggregateFromMessageCountMap(m MessageCountMap) Aggregate {
	if m == nil {
		return nil
	}
	result := make([]error, 0, len(m))
	for errStr, count := range m {
		var countStr string
		if count > 1 {
			countStr = fmt.Sprintf(" (repeated %v times)", count)
		}
		result = append(result, fmt.Errorf("%v%v", errStr, countStr))
	}
	return NewAggregate(result)
}

// Reduce will return err or, if err is an Aggregate and only has one item,
// the first item in the aggregate.
// Reduce 返回 err 或，如果 err 是一个 Aggregate 并且只有一个项目，则 Aggregate 中的第一个项目。
func Reduce(err error) error {
	if agg, ok := err.(Aggregate); ok && err != nil {
		switch len(agg.Errors()) {
		case 1:
			return agg.Errors()[0]
		case 0:
			return nil
		}
	}
	return err
}

// AggregateGoroutines runs the provided functions in parallel, stuffing all
// non-nil errors into the returned Aggregate.
// Returns nil if all the functions complete successfully.
// AggregateGoroutines 并行运行提供的函数，将所有非 nil 错误放入返回的 Aggregate 中。
// 如果所有函数都成功完成，则返回 nil。
func AggregateGoroutines(funcs ...func() error) Aggregate {
	errChan := make(chan error, len(funcs))
	for _, f := range funcs {
		go func(f func() error) { errChan <- f() }(f)
	}
	errs := make([]error, 0)
	for i := 0; i < cap(errChan); i++ {
		if err := <-errChan; err != nil {
			errs = append(errs, err)
		}
	}
	return NewAggregate(errs)
}

// ErrPreconditionViolated is returned when the precondition is violated
// ErrPreconditionViolated 在预条件违反时返回
var ErrPreconditionViolated = errors.New("precondition is violated")
