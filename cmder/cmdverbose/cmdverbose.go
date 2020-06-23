package cmdverbose

import (
	"fmt"
	"io"
	"os"
)

const (
	// EnvVerbose 启用调试环境变量
	EnvVerbose = "CLOUD189_VERBOSE"
)

var (
	// IsVerbose 是否调试
	IsVerbose = os.Getenv(EnvVerbose) == "1"

	// Outputs 输出
	Outputs = []io.Writer{os.Stderr}
)

// CmdVerbose 调试
type CmdVerbose struct {
	Module string
}

func New(module string) *CmdVerbose {
	return &CmdVerbose{
		Module: module,
	}
}

// Info 提示
func (pv *CmdVerbose) Info(l string) {
	Verbosef("DEBUG: %s INFO: %s\n", pv.Module, l)
}

// Infof 提示, 格式输出
func (pv *CmdVerbose) Infof(format string, a ...interface{}) {
	Verbosef("DEBUG: %s INFO: %s", pv.Module, fmt.Sprintf(format, a...))
}

// Warn 警告
func (pv *CmdVerbose) Warn(l string) {
	Verbosef("DEBUG: %s WARN: %s\n", pv.Module, l)
}

// Warnf 警告, 格式输出
func (pv *CmdVerbose) Warnf(format string, a ...interface{}) {
	Verbosef("DEBUG: %s WARN: %s", pv.Module, fmt.Sprintf(format, a...))
}

// Verbosef 调试格式输出
func Verbosef(format string, a ...interface{}) (n int, err error) {
	if IsVerbose {
		for _, Output := range Outputs {
			n1, err := fmt.Fprintf(Output, TimePrefix()+" "+format, a...)
			n += n1
			if err != nil {
				return n, err
			}
		}
	}
	return
}

// Verboseln 调试输出一行
func Verboseln(a ...interface{}) (n int, err error) {
	if IsVerbose {
		for _, Output := range Outputs {
			n1, err := fmt.Fprint(Output, TimePrefix()+" ")
			n += n1
			if err != nil {
				return n, err
			}
			n2, err := fmt.Fprintln(Output, a...)
			n += n2
			if err != nil {
				return n, err
			}
		}
	}
	return
}
