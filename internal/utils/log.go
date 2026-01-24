package utils

import (
	"fmt"

	"gitee.com/MM-Q/colorlib"
)

var (
	// CL 颜色实例
	CL = colorlib.GetCL()
)

// PrintPrefix 打印前缀
const PrintPrefix = "gob:"

func Log(msg string) {
	CL.Greenf("%s %s\n", PrintPrefix, msg)
}

func Logf(format string, args ...interface{}) {
	CL.Greenf("%s %s", PrintPrefix, fmt.Sprintf(format, args...))
}

func TaskLog(taskName, msg string) {
	CL.Greenf("%s [%s] %s\n", PrintPrefix, taskName, msg)
}

func TaskLogf(taskName, format string, args ...interface{}) {
	CL.Greenf("%s [%s] %s", PrintPrefix, taskName, fmt.Sprintf(format, args...))
}

func TaskErr(taskName, msg string) {
	CL.Redf("%s [%s] %s\n", PrintPrefix, taskName, msg)
}

func TaskErrf(taskName, format string, args ...interface{}) {
	CL.Redf("%s [%s] %s", PrintPrefix, taskName, fmt.Sprintf(format, args...))
}
