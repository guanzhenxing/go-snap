package errors

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

// 测试Frame的基本功能
func TestFrame(t *testing.T) {
	var pc [1]uintptr
	n := runtime.Callers(1, pc[:])
	if n == 0 {
		t.Fatal("runtime.Callers返回0，无法获取当前帧")
	}

	frame := Frame(pc[0])

	// 测试pc方法
	if frame.pc() != pc[0]-1 {
		t.Errorf("Frame.pc()返回错误的程序计数器值")
	}

	// 测试file方法
	file := frame.file()
	if !strings.Contains(file, "stack_test.go") {
		t.Errorf("Frame.file()应该包含当前文件名，但得到: %s", file)
	}

	// 测试line方法
	line := frame.line()
	if line <= 0 {
		t.Errorf("Frame.line()应该返回正确的行号，但得到: %d", line)
	}

	// 测试name方法
	name := frame.name()
	if !strings.Contains(name, "TestFrame") {
		t.Errorf("Frame.name()应该包含当前函数名，但得到: %s", name)
	}
}

// 测试Frame的格式化
func TestFrameFormat(t *testing.T) {
	var pc [1]uintptr
	runtime.Callers(1, pc[:])
	frame := Frame(pc[0])

	// 测试%s格式化
	s := fmt.Sprintf("%s", frame)
	if !strings.Contains(s, "stack_test.go") {
		t.Errorf("使用%%s格式化Frame应该包含文件名，但得到: %s", s)
	}

	// 测试%d格式化
	d := fmt.Sprintf("%d", frame)
	if d == "" || d == "0" {
		t.Errorf("使用%%d格式化Frame应该包含行号，但得到: %s", d)
	}

	// 测试%n格式化
	n := fmt.Sprintf("%n", frame)
	if !strings.Contains(n, "TestFrameFormat") {
		t.Errorf("使用%%n格式化Frame应该包含函数名，但得到: %s", n)
	}

	// 测试%v格式化
	v := fmt.Sprintf("%v", frame)
	if !strings.Contains(v, "stack_test.go:") {
		t.Errorf("使用%%v格式化Frame应该包含文件名和行号，但得到: %s", v)
	}

	// 测试%+s格式化
	plus_s := fmt.Sprintf("%+s", frame)
	if !strings.Contains(plus_s, "TestFrameFormat") && !strings.Contains(plus_s, "stack_test.go") {
		t.Errorf("使用%%+s格式化Frame应该包含函数名和文件路径，但得到: %s", plus_s)
	}
}

// 测试StackTrace的格式化
func TestStackTraceFormat(t *testing.T) {
	// 使用callers创建一个stack
	st := callers().StackTrace()

	// 确保堆栈跟踪不为空
	if len(st) == 0 {
		t.Fatal("创建的堆栈跟踪为空")
	}

	// 测试%s格式化
	s := fmt.Sprintf("%s", st)
	if !strings.Contains(s, "[") || !strings.Contains(s, "]") {
		t.Errorf("使用%%s格式化StackTrace应该产生数组格式，但得到: %s", s)
	}

	// 测试%v格式化
	v := fmt.Sprintf("%v", st)
	if !strings.Contains(v, "[") || !strings.Contains(v, "]") {
		t.Errorf("使用%%v格式化StackTrace应该产生数组格式，但得到: %s", v)
	}

	// 测试%+v格式化
	plus_v := fmt.Sprintf("%+v", st)
	if !strings.Contains(plus_v, "\n") {
		t.Errorf("使用%%+v格式化StackTrace应该产生多行输出，但得到: %s", plus_v)
	}
}

// 测试MarshalText方法
func TestFrameMarshalText(t *testing.T) {
	var pc [1]uintptr
	runtime.Callers(1, pc[:])
	frame := Frame(pc[0])

	text, err := frame.MarshalText()
	if err != nil {
		t.Fatalf("Frame.MarshalText()返回错误: %v", err)
	}

	textStr := string(text)
	if !strings.Contains(textStr, "TestFrameMarshalText") || !strings.Contains(textStr, "stack_test.go") {
		t.Errorf("Frame.MarshalText()应该包含函数名和文件路径，但得到: %s", textStr)
	}
}

// 测试stack的Format方法
func TestStackFormat(t *testing.T) {
	s := callers()

	// 只测试+v标志，因为这是唯一实现的格式
	formatted := fmt.Sprintf("%+v", s)
	if !strings.Contains(formatted, "\n") {
		t.Errorf("使用%%+v格式化stack应该产生多行输出，但得到: %s", formatted)
	}
}

// 测试funcname函数
func TestFuncname(t *testing.T) {
	// 测试一些常见的函数名格式
	cases := []struct {
		input    string
		expected string
	}{
		{"github.com/user/package.Function", "Function"},
		{"main.function", "function"},
		{"runtime.goexit", "goexit"},
		// 根据实际代码行为，Type.Method的情况应该是返回'Type.Method'
		{"some/path/file.Type.Method", "Type.Method"},
	}

	for _, c := range cases {
		result := funcname(c.input)
		if result != c.expected {
			t.Errorf("funcname(%q)返回 %q，期望 %q", c.input, result, c.expected)
		}
	}
}

// 测试stack的StackTrace方法
func TestStackStackTrace(t *testing.T) {
	s := callers()
	st := s.StackTrace()

	if len(st) != len(*s) {
		t.Errorf("stack.StackTrace()应该返回相同长度的StackTrace，但得到 %d vs %d", len(st), len(*s))
	}

	for i := range st {
		if uintptr(st[i]) != (*s)[i] {
			t.Errorf("stack.StackTrace()[%d]值错误", i)
		}
	}
}

// 测试未知函数的Frame行为
func TestFrameUnknownFunc(t *testing.T) {
	// 创建一个指向无效位置的Frame
	frame := Frame(0)

	// 测试方法行为
	if frame.name() != "unknown" {
		t.Errorf("对于未知函数，Frame.name()应该返回'unknown'，但得到: %s", frame.name())
	}

	if frame.file() != "unknown" {
		t.Errorf("对于未知函数，Frame.file()应该返回'unknown'，但得到: %s", frame.file())
	}

	if frame.line() != 0 {
		t.Errorf("对于未知函数，Frame.line()应该返回0，但得到: %d", frame.line())
	}

	// 测试MarshalText方法
	text, err := frame.MarshalText()
	if err != nil {
		t.Fatalf("未知函数的Frame.MarshalText()返回错误: %v", err)
	}

	if string(text) != "unknown" {
		t.Errorf("未知函数的Frame.MarshalText()应该返回'unknown'，但得到: %s", string(text))
	}
}

// 测试errorFormatInfo函数
func TestFormatDetailed(t *testing.T) {
	err := Wrap(fmt.Errorf("原始错误"), "包装错误")

	var str strings.Builder
	formatDetailed(err, &str, true)

	result := str.String()
	if !strings.Contains(result, "包装错误") {
		t.Errorf("formatDetailed应该包含错误消息，但得到: %s", result)
	}

	if !strings.Contains(result, ";") {
		t.Errorf("formatDetailed启用堆栈应该包含错误链分隔符，但得到: %s", result)
	}
}

// 测试unwrapErrorChain函数
func TestUnwrapErrorChain(t *testing.T) {
	// 创建一个错误链
	baseErr := fmt.Errorf("基础错误")
	err1 := Wrap(baseErr, "错误1")
	err2 := Wrap(err1, "错误2")

	// 解开错误链
	chain := unwrapErrorChain(err2)

	// 检查链中的错误数量
	if len(chain) != 3 {
		t.Errorf("错误链应该包含3个错误，但得到: %d", len(chain))
	}

	// 检查链中的错误顺序
	if chain[0] != err2 || chain[2] != baseErr {
		t.Error("错误链顺序错误")
	}

	// 测试nil错误
	nilChain := unwrapErrorChain(nil)
	if len(nilChain) != 0 {
		t.Errorf("nil错误的链应该为空，但得到长度: %d", len(nilChain))
	}
}

// 测试extractErrorFormatInfo函数
func TestExtractErrorFormatInfo(t *testing.T) {
	// 测试contextualError
	ceErr := New("上下文错误")
	ceInfo := extractErrorFormatInfo(ceErr)

	if ceInfo.err != "上下文错误" {
		t.Errorf("提取的错误消息错误，期望'上下文错误'，得到: %s", ceInfo.err)
	}

	if ceInfo.code != UnknownError.Code() {
		t.Errorf("提取的错误码错误，期望 %d，得到: %d", UnknownError.Code(), ceInfo.code)
	}

	// 测试标准错误
	stdErr := fmt.Errorf("标准错误")
	stdInfo := extractErrorFormatInfo(stdErr)

	if stdInfo.err != "标准错误" {
		t.Errorf("提取的标准错误消息错误，期望'标准错误'，得到: %s", stdInfo.err)
	}

	if stdInfo.code != UnknownError.Code() {
		t.Errorf("提取的标准错误码错误，期望 %d，得到: %d", UnknownError.Code(), stdInfo.code)
	}
}

//=====================================================
// 惰性堆栈测试
//=====================================================

// 测试惰性堆栈是否正确延迟捕获堆栈
func TestLazyStack(t *testing.T) {
	// 使用自定义函数来创建惰性堆栈
	var ls *lazyStack
	captureStack := func() *lazyStack {
		// 在这个函数中创建堆栈，这样我们可以确保它包含在堆栈跟踪中
		s := newLazyStack(2)
		return s
	}

	// 调用函数捕获堆栈
	ls = captureStack()

	// 堆栈还未捕获
	if ls.captured {
		t.Error("新创建的lazyStack应该还未捕获堆栈")
	}

	// 获取堆栈跟踪，此时应该触发堆栈捕获
	st := ls.StackTrace()

	// 确认堆栈已捕获
	if !ls.captured {
		t.Error("调用StackTrace()后应该已捕获堆栈")
	}

	// 确认堆栈跟踪不为空
	if len(st) == 0 {
		t.Error("堆栈跟踪不应为空")
	}

	// 检查堆栈跟踪包含函数名
	found := false
	for _, frame := range st {
		file := frame.file()
		if strings.Contains(file, "stack_test.go") {
			found = true
			break
		}
	}

	if !found {
		frames := make([]string, len(st))
		for i, frame := range st {
			frames[i] = frame.file()
		}
		t.Errorf("堆栈跟踪应包含当前文件名，但得到: %v", frames)
	}
}

// 测试不同的堆栈捕获模式
func TestStackCaptureMode(t *testing.T) {
	// 保存原始配置
	originalMode := DefaultStackCaptureMode
	defer func() {
		DefaultStackCaptureMode = originalMode
	}()

	// 测试Never模式
	DefaultStackCaptureMode = StackCaptureModeNever
	stack := createStackProvider()
	if stack != nil {
		t.Error("StackCaptureModeNever模式应返回nil堆栈提供者")
	}

	// 测试Immediate模式
	DefaultStackCaptureMode = StackCaptureModeImmediate
	stack = createStackProvider()
	if stack == nil {
		t.Error("StackCaptureModeImmediate模式不应返回nil堆栈提供者")
	}
	// 检查是否是立即堆栈
	if _, ok := stack.(*lazyStack); ok {
		t.Error("StackCaptureModeImmediate模式不应返回lazyStack类型")
	}

	// 测试Deferred模式
	DefaultStackCaptureMode = StackCaptureModeDeferred
	stack = createStackProvider()
	if stack == nil {
		t.Error("StackCaptureModeDeferred模式不应返回nil堆栈提供者")
	}
	_, ok := stack.(*lazyStack)
	if !ok {
		t.Error("StackCaptureModeDeferred模式应返回*lazyStack类型")
	}
}

// 测试创建并使用不同模式的错误
func TestNewWithStackControl(t *testing.T) {
	// 测试从不捕获堆栈
	err := NewWithStackControl("无堆栈错误", StackCaptureModeNever)
	if st, ok := err.(StackTracer); ok && st.StackTrace() != nil {
		t.Error("StackCaptureModeNever模式的错误不应有堆栈跟踪")
	}

	// 测试立即捕获堆栈
	err = NewWithStackControl("立即堆栈错误", StackCaptureModeImmediate)
	st, ok := err.(StackTracer)
	if !ok || st.StackTrace() == nil {
		t.Error("StackCaptureModeImmediate模式的错误应有堆栈跟踪")
	}

	// 测试延迟捕获堆栈
	err = NewWithStackControl("延迟堆栈错误", StackCaptureModeDeferred)
	st, ok = err.(StackTracer)
	if !ok || st.StackTrace() == nil {
		t.Error("StackCaptureModeDeferred模式的错误应有堆栈跟踪")
	}
}

//=====================================================
// 性能基准测试
//=====================================================

// 性能比较: 传统堆栈捕获 vs 惰性堆栈捕获
func BenchmarkStackCapture(b *testing.B) {
	b.Run("Immediate", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = callers()
		}
	})

	b.Run("Lazy", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = newLazyStack(3)
		}
	})
}

// 性能比较: 创建普通错误 vs 惰性堆栈错误 vs 无堆栈错误
func BenchmarkErrorCreation(b *testing.B) {
	// 保存原始配置
	originalMode := DefaultStackCaptureMode
	defer func() {
		DefaultStackCaptureMode = originalMode
	}()

	b.Run("Traditional", func(b *testing.B) {
		DefaultStackCaptureMode = StackCaptureModeImmediate
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = New("test error")
		}
	})

	b.Run("Lazy", func(b *testing.B) {
		DefaultStackCaptureMode = StackCaptureModeDeferred
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = New("test error")
		}
	})

	b.Run("None", func(b *testing.B) {
		DefaultStackCaptureMode = StackCaptureModeNever
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = New("test error")
		}
	})
}

// 堆栈格式化性能比较: 立即捕获 vs 惰性捕获
func BenchmarkStackFormatting(b *testing.B) {
	imm := callers()
	lazy := newLazyStack(3)
	// 确保已捕获堆栈
	_ = lazy.StackTrace()

	b.Run("Immediate", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = fmt.Sprintf("%+v", imm)
		}
	})

	b.Run("Lazy", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = fmt.Sprintf("%+v", lazy)
		}
	})
}

// 测试lazyStack的Format方法
func TestLazyStackFormat(t *testing.T) {
	ls := newLazyStack(2)

	// 使用%+v格式化
	formatted := fmt.Sprintf("%+v", ls)
	if !strings.Contains(formatted, "\n") {
		t.Errorf("使用%%+v格式化lazyStack应该产生多行输出，但得到: %s", formatted)
	}

	// 确认堆栈已被捕获
	if !ls.captured {
		t.Error("格式化后应该已捕获堆栈")
	}
}

// 测试callersWithDepth函数
func TestCallersWithDepth(t *testing.T) {
	st := callersWithDepth(2, 10)

	// 检查返回的堆栈不为空
	if len(*st) == 0 {
		t.Error("callersWithDepth应该返回非空堆栈")
	}

	// 测试不同的深度参数
	st2 := callersWithDepth(2, 5)
	if len(*st2) > 5 {
		t.Error("callersWithDepth深度限制不生效")
	}
}

// 测试所有堆栈捕获模式
func TestAllStackCaptureModes(t *testing.T) {
	// 保存原始配置
	originalMode := DefaultStackCaptureMode
	defer func() {
		DefaultStackCaptureMode = originalMode
	}()

	modes := []StackCaptureMode{
		StackCaptureModeNever,
		StackCaptureModeImmediate,
		StackCaptureModeDeferred,
		StackCaptureModeModeSampled,
	}

	for _, mode := range modes {
		DefaultStackCaptureMode = mode
		stack := createStackProvider()

		// 根据模式验证结果
		switch mode {
		case StackCaptureModeNever:
			if stack != nil {
				t.Errorf("StackCaptureModeNever模式应返回nil，但得到: %T", stack)
			}
		case StackCaptureModeImmediate:
			if stack == nil || reflect.TypeOf(stack).String() == "*errors.lazyStack" {
				t.Errorf("StackCaptureModeImmediate模式不应返回nil或lazyStack，但得到: %T", stack)
			}
		case StackCaptureModeDeferred:
			if stack == nil || reflect.TypeOf(stack).String() != "*errors.lazyStack" {
				t.Errorf("StackCaptureModeDeferred模式应返回lazyStack，但得到: %T", stack)
			}
		case StackCaptureModeModeSampled:
			// 采样模式可能返回nil或stack，不做具体断言
		}
	}
}
