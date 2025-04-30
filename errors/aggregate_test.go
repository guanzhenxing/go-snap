package errors

import (
	"fmt"
	"strings"
	"testing"
)

// 测试错误聚合
func TestAggregate(t *testing.T) {
	err1 := New("错误1")
	err2 := New("错误2")

	// 创建聚合
	agg := NewAggregate([]error{err1, err2})

	if agg == nil {
		t.Fatal("NewAggregate应该返回一个错误，但返回了nil")
	}

	// 检查聚合中的错误数量
	if len(agg.Errors()) != 2 {
		t.Errorf("聚合错误数量错误: 期望 2, 得到 %d", len(agg.Errors()))
	}

	// 检查错误字符串
	errStr := agg.Error()
	if !strings.Contains(errStr, "错误1") || !strings.Contains(errStr, "错误2") {
		t.Errorf("聚合错误消息应该包含所有错误: %q", errStr)
	}

	// 测试空聚合
	emptyAgg := NewAggregate(nil)
	if emptyAgg != nil {
		t.Error("空错误列表应该返回nil聚合")
	}

	// 测试仅包含nil的聚合
	nilAgg := NewAggregate([]error{nil})
	if nilAgg != nil {
		t.Error("只包含nil的错误列表应该返回nil聚合")
	}
}

// 测试错误聚合的Flatten
func TestFlatten(t *testing.T) {
	err1 := New("错误1")
	err2 := New("错误2")
	err3 := New("错误3")

	// 创建嵌套聚合
	innerAgg := NewAggregate([]error{err1, err2})
	outerAgg := NewAggregate([]error{err3, innerAgg})

	// 展平聚合
	flattened := Flatten(outerAgg)

	// 检查展平后的错误数量
	if len(flattened.Errors()) != 3 {
		t.Errorf("展平后错误数量错误: 期望 3, 得到 %d", len(flattened.Errors()))
	}

	// 检查所有错误是否都存在
	errs := flattened.Errors()
	found := make(map[string]bool)
	for _, err := range errs {
		found[err.Error()] = true
	}

	if !found["错误1"] || !found["错误2"] || !found["错误3"] {
		t.Error("展平的错误应该包含所有原始错误")
	}

	// 测试nil聚合的Flatten
	nilResult := Flatten(nil)
	if nilResult != nil {
		t.Error("Flatten(nil)应该返回nil")
	}
}

// 测试Reduce
func TestReduce(t *testing.T) {
	err1 := New("单个错误")

	// 创建只有一个错误的聚合
	agg := NewAggregate([]error{err1})

	// 减少聚合
	reduced := Reduce(agg)

	// 检查减少后是否等于原始错误
	if reduced != err1 {
		t.Errorf("减少后错误错误: 期望 %q, 得到 %q", err1, reduced)
	}

	// 测试空聚合
	emptyAgg := NewAggregate([]error{})
	reducedEmpty := Reduce(emptyAgg)
	if reducedEmpty != nil {
		t.Error("减少空聚合应该返回nil")
	}

	// 测试多个错误的聚合
	err2 := New("错误2")
	multiAgg := NewAggregate([]error{err1, err2})
	reducedMulti := Reduce(multiAgg)

	// 检查是否仍然是聚合类型
	multiAggResult, ok := reducedMulti.(Aggregate)
	if !ok {
		t.Error("减少多个错误的聚合应该返回Aggregate类型")
		return
	}

	// 检查错误数量是否相同
	if len(multiAggResult.Errors()) != 2 {
		t.Errorf("减少后的聚合应该有2个错误，但有 %d 个", len(multiAggResult.Errors()))
	}

	// 检查错误内容是否相同
	errs := multiAggResult.Errors()
	if errs[0].Error() != err1.Error() || errs[1].Error() != err2.Error() {
		t.Error("减少后的聚合应该包含相同的错误")
	}

	// 测试直接错误的减少
	directErr := fmt.Errorf("直接错误")
	reducedDirect := Reduce(directErr)
	if reducedDirect != directErr {
		t.Error("减少直接错误应该返回相同的错误")
	}
}

// 测试NewAggregateFromCountMap
func TestNewAggregateFromCountMap(t *testing.T) {
	countMap := MessageCountMap{
		"文件未找到": 3,
		"权限被拒绝": 1,
	}

	agg := NewAggregateFromCountMap(countMap)

	if agg == nil {
		t.Fatal("NewAggregateFromCountMap应该返回一个聚合错误")
	}

	// 检查错误消息
	errorMsg := agg.Error()
	if !strings.Contains(errorMsg, "文件未找到") || !strings.Contains(errorMsg, "权限被拒绝") {
		t.Errorf("聚合错误消息应该包含所有错误: %q", errorMsg)
	}

	// 检查重复计数
	if !strings.Contains(errorMsg, "重复 3 次") {
		t.Errorf("聚合错误消息应该包含重复计数: %q", errorMsg)
	}

	// 测试空映射
	emptyAgg := NewAggregateFromCountMap(nil)
	if emptyAgg != nil {
		t.Error("空计数映射应该返回nil聚合")
	}
}

// 测试聚合错误的Is方法
func TestAggregateIs(t *testing.T) {
	targetErr := fmt.Errorf("目标错误")
	err1 := New("错误1")
	err2 := targetErr

	// 创建包含目标错误的聚合
	agg := NewAggregate([]error{err1, err2})

	// 测试Is方法
	if !agg.Is(targetErr) {
		t.Error("聚合的Is方法应该识别包含的错误")
	}

	// 测试不包含的错误
	otherErr := fmt.Errorf("其他错误")
	if agg.Is(otherErr) {
		t.Error("聚合的Is方法不应该匹配不包含的错误")
	}
}

// 测试单个错误的聚合格式化
func TestSingleErrorAggregateFormat(t *testing.T) {
	err := New("单个错误")
	agg := NewAggregate([]error{err})

	// 单个错误的聚合应该直接返回该错误的消息
	if agg.Error() != "单个错误" {
		t.Errorf("单个错误的聚合应该返回原始错误消息，但得到: %q", agg.Error())
	}
}

// 测试重复错误的聚合格式化
func TestDuplicateErrorsAggregateFormat(t *testing.T) {
	err1 := New("相同错误")
	err2 := New("相同错误")
	agg := NewAggregate([]error{err1, err2})

	// 重复错误应该只显示一次
	if agg.Error() != "相同错误" {
		t.Errorf("重复错误的聚合应该只显示一次错误消息，但得到: %q", agg.Error())
	}
}

// 测试聚合错误的ErrPreconditionViolated
func TestErrPreconditionViolated(t *testing.T) {
	if ErrPreconditionViolated == nil {
		t.Error("ErrPreconditionViolated不应该为nil")
	}

	if ErrPreconditionViolated.Error() != "前置条件被违反" {
		t.Errorf("ErrPreconditionViolated消息错误，得到: %q", ErrPreconditionViolated.Error())
	}
}

// 测试复杂嵌套聚合的Flatten
func TestComplexFlatten(t *testing.T) {
	err1 := New("错误1")
	err2 := New("错误2")
	err3 := New("错误3")
	err4 := New("错误4")

	// 创建复杂嵌套聚合
	innerAgg1 := NewAggregate([]error{err1, err2})
	innerAgg2 := NewAggregate([]error{err3})
	middleAgg := NewAggregate([]error{innerAgg1, innerAgg2})
	outerAgg := NewAggregate([]error{middleAgg, err4})

	// 展平聚合
	flattened := Flatten(outerAgg)

	// 检查展平后的错误数量
	if len(flattened.Errors()) != 4 {
		t.Errorf("复杂嵌套展平后错误数量错误: 期望 4, 得到 %d", len(flattened.Errors()))
	}

	// 检查所有错误是否都存在
	errs := flattened.Errors()
	found := make(map[string]bool)
	for _, err := range errs {
		found[err.Error()] = true
	}

	if !found["错误1"] || !found["错误2"] || !found["错误3"] || !found["错误4"] {
		t.Error("展平的错误应该包含所有原始错误")
	}
}
