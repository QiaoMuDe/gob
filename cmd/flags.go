package cmd

import (
	"gitee.com/MM-Q/qflag"
)

var (
	// generateConfigFlag --generate-config, -gcf 生成默认配置文件
	generateConfigFlag *qflag.BoolFlag
	// forceFlag --force, -f 强制操作（用于生成配置时覆盖已存在文件）
	forceFlag *qflag.BoolFlag
	// listFlag --list, -l 列出可用的构建任务
	listFlag *qflag.BoolFlag
	// runFlag --run 运行指定的构建任务（自动在 gobf/ 目录下查找）
	runFlag *qflag.StringFlag
)
