package errors

import (
	"fmt"
	"io"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
)

//=====================================================
// 堆栈捕获模式
//=====================================================

// StackCaptureMode 指定堆栈跟踪捕获的模式
type StackCaptureMode int

const (
	// StackCaptureModeNever 表示不捕获堆栈
	StackCaptureModeNever StackCaptureMode = iota

	// StackCaptureModeDeferred 表示惰性捕获堆栈（仅在需要时捕获）
	StackCaptureModeDeferred

	// StackCaptureModeImmediate 表示立即捕获堆栈
	StackCaptureModeImmediate

	// StackCaptureModeModeSampled 表示采样捕获堆栈（每N个错误捕获一次）
	StackCaptureModeModeSampled
)

// 全局配置
var (
	// DefaultStackCaptureMode 是默认的堆栈捕获模式
	DefaultStackCaptureMode = StackCaptureModeDeferred

	// DefaultStackDepth 是默认的堆栈捕获深度
	DefaultStackDepth = 32

	// SamplingRate 是采样捕获模式下的采样率（每N个错误捕获一次）
	SamplingRate = 10
)

// 采样计数器
var stackSampleCounter int32 = 0

// StackProvider 是提供堆栈跟踪功能的接口
type StackProvider interface {
	StackTrace() StackTrace
}

//=====================================================
// 堆栈帧与跟踪
//=====================================================

// Frame 表示堆栈帧内的程序计数器。
// 由于历史原因，如果将Frame解释为uintptr，
// 其值表示程序计数器 + 1。
type Frame uintptr

// pc 返回此帧的程序计数器；
// 多个帧可能有相同的PC值。
func (f Frame) pc() uintptr { return uintptr(f) - 1 }

// file 返回包含此Frame的PC的函数的
// 文件的完整路径。
func (f Frame) file() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return "unknown"
	}
	file, _ := fn.FileLine(f.pc())
	return file
}

// line 返回此Frame的PC的函数的
// 源代码的行号。
func (f Frame) line() int {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return 0
	}
	_, line := fn.FileLine(f.pc())
	return line
}

// name 返回此函数的名称（如果已知）。
func (f Frame) name() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return "unknown"
	}
	return fn.Name()
}

// Format 根据fmt.Formatter接口格式化帧。
//
//	%s    源文件
//	%d    源行号
//	%n    函数名
//	%v    等同于 %s:%d
//
// Format接受改变某些动词打印的标志，如下：
//
//	%+s   函数名和源文件相对于编译时GOPATH的路径
//	      用\n\t分隔 (<函数名>\n\t<路径>)
//	%+v   等同于 %+s:%d
func (f Frame) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		switch {
		case s.Flag('+'):
			io.WriteString(s, f.name())
			io.WriteString(s, "\n\t")
			io.WriteString(s, f.file())
		default:
			io.WriteString(s, path.Base(f.file()))
		}
	case 'd':
		io.WriteString(s, strconv.Itoa(f.line()))
	case 'n':
		io.WriteString(s, funcname(f.name()))
	case 'v':
		f.Format(s, 's')
		io.WriteString(s, ":")
		f.Format(s, 'd')
	}
}

// MarshalText 将堆栈跟踪帧格式化为文本字符串。输出与
// fmt.Sprintf("%+v", f)相同，但没有换行符或制表符。
func (f Frame) MarshalText() ([]byte, error) {
	name := f.name()
	if name == "unknown" {
		return []byte(name), nil
	}
	return []byte(fmt.Sprintf("%s %s:%d", name, f.file(), f.line())), nil
}

// StackTrace 是从内层（最新）到外层（最旧）的帧堆栈。
type StackTrace []Frame

// Format 根据fmt.Formatter接口格式化帧堆栈。
//
//	%s	列出堆栈中每个帧的源文件
//	%v	列出堆栈中每个帧的源文件和行号
//
// Format接受改变某些动词打印的标志，如下：
//
//	%+v   打印堆栈中每个帧的文件名、函数和行号。
func (st StackTrace) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case s.Flag('+'):
			for _, f := range st {
				io.WriteString(s, "\n")
				f.Format(s, verb)
			}
		case s.Flag('#'):
			fmt.Fprintf(s, "%#v", []Frame(st))
		default:
			st.formatSlice(s, verb)
		}
	case 's':
		st.formatSlice(s, verb)
	}
}

// formatSlice 将此StackTrace作为Frame切片格式化到给定缓冲区中，
// 仅在使用'%s'或'%v'调用时有效。
func (st StackTrace) formatSlice(s fmt.State, verb rune) {
	io.WriteString(s, "[")
	for i, f := range st {
		if i > 0 {
			io.WriteString(s, " ")
		}
		f.Format(s, verb)
	}
	io.WriteString(s, "]")
}

//=====================================================
// 传统堆栈捕获
//=====================================================

// stack 表示程序计数器的堆栈。
type stack []uintptr

func (s *stack) Format(st fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case st.Flag('+'):
			for _, pc := range *s {
				f := Frame(pc)
				fmt.Fprintf(st, "\n%+v", f)
			}
		}
	}
}

// StackTrace 返回堆栈的StackTrace表示。
func (s *stack) StackTrace() StackTrace {
	f := make([]Frame, len(*s))
	for i := 0; i < len(f); i++ {
		f[i] = Frame((*s)[i])
	}
	return f
}

// callers 返回当前goroutine的调用帧，跳过前3帧。
func callers() *stack {
	return callersWithDepth(3, DefaultStackDepth)
}

// funcname 从完整的函数名中提取短函数名。
func funcname(name string) string {
	i := strings.LastIndex(name, "/")
	name = name[i+1:]
	i = strings.Index(name, ".")
	if i == -1 {
		return name
	}
	return name[i+1:]
}

//=====================================================
// 惰性堆栈捕获
//=====================================================

// lazyStack 是延迟捕获堆栈的数据结构，仅在需要时捕获。
type lazyStack struct {
	callerSkip int       // 跳过的调用帧数
	pcs        []uintptr // 程序计数器集合
	frames     []Frame   // 帧缓存
	captured   bool      // 是否已捕获堆栈
}

// newLazyStack 创建一个新的懒惰堆栈捕获结构。
func newLazyStack(skip int) *lazyStack {
	return &lazyStack{callerSkip: skip}
}

// capture 实际捕获堆栈跟踪。
func (ls *lazyStack) capture() {
	if ls.captured {
		return
	}

	// 分配足够大的切片来存储整个堆栈跟踪
	ls.pcs = make([]uintptr, DefaultStackDepth)

	// 跳过runtime.Callers和此函数
	n := runtime.Callers(ls.callerSkip, ls.pcs)
	ls.pcs = ls.pcs[:n] // 截断为实际大小

	ls.captured = true
}

// StackTrace 返回延迟堆栈的StackTrace表示。
func (ls *lazyStack) StackTrace() StackTrace {
	ls.capture() // 确保已捕获堆栈

	if ls.frames == nil {
		ls.frames = make([]Frame, len(ls.pcs))
		for i, pc := range ls.pcs {
			ls.frames[i] = Frame(pc)
		}
	}

	return ls.frames
}

// Format 实现了fmt.Formatter接口。
func (ls *lazyStack) Format(s fmt.State, verb rune) {
	st := ls.StackTrace()
	st.Format(s, verb)
}

//=====================================================
// 堆栈创建工具函数
//=====================================================

// createStackProvider 根据当前堆栈捕获模式创建适当的堆栈提供者。
func createStackProvider() StackProvider {
	switch DefaultStackCaptureMode {
	case StackCaptureModeNever:
		return nil
	case StackCaptureModeImmediate:
		return callers()
	case StackCaptureModeDeferred:
		return newLazyStack(4) // 多跳过一帧以考虑此函数调用
	case StackCaptureModeModeSampled:
		// 使用原子操作增加计数器
		count := atomic.AddInt32(&stackSampleCounter, 1)
		if count%int32(SamplingRate) == 0 {
			return callers()
		}
		return nil
	default:
		return newLazyStack(4) // 默认为延迟模式
	}
}

// callersWithDepth 返回当前goroutine的调用帧，
// 跳过指定数量的帧并捕获指定深度。
func callersWithDepth(skip, depth int) *stack {
	var pcs [64]uintptr
	n := runtime.Callers(skip, pcs[:])
	var st stack = pcs[0:n]
	return &st
}
