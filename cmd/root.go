// run.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"gitee.com/MM-Q/gob/cmd/initcmd"
	"gitee.com/MM-Q/gob/internal/types"
	"gitee.com/MM-Q/gob/internal/utils"
	"gitee.com/MM-Q/qflag"
	"gitee.com/MM-Q/verman"
)

// InitAndRun 初始化并运行命令行参数
func InitAndRun() {
	// 注册全局标志
	generateConfigFlag = qflag.Root.Bool("generate-config", "gcf", false, "生成默认配置文件")
	forceFlag = qflag.Root.Bool("force", "f", false, "强制操作（覆盖已存在文件）")
	listFlag = qflag.Root.Bool("list", "l", false, "列出可用的构建任务")
	runFlag = qflag.Root.String("run", "", "", "运行指定的构建任务（自动在 gobf/ 目录下查找）")

	// 设置命令行工具的配置
	rootCmdCfg := qflag.CmdConfig{
		UsageSyntax: fmt.Sprintf("%s [options] [build-file]", filepath.Base(os.Args[0])),
		UseChinese:  true,
		Completion:  true,
		Desc:        "gob 构建工具 - 支持自定义安装路径和跨平台构建的Go项目构建工具",
		Version:     verman.V.Version(),
		Notes: []string{
			"[build-file] 指定gob配置文件路径, 默认为gob.toml",
			"所有构建参数必须通过配置文件指定，不再支持命令行参数",
		},
		Examples: []qflag.ExampleInfo{

			{
				Desc:  "生成默认配置文件 (gob.toml)",
				Usage: fmt.Sprintf("%s --generate-config", os.Args[0]),
			},
			{
				Desc:  "列出可用的构建任务",
				Usage: fmt.Sprintf("%s --list", os.Args[0]),
			},
			{
				Desc:  "运行指定的构建任务（快捷方式）",
				Usage: fmt.Sprintf("%s --run dev", os.Args[0]),
			},
			{
				Desc:  "使用指定配置文件构建",
				Usage: fmt.Sprintf("%s gobf/dev.toml", os.Args[0]),
			},
			{
				Desc:  "使用默认配置文件构建",
				Usage: os.Args[0],
			},
		},
	}
	qflag.ApplyConfig(rootCmdCfg)

	qflag.Root.SetRun(run)

	// 注册子命令
	if err := qflag.Root.AddSubCmd(initcmd.InitCmd); err != nil {
		fmt.Printf("err: %v\n", err)
		os.Exit(1)
	}

	// 解析命令行参数
	if err := qflag.ParseAndRoute(); err != nil {
		fmt.Printf("err: %v\n", err)
		os.Exit(1)
	}
}

// run 运行 gob 构建工具
func run(cmd *qflag.Cmd) error {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("%s panic: %v\nstack: %s\n", types.PrintPrefix, err, debug.Stack())
			os.Exit(1)
		}
	}()

	// 记录构建开始时间
	startTime := time.Now()
	defer func() {
		// 获取构建耗时
		duration := time.Since(startTime)
		// 格式化耗时为秒并保留两位小数
		utils.CL.Greenf("%s 本次构建耗时 %.2fs\n", types.PrintPrefix, duration.Seconds())
	}()

	// 处理--generate-config参数: 生成默认配置文件
	if generateConfigFlag.Get() {
		// 生成默认配置文件
		if err := types.GenerateDefaultConfig(forceFlag.Get()); err != nil {
			utils.CL.PrintError(err)
			os.Exit(1)
		}
		utils.CL.Greenf("%s 已生成构建文件: %s\n", types.PrintPrefix, types.GobBuildFile)
		os.Exit(0)
	}

	// 处理--list参数: 列出可用的构建任务
	if listFlag.Get() {
		if err := listBuildTasks(); err != nil {
			utils.CL.PrintError(err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// 声明配置文件路径变量
	var configFilePath string

	// 处理--run参数: 运行指定的构建任务（快捷方式）
	runTask := runFlag.Get()
	if runTask != "" {
		// 自动构建配置文件路径：gobf/<task-name>.toml
		configFilePath = filepath.Join("gobf", fmt.Sprintf("%s.toml", runTask))
	} else {
		// 获取非标志参数0作为配置文件路径
		configFilePath = filepath.Clean(qflag.Root.Arg(0))

		// 如果命令行参数0为空, 则使用默认配置文件路径
		if configFilePath == "" || configFilePath == "." {
			configFilePath = types.GobBuildFile
		}
	}

	// 检查配置文件是否存在
	if _, statErr := os.Stat(configFilePath); statErr != nil {
		utils.CL.PrintErrorf("配置文件 %s 不存在\n", configFilePath)
		utils.CL.Yellow("提示：")
		utils.CL.Yellow("  1. 运行 'gob init' 初始化构建配置 (生成 gobf/ 目录)")
		utils.CL.Yellow("  2. 运行 'gob --generate-config' 生成默认配置文件 (gob.toml)")
		utils.CL.Yellow("  3. 使用 'gob <配置文件路径>' 指定配置文件")
		utils.CL.Yellow("  4. 运行 'gob --list' 列出可用任务")
		utils.CL.Yellow("  5. 运行 'gob --run <任务名称>' 运行指定的构建任务")
		os.Exit(1)
	}

	// 创建配置结构体
	config := &types.GobConfig{}

	// 加载配置文件
	if err := loadAndValidateConfig(config, configFilePath); err != nil {
		utils.CL.PrintError(err)
		os.Exit(1)
	}

	// 设置颜色输出
	utils.CL.SetColor(config.Build.UI.Color)
	utils.CL.Greenf("%s 配置文件: %s\n", types.PrintPrefix, configFilePath)

	// 第一阶段：执行检查和准备阶段
	utils.CL.Greenf("%s 开始构建准备\n", types.PrintPrefix)
	if err := utils.CheckBaseEnv(config); err != nil {
		utils.CL.PrintErrorf("%v\n", err)
		os.Exit(1)
	}

	// 检查批量构建和安装选项是否同时启用
	if config.Build.Target.Batch && config.Install.Install {
		utils.CL.PrintError("不能同时使用批量构建和安装选项")
		os.Exit(1)
	}

	// 检查安装和zip选项是否同时启用
	if config.Install.Install && config.Build.Output.Zip {
		utils.CL.PrintError("不能同时使用安装和zip选项")
		os.Exit(1)
	}

	// 第二阶段: 根据参数获取git信息
	if config.Build.Git.Inject {
		utils.CL.Greenf("%s 获取Git元数据\n", types.PrintPrefix)
		if err := utils.GetGitMetaData(config.Build.TimeoutDuration, verman.V, config); err != nil {
			utils.CL.PrintErrorf("Git信息获取失败: %v\n", err)
			os.Exit(1)
		}
	}

	// 如果不是批量模式, 强制设置为仅构建当前平台
	if !config.Build.Target.Batch {
		config.Build.Target.CurrentPlatformOnly = true
	}

	// 执行构建
	if err := buildBatch(verman.V, config); err != nil {
		utils.CL.PrintError(err.Error())
		os.Exit(1)
	}

	return nil
}
