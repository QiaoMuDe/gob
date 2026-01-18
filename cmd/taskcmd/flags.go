package taskcmd

import (
	"gitee.com/MM-Q/qflag"
)

// 任务命令相关的标志变量
var (
	TaskCmd   *qflag.Cmd        // 任务命令
	listFlag  *qflag.BoolFlag   // 列出任务标志
	runFlag   *qflag.StringFlag // 运行任务标志
	initFlag  *qflag.BoolFlag   // 初始化配置文件标志
	forceFlag *qflag.BoolFlag   // 强制覆盖标志
	fileFlag  *qflag.StringFlag // 指定任务文件路径标志
)

// init 初始化任务命令及其标志
func init() {
	// 创建任务命令
	TaskCmd = qflag.NewCmd("task", "t", qflag.ExitOnError)

	// 注册任务命令的标志（不是全局标志）
	listFlag = TaskCmd.Bool("list", "l", false, "列出所有可用任务")
	runFlag = TaskCmd.String("run", "r", "", "运行指定任务")
	initFlag = TaskCmd.Bool("init", "i", false, "初始化任务配置文件")
	forceFlag = TaskCmd.Bool("force", "f", false, "强制覆盖已存在文件")
	fileFlag = TaskCmd.String("file", "", "", "指定任务配置文件路径")

	// 配置任务命令
	taskCmdCfg := qflag.CmdConfig{
		Desc:       "任务编排工具",
		UseChinese: true,
		Examples: []qflag.ExampleInfo{
			{Desc: "初始化任务配置", Usage: "gob task --init"},
			{Desc: "列出所有任务", Usage: "gob task --list"},
			{Desc: "运行指定任务", Usage: "gob task --run deploy"},
			{Desc: "指定任务文件运行任务", Usage: "gob task --run deploy --file custom.toml"},
			{Desc: "强制覆盖配置文件", Usage: "gob task --init --force"},
		},
		Notes: []string{
			"默认的任务编排文件名: task.toml, Task.toml",
			"未指定任务文件时，默认使用当前目录下的任务配置文件",
		},
	}
	TaskCmd.ApplyConfig(taskCmdCfg)
	TaskCmd.SetRun(run)
}
