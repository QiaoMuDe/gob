package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gitee.com/MM-Q/gob/internal/cmd/initcmd"
	"gitee.com/MM-Q/gob/internal/globls"
	"gitee.com/MM-Q/qflag"
	"gitee.com/MM-Q/verman"
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

// InitAndRun 初始化并运行命令行参数
func InitAndRun() {
	envFlag = qflag.Root.Map("env", "e", map[string]string{}, "指定环境变量,格式为: key=value")
	outputFlag = qflag.Root.String("output", "o", globls.DefaultOutputDir, "指定输出目录")
	nameFlag = qflag.Root.String("name", "n", globls.DefaultAppName, "指定输出文件名")
	mainFlag = qflag.Root.String("main", "m", globls.DefaultMainFile, "指定main文件")
	vendorFlag = qflag.Root.Bool("use-vendor", "uv", false, "在编译时使用 vendor 目录")
	gitFlag = qflag.Root.Bool("git", "g", false, "在编译时注入 git 信息")
	simpleNameFlag = qflag.Root.Bool("simple-name", "sn", false, "简单名称")
	proxyFlag = qflag.Root.String("proxy", "p", globls.DefaultGoProxy, "设置go代理")
	cgoFlag = qflag.Root.Bool("cgo", "", false, "启用cgo")
	colorFlag = qflag.Root.Bool("color", "c", false, "启用颜色输出")
	batchFlag = qflag.Root.Bool("batch", "b", false, "批量编译")
	zipFlag = qflag.Root.Bool("zip", "z", false, "在编译时打包输出文件为 zip 文件")
	installFlag = qflag.Root.Bool("install", "i", false, "安装编译后的二进制文件")
	forceFlag = qflag.Root.Bool("force", "f", false, "执行强制操作")
	currentPlatformOnlyFlag = qflag.Root.Bool("current-platform-only", "cpo", false, "仅编译当前平台")
	installPathFlag = qflag.Root.String("install-path", "ip", getDefaultInstallPath(), "指定安装路径, 优先于GOPATH环境变量")
	generateConfigFlag = qflag.Root.Bool("generate-config", "gcf", false, "生成默认配置文件")
	testFlag = qflag.Root.Bool("test", "t", false, "在构建前运行单元测试")
	skipCheckFlag = qflag.Root.Bool("skip-check", "sc", false, "跳过构建前检查")
	timeoutFlag = qflag.Root.Duration("timeout", "", 30*time.Second, "构建超时时间(秒)")

	// 设置命令行工具的配置
	rootCmdCfg := qflag.CmdConfig{
		UsageSyntax: fmt.Sprintf("%s [options] [build-file]", filepath.Base(os.Args[0])),
		UseChinese:  true,
		Completion:  true,
		Desc:        "gob 构建工具 - 支持自定义安装路径和跨平台构建的Go项目构建工具",
		Version:     verman.V.Version(),
		Notes:       []string{"[build-file] 可选参数, 指定gob配置文件路径, 默认为gob.toml", "默认在当前目录下寻找gob.toml构建文件, 如果不存在, 则使用命令行参数进行构建"},
		Examples: []qflag.ExampleInfo{
			{
				Desc:  "生成默认配置文件",
				Usage: fmt.Sprintf("%s -gcf", os.Args[0]),
			},
			{
				Desc:  "初始化gob构建文件",
				Usage: fmt.Sprintf("%s init", os.Args[0]),
			},
		},
	}
	qflag.ApplyConfig(rootCmdCfg)

	qflag.Root.SetRun(Run)

	// 注册子命令
	if err := qflag.Root.AddSubCmd(initcmd.InitCmd); err != nil {
		fmt.Printf("err: %v\n", err)
		os.Exit(1)
	}

	// 解析命令行参数 - 仅在非测试模式下执行
	if !isTestMode() {
		if err := qflag.ParseAndRoute(); err != nil {
			fmt.Printf("err: %v\n", err)
			os.Exit(1)
		}
	}
}
