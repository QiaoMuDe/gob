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

	// 第一阶段：执行检查和准备

	// 第二阶段：根据参数处理或获取git信息

	// 第三阶段：根据参数执行构建或打包
}
