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
