package errors

import (
	"reflect"
	"sort"
)

// Empty is public since it is used by some internal API objects for conversions between external
// string arrays and internal sets, and conversion logic requires public types today.
// Empty 是公共的，因为它被一些内部 API 对象用于在外部字符串数组和内部集之间进行转换，并且转换逻辑需要公共类型。
type Empty struct{}

// String is a set of strings, implemented via map[string]struct{} for minimal memory consumption.
// String 是一个字符串的集合，通过 map[string]struct{} 实现最小内存消耗。
type String map[string]Empty

// NewString creates a String from a list of values.
func NewString(items ...string) String {
	ss := String{}
	ss.Insert(items...)
	return ss
}

// StringKeySet creates a String from a keys of a map[string](? extends interface{}).
// If the value passed in is not actually a map, this will panic.
// StringKeySet 从 map[string](? extends interface{}) 的键创建一个 String。
// 如果传入的值实际上不是一个 map，这将 panic。
func StringKeySet(theMap interface{}) String {
	v := reflect.ValueOf(theMap)
	ret := String{}

	for _, keyValue := range v.MapKeys() {
		ret.Insert(keyValue.Interface().(string))
	}
	return ret
}

// Insert adds items to the set.
// Insert 将项目添加到集合中。
func (s String) Insert(items ...string) String {
	for _, item := range items {
		s[item] = Empty{}
	}
	return s
}

// Delete removes all items from the set.
// Delete 从集合中删除所有项目。
func (s String) Delete(items ...string) String {
	for _, item := range items {
		delete(s, item)
	}
	return s
}

// Has returns true if and only if item is contained in the set.
// Has 如果且仅当项目包含在集合中时返回 true。
func (s String) Has(item string) bool {
	_, contained := s[item]
	return contained
}

// HasAll returns true if and only if all items are contained in the set.
// HasAll 如果且仅当所有项目包含在集合中时返回 true。
func (s String) HasAll(items ...string) bool {
	for _, item := range items {
		if !s.Has(item) {
			return false
		}
	}
	return true
}

// HasAny returns true if any items are contained in the set.
// HasAny 如果任何项目包含在集合中时返回 true。
func (s String) HasAny(items ...string) bool {
	for _, item := range items {
		if s.Has(item) {
			return true
		}
	}
	return false
}

// Difference returns a set of objects that are not in s2
// For example:
// s = {a1, a2, a3}
// s2 = {a1, a2, a4, a5}
// s.Difference(s2) = {a3}
// s2.Difference(s) = {a4, a5}
// Difference 返回一个集合，其中包含 s 中的项目，但不包含 s2 中的项目。
func (s String) Difference(s2 String) String {
	result := NewString()
	for key := range s {
		if !s2.Has(key) {
			result.Insert(key)
		}
	}
	return result
}

// Union returns a new set which includes items in either s or s2.
// For example:
// s = {a1, a2}
// s2 = {a3, a4}
// s.Union(s2) = {a1, a2, a3, a4}
// s2.Union(s) = {a1, a2, a3, a4}
// Union 返回一个新集合，其中包含 s 或 s2 中的项目。
func (s String) Union(s2 String) String {
	result := NewString()
	for key := range s {
		result.Insert(key)
	}
	for key := range s2 {
		result.Insert(key)
	}
	return result
}

// Intersection returns a new set which includes the item in BOTH s and s2
// For example:
// s = {a1, a2}
// s2 = {a2, a3}
// s.Intersection(s2) = {a2}
// Intersection 返回一个新集合，其中包含 s 和 s2 中的项目。
func (s String) Intersection(s2 String) String {
	var walk, other String
	result := NewString()
	if s.Len() < s2.Len() {
		walk = s
		other = s2
	} else {
		walk = s2
		other = s
	}
	for key := range walk {
		if other.Has(key) {
			result.Insert(key)
		}
	}
	return result
}

// IsSuperset returns true if and only if s is a superset of s2.
// IsSuperset 如果且仅当 s 是 s2 的超集时返回 true。
func (s String) IsSuperset(s2 String) bool {
	for item := range s2 {
		if !s.Has(item) {
			return false
		}
	}
	return true
}

// Equal returns true if and only if s is equal (as a set) to s2.
// Two sets are equal if their membership is identical.
// (In practice, this means same elements, order doesn't matter)
// Equal 如果且仅当 s 和 s2 相等（作为集合）时返回 true。
func (s String) Equal(s2 String) bool {
	return len(s) == len(s2) && s.IsSuperset(s2)
}

// sortableSliceOfString 是一个字符串的切片，用于排序。
type sortableSliceOfString []string

// Len 返回切片的长度。
func (s sortableSliceOfString) Len() int           { return len(s) }
// Less 返回切片中的第 i 个元素是否小于第 j 个元素。
func (s sortableSliceOfString) Less(i, j int) bool { return lessString(s[i], s[j]) }
func (s sortableSliceOfString) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// List returns the contents as a sorted string slice.
// List 返回一个排序后的字符串切片。
func (s String) List() []string {
	res := make(sortableSliceOfString, 0, len(s))
	for key := range s {
		res = append(res, key)
	}
	sort.Sort(res)
	return []string(res)
}

// UnsortedList returns the slice with contents in random order.
// UnsortedList 返回一个随机顺序的字符串切片。
func (s String) UnsortedList() []string {
	res := make([]string, 0, len(s))
	for key := range s {
		res = append(res, key)
	}
	return res
}

// PopAny returns a single element from the set.
// PopAny 从集合中返回一个随机元素。
func (s String) PopAny() (string, bool) {
	for key := range s {
		s.Delete(key)
		return key, true
	}
	var zeroValue string
	return zeroValue, false
}

// Len returns the size of the set.
// Len 返回集合的大小。
func (s String) Len() int {
	return len(s)
}

// lessString 返回 lhs 是否小于 rhs。
func lessString(lhs, rhs string) bool {
	return lhs < rhs
}
