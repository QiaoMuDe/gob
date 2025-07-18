package cmd

import (
	"fmt"
	"os"
	"strings"

	"gitee.com/MM-Q/gob/internal/globls"
	"gitee.com/MM-Q/qflag"
	"gitee.com/MM-Q/verman"
)

var (
	envFlag                 *qflag.MapFlag    // --env, -e 指定环境变量
	outputFlag              *qflag.PathFlag   // --output, -o 指定输出目录
	nameFlag                *qflag.StringFlag // --name, -n 指定输出文件名
	mainFlag                *qflag.PathFlag   // --main, -m 指定入口文件
	ldflagsFlag             *qflag.StringFlag // --ldflags, -l 指定链接器标志
	vendorFlag              *qflag.BoolFlag   // --use-vendor, -uv 在编译时使用 vendor 目录
	gitFlag                 *qflag.BoolFlag   // --git, -g 在编译时注入 git 信息
	simpleNameFlag          *qflag.BoolFlag   // --simple-name, -sn 简单名称
	proxyFlag               *qflag.StringFlag // --proxy, -p 设置代理
	cgoFlag                 *qflag.BoolFlag   // --enable-cgo, -ec 启用cgo
	colorFlag               *qflag.BoolFlag   // --color, -c 启用颜色输出
	installFlag             *qflag.BoolFlag   // --install, -i 安装编译后的二进制文件
	forceFlag               *qflag.BoolFlag   // --force, -f 执行强制操作
	batchFlag               *qflag.BoolFlag   // --batch, -b 批量编译
	currentPlatformOnlyFlag *qflag.BoolFlag   // --current-platform-only, -cpo 仅编译当前平台
	zipFlag                 *qflag.BoolFlag   // --zip, -z 在编译时打包输出文件为 zip 文件
	installPathFlag         *qflag.PathFlag   // --install-path, -ip 指定安装路径
)

// isTestMode 判断当前是否为测试模式
//
// 原理：
//   - 判断可执行文件名是否以 .test.exe 结尾
func isTestMode() bool {
	// 获取当前可执行文件的路径
	exePath, err := os.Executable()
	if err != nil {
		return false
	}

	// 判断可执行文件名是否以 .test.exe 结尾
	return strings.HasSuffix(exePath, ".test.exe")
}

// init 初始化命令行参数
func init() {
	envFlag = qflag.Map("env", "e", map[string]string{}, "指定环境变量,格式为: key=value")
	outputFlag = qflag.Path("output", "o", globls.DefaultOutputDir, "指定输出目录")
	nameFlag = qflag.String("name", "n", globls.DefaultAppName, "指定输出文件名")
	mainFlag = qflag.Path("main", "m", globls.DefaultMainFile, "指定main文件")
	ldflagsFlag = qflag.String("ldflags", "l", globls.DefaultLDFlags, "指定链接器标志")
	vendorFlag = qflag.Bool("use-vendor", "uv", false, "在编译时使用 vendor 目录")
	gitFlag = qflag.Bool("git", "g", false, "在编译时注入 git 信息")
	simpleNameFlag = qflag.Bool("simple-name", "sn", false, "简单名称")
	proxyFlag = qflag.String("proxy", "p", globls.DefaultGoProxy, "设置go代理")
	cgoFlag = qflag.Bool("cgo", "", false, "启用cgo")
	colorFlag = qflag.Bool("color", "c", false, "启用颜色输出")
	batchFlag = qflag.Bool("batch", "b", false, "批量编译")
	zipFlag = qflag.Bool("zip", "z", false, "在编译时打包输出文件为 zip 文件")
	installFlag = qflag.Bool("install", "i", false, "安装编译后的二进制文件")
	forceFlag = qflag.Bool("force", "f", false, "执行强制操作")
	currentPlatformOnlyFlag = qflag.Bool("current-platform-only", "cpo", false, "仅编译当前平台")
	installPathFlag = qflag.Path("install-path", "ip", getDefaultInstallPath(), "指定安装路径, 优先于GOPATH环境变量")

	// 设置命令行工具的描述
	qflag.SetDescription("gob 构建工具 - 支持自定义安装路径和跨平台构建的Go项目构建工具")

	// 启用自动补全
	qflag.SetEnableCompletion(true)

	// 启用中文
	qflag.SetUseChinese(true)

	// 设置版本信息
	v := verman.Get()
	qflag.SetVersion(fmt.Sprintf("%s %s", v.AppName, v.GitVersion))

	// 解析命令行参数 - 仅在非测试模式下执行
	if !isTestMode() {
		if err := qflag.Parse(); err != nil {
			fmt.Printf("err: %v\n", err)
			os.Exit(1)
		}
	}
}
