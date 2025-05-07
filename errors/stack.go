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

// Frame represents a program counter inside a stack frame.
// For historical reasons if Frame is interpreted as a uintptr
// its value represents the program counter + 1.
type Frame uintptr

// pc returns the program counter for this frame;
// multiple frames may have the same PC value.
func (f Frame) pc() uintptr { return uintptr(f) - 1 }

// file returns the full path to the file that contains the
// function for this Frame's pc.
func (f Frame) file() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return "unknown"
	}
	file, _ := fn.FileLine(f.pc())
	return file
}

// line returns the line number of source code of the
// function for this Frame's pc.
func (f Frame) line() int {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return 0
	}
	_, line := fn.FileLine(f.pc())
	return line
}

// name returns the name of this function, if known.
func (f Frame) name() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return "unknown"
	}
	return fn.Name()
}

// Format formats the frame according to the fmt.Formatter interface.
//
//	%s    source file
//	%d    source line
//	%n    function name
//	%v    equivalent to %s:%d
//
// Format accepts flags that alter the printing of some verbs, as follows:
//
//	%+s   function name and path of source file relative to the compile time
//	      GOPATH separated by \n\t (<funcname>\n\t<path>)
//	%+v   equivalent to %+s:%d
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

// MarshalText formats a stacktrace Frame as a text string. The output is the
// same as that of fmt.Sprintf("%+v", f), but without newlines or tabs.
func (f Frame) MarshalText() ([]byte, error) {
	name := f.name()
	if name == "unknown" {
		return []byte(name), nil
	}
	return []byte(fmt.Sprintf("%s %s:%d", name, f.file(), f.line())), nil
}

// StackTrace is stack of Frames from innermost (newest) to outermost (oldest).
type StackTrace []Frame

// Format formats the stack of Frames according to the fmt.Formatter interface.
//
//	%s	lists source files for each Frame in the stack
//	%v	lists the source file and line number for each Frame in the stack
//
// Format accepts flags that alter the printing of some verbs, as follows:
//
//	%+v   Prints filename, function, and line number for each Frame in the stack.
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

// formatSlice will format this StackTrace into the given buffer as a slice of
// Frame, only valid when called with '%s' or '%v'.
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

// stack represents a stack of program counters.
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

func (s *stack) StackTrace() StackTrace {
	f := make([]Frame, len(*s))
	for i := 0; i < len(f); i++ {
		f[i] = Frame((*s)[i])
	}
	return f
}

func callers() *stack {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	var st stack = pcs[0:n]
	return &st
}

// funcname removes the path prefix component of a function's name reported by func.Name().
func funcname(name string) string {
	i := strings.LastIndex(name, "/")
	name = name[i+1:]
	i = strings.Index(name, ".")
	return name[i+1:]
}

//=====================================================
// 惰性堆栈捕获
//=====================================================

// lazyStack 提供惰性堆栈捕获功能
type lazyStack struct {
	callerSkip int       // 跳过的调用帧数
	pcs        []uintptr // 程序计数器集合
	frames     []Frame   // 帧缓存
	captured   bool      // 是否已捕获堆栈
}

// newLazyStack 创建一个新的惰性堆栈
func newLazyStack(skip int) *lazyStack {
	return &lazyStack{
		callerSkip: skip,
		captured:   false,
	}
}

// capture 捕获堆栈
func (ls *lazyStack) capture() {
	if ls.captured {
		return
	}

	depth := DefaultStackDepth
	pcs := make([]uintptr, depth)
	// 使用存储的跳过层数
	n := runtime.Callers(ls.callerSkip, pcs)
	ls.pcs = pcs[0:n]
	ls.captured = true
}

// StackTrace 实现StackProvider接口
func (ls *lazyStack) StackTrace() StackTrace {
	ls.capture()

	if ls.frames == nil {
		ls.frames = make([]Frame, len(ls.pcs))
		for i := range ls.pcs {
			ls.frames[i] = Frame(ls.pcs[i])
		}
	}

	return ls.frames
}

// Format 实现fmt.Formatter接口
func (ls *lazyStack) Format(s fmt.State, verb rune) {
	ls.capture()

	switch verb {
	case 'v':
		switch {
		case s.Flag('+'):
			for _, pc := range ls.pcs {
				f := Frame(pc)
				fmt.Fprintf(s, "\n%+v", f)
			}
		}
	}
}

//=====================================================
// 堆栈创建工具函数
//=====================================================

// createStackProvider 根据配置创建堆栈提供者
func createStackProvider() StackProvider {
	switch DefaultStackCaptureMode {
	case StackCaptureModeNever:
		return nil
	case StackCaptureModeImmediate:
		return callers()
	case StackCaptureModeModeSampled:
		counter := atomic.AddInt32(&stackSampleCounter, 1)
		if counter%int32(SamplingRate) == 0 {
			return callers()
		}
		return nil
	default: // StackCaptureModeDeferred
		return newLazyStack(4)
	}
}

// callersWithDepth 以指定深度捕获调用栈
func callersWithDepth(skip, depth int) *stack {
	pcs := make([]uintptr, depth)
	n := runtime.Callers(skip, pcs)
	var st stack = pcs[0:n]
	return &st
}
