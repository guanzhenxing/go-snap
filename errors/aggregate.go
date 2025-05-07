// Package errors 提供错误处理原语，包括错误聚合功能。
package errors

import (
	"errors"
	"fmt"
)

// MessageCountMap 包含每个错误消息的出现次数。
type MessageCountMap map[string]int

// Aggregate 表示一个包含多个错误的对象，但不一定具有单一的语义含义。
// 该聚合可与`errors.Is()`一起使用，以检查特定错误类型的存在。
// 不支持Errors.As()，因为调用者可能只关心多个匹配给定类型的错误中的特定一个。
type Aggregate interface {
	error
	Errors() []error
	Is(error) bool
}

// NewAggregate 将错误切片转换为Aggregate接口，
// 该接口本身是error接口的实现。如果切片为空，则返回nil。
// 它会检查输入错误列表的任何元素是否为nil，以避免调用Error()时出现空指针panic。
//
// 示例:
//
//	err1 := errors.New("错误 1")
//	err2 := errors.New("错误 2")
//	agg := errors.NewAggregate([]error{err1, err2})
//	fmt.Println(agg.Error()) // 输出: [错误 1, 错误 2]
func NewAggregate(errs []error) Aggregate {
	if len(errs) == 0 {
		return nil
	}
	// 以防输入错误列表包含nil
	var validErrs []error
	for _, err := range errs {
		if err != nil {
			validErrs = append(validErrs, err)
		}
	}
	if len(validErrs) == 0 {
		return nil
	}
	return aggregate(validErrs)
}

// 这个辅助类型实现了error和Errors接口。保持其私有性
// 可防止人们创建0个错误的聚合，这不是一个错误，
// 但确实满足error接口。
type aggregate []error

// Error 实现error接口的一部分。
func (agg aggregate) Error() string {
	if len(agg) == 0 {
		// 这种情况不应发生，因为我们已过滤零错误的聚合。
		return ""
	}

	if len(agg) == 1 {
		// 单个错误的聚合直接返回原始错误消息，不加方括号
		return agg[0].Error()
	}

	seenErrs := make(map[string]struct{})
	result := "["
	for i, err := range agg {
		msg := err.Error()
		if _, ok := seenErrs[msg]; ok {
			continue
		}
		seenErrs[msg] = struct{}{}

		if i > 0 {
			result += ", "
		}
		result += msg
	}
	result += "]"

	// 特殊情况：如果去重后只有一个错误消息，去掉方括号
	if len(seenErrs) == 1 {
		return result[1 : len(result)-1]
	}

	return result
}

// Errors 返回组成这个聚合的所有错误。
func (agg aggregate) Errors() []error {
	return []error(agg)
}

// Is 实现了errors.Is接口。如果target在错误聚合中的任何错误中找到，则返回true。
// 如果target本身是一个错误聚合，则必须包含完全相同的错误集（顺序无关）。
func (agg aggregate) Is(target error) bool {
	// 如果目标是另一个聚合错误，则检查它们是否包含相同的错误集
	targetAgg, ok := target.(Aggregate)
	if ok {
		targetErrs := targetAgg.Errors()
		if len(targetErrs) != len(agg) {
			return false
		}

		// 确保每个错误都存在于两个聚合中
		// 注意：这是一个O(n²)操作，对于大量错误可能需要优化
		matches := 0
		for _, err1 := range agg {
			for _, err2 := range targetErrs {
				if errors.Is(err1, err2) {
					matches++
					break
				}
			}
		}
		return matches == len(agg)
	}

	// 如果目标是单个错误，检查它是否存在于聚合中
	for _, err := range agg {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

// Flatten 接受一个可能以任意嵌套方式包含其他Aggregate的Aggregate，
// 并递归地将它们全部展平为单个Aggregate。
//
// 示例:
//
//	agg1 := errors.NewAggregate([]error{errors.New("错误 1"), errors.New("错误 2")})
//	agg2 := errors.NewAggregate([]error{errors.New("错误 3"), agg1})
//	flattened := errors.Flatten(agg2)
//	// flattened包含所有三个错误，没有嵌套
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

// Reduce 将返回err，或者如果err是一个只有一个项目的Aggregate，
// 则返回Aggregate中的第一个项目。
// 这对于处理可能只包含单个错误的多错误情况非常有用，
// 您希望将其作为单个错误处理。
//
// 示例:
//
//	err := errors.NewAggregate([]error{errors.New("单个错误")})
//	reduced := errors.Reduce(err) // 返回单个错误
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

// NewAggregateFromCountMap 将MessageCountMap转换为Aggregate。
// 这对于从计数器创建错误消息很有用。
//
// 示例:
//
//	m := errors.MessageCountMap{
//	    "文件未找到": 5,
//	    "权限被拒绝": 2,
//	}
//	agg := errors.NewAggregateFromCountMap(m)
//	fmt.Println(agg.Error()) // 输出将包括计数
func NewAggregateFromCountMap(m MessageCountMap) Aggregate {
	if m == nil {
		return nil
	}
	result := make([]error, 0, len(m))
	for errStr, count := range m {
		var countStr string
		if count > 1 {
			countStr = fmt.Sprintf(" (重复 %v 次)", count)
		}
		result = append(result, fmt.Errorf("%v%v", errStr, countStr))
	}
	return NewAggregate(result)
}

// ErrPreconditionViolated 在前置条件被违反时返回
var ErrPreconditionViolated = errors.New("前置条件被违反")
