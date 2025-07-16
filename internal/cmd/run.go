// run.go
package cmd

import (
	"gitee.com/MM-Q/gob/internal/globls"
)

func Run() {
	// 默认关闭颜色输出
	if !colorFlag.Get() {
		globls.CL.SetNoColor(true)
	}

	globls.CL.PrintDbg("gob 构建工具")
}
