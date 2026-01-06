package cmd

import (
	"gitee.com/MM-Q/qflag"
)

var (
	envFlag                 *qflag.MapFlag      // --env, -e 指定环境变量
	outputFlag              *qflag.StringFlag   // --output, -o 指定输出目录
	nameFlag                *qflag.StringFlag   // --name, -n 指定输出文件名
	mainFlag                *qflag.StringFlag   // --main, -m 指定入口文件
	vendorFlag              *qflag.BoolFlag     // --use-vendor, -uv 在编译时使用 vendor 目录
	gitFlag                 *qflag.BoolFlag     // --git, -g 在编译时注入 git 信息
	simpleNameFlag          *qflag.BoolFlag     // --simple-name, -sn 简单名称
	proxyFlag               *qflag.StringFlag   // --proxy, -p 设置代理
	cgoFlag                 *qflag.BoolFlag     // --enable-cgo, -ec 启用cgo
	colorFlag               *qflag.BoolFlag     // --color, -c 启用颜色输出
	installFlag             *qflag.BoolFlag     // --install, -i 安装编译后的二进制文件
	forceFlag               *qflag.BoolFlag     // --force, -f 执行强制操作
	batchFlag               *qflag.BoolFlag     // --batch, -b 批量编译
	currentPlatformOnlyFlag *qflag.BoolFlag     // --current-platform-only, -cpo 仅编译当前平台
	zipFlag                 *qflag.BoolFlag     // --zip, -z 在编译时打包输出文件为 zip 文件
	installPathFlag         *qflag.StringFlag   // --install-path, -ip 指定安装路径
	generateConfigFlag      *qflag.BoolFlag     // --generate-config 生成默认配置文件
	testFlag                *qflag.BoolFlag     // --test 在构建前运行单元测试
	skipCheckFlag           *qflag.BoolFlag     // --skip-check, -sc 跳过构建前检查
	timeoutFlag             *qflag.DurationFlag // --timeout 构建超时时间
)
