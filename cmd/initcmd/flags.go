package initcmd

import "gitee.com/MM-Q/qflag"

var (
	InitCmd   *qflag.Cmd
	forceFlag *qflag.BoolFlag   // --force, -f 强制生成
	nameFlag  *qflag.StringFlag // --name, -n 指定生成的项目名称
)
