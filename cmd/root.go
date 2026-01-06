// run.go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"gitee.com/MM-Q/comprx"
	"gitee.com/MM-Q/gob/cmd/initcmd"
	"gitee.com/MM-Q/gob/internal/globls"
	"gitee.com/MM-Q/qflag"
	"gitee.com/MM-Q/shellx"
	"gitee.com/MM-Q/verman"
)

// BuildContext 构建上下文, 封装构建所需的所有参数
type BuildContext struct {
	VerMan      *verman.Info // verman对象
	Env         []string     // 环境变量
	SysPlatform string       // 系统平台
	SysArch     string       // 系统架构
	Config      *gobConfig   // 配置对象
}

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
				Desc:  "初始化gob构建文件 (生成 gobf/ 目录)",
				Usage: fmt.Sprintf("%s init", os.Args[0]),
			},
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

	// 解析命令行参数 - 仅在非测试模式下执行
	if !isTestMode() {
		if err := qflag.ParseAndRoute(); err != nil {
			fmt.Printf("err: %v\n", err)
			os.Exit(1)
		}
	}
}

// run 运行 gob 构建工具
func run(cmd *qflag.Cmd) error {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("%s panic: %v\nstack: %s\n", globls.PrintPrefix, err, debug.Stack())
			os.Exit(1)
		}
	}()

	// 记录构建开始时间
	startTime := time.Now()
	defer func() {
		// 获取构建耗时
		duration := time.Since(startTime)
		// 格式化耗时为秒并保留两位小数
		globls.CL.Greenf("%s 本次构建耗时 %.2fs\n", globls.PrintPrefix, duration.Seconds())
	}()

	// 处理--generate-config参数: 生成默认配置文件
	if generateConfigFlag.Get() {
		// 生成默认配置
		defaultConfig := getDefaultConfig()

		// 生成默认配置文件
		if err := generateDefaultConfig(defaultConfig); err != nil {
			globls.CL.PrintErrorf("%v\n", err)
			os.Exit(1)
		}
		globls.CL.Greenf("%s 已生成构建文件: %s\n", globls.PrintPrefix, globls.GobBuildFile)
		os.Exit(0)
	}

	// 处理--list参数: 列出可用的构建任务
	if listFlag.Get() {
		if err := listBuildTasks(); err != nil {
			globls.CL.PrintErrorf("%v\n", err)
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
			configFilePath = globls.GobBuildFile
		}
	}

	// 检查配置文件是否存在
	if _, statErr := os.Stat(configFilePath); statErr != nil {
		globls.CL.PrintErrorf("配置文件 %s 不存在\n", configFilePath)
		globls.CL.Yellow("提示：")
		globls.CL.Yellow("  1. 运行 'gob init' 初始化构建配置 (生成 gobf/ 目录)")
		globls.CL.Yellow("  2. 运行 'gob --generate-config' 生成默认配置文件 (gob.toml)")
		globls.CL.Yellow("  3. 使用 'gob <配置文件路径>' 指定配置文件")
		globls.CL.Yellow("  4. 运行 'gob --list' 列出可用任务")
		globls.CL.Yellow("  5. 运行 'gob --run <任务名称>' 运行指定的构建任务")
		os.Exit(1)
	}

	// 创建配置结构体
	config := &gobConfig{}

	// 加载配置文件
	if err := loadAndValidateConfig(config, configFilePath); err != nil {
		globls.CL.PrintErrorf("%v\n", err)
		os.Exit(1)
	}

	// 设置颜色输出
	globls.CL.SetColor(config.Build.UI.Color)
	globls.CL.Greenf("%s 配置文件: %s\n", globls.PrintPrefix, configFilePath)

	// 第一阶段：执行检查和准备阶段
	globls.CL.Greenf("%s 开始构建准备\n", globls.PrintPrefix)
	if err := checkBaseEnv(config); err != nil {
		globls.CL.PrintErrorf("%v\n", err)
		os.Exit(1)
	}

	// 检查批量构建和安装选项是否同时启用
	if config.Build.Target.Batch && config.Install.Install {
		globls.CL.PrintErrorf("不能同时使用批量构建和安装选项")
		os.Exit(1)
	}

	// 检查安装和zip选项是否同时启用
	if config.Install.Install && config.Build.Output.Zip {
		globls.CL.PrintErrorf("不能同时使用安装和zip选项")
		os.Exit(1)
	}

	// 第二阶段: 根据参数获取git信息
	if config.Build.Git.Inject {
		globls.CL.Greenf("%s 获取Git元数据\n", globls.PrintPrefix)
		if err := getGitMetaData(config.Build.TimeoutDuration, verman.V, config); err != nil {
			globls.CL.PrintErrorf("Git信息获取失败: %v\n", err)
			os.Exit(1)
		}
	}

	// 如果不是批量模式, 强制设置为仅构建当前平台
	if !config.Build.Target.Batch {
		config.Build.Target.CurrentPlatformOnly = true
	}

	// 执行构建
	if err := buildBatch(verman.V, config); err != nil {
		globls.CL.PrintError(err.Error())
		os.Exit(1)
	}

	return nil
}

// buildSingle 执行单个平台和架构的构建
//
// 参数:
//   - ctx: 构建上下文, 包含所有构建所需的参数
//
// 返回值:
//   - error: 错误信息
func buildSingle(ctx *BuildContext) error {
	// 获取构建命令 - 创建副本避免修改全局模板
	buildCmds := make([]string, len(ctx.Config.Build.Command.Build))
	copy(buildCmds, ctx.Config.Build.Command.Build)

	// 生成输出路径
	outputPath := filepath.Join(ctx.Config.Build.Output.Dir, genOutputName(ctx.Config.Build.Output.Name, ctx.Config.Build.Output.Simple, ctx.VerMan.GitVersion, ctx.SysPlatform, ctx.SysArch, ctx.Config.Build.Target.Batch))

	// 动态替换命令中的占位符
	for i, cmd := range buildCmds {
		switch cmd {
		case "{{ldflags}}": // 替换链接器标志
			if ctx.Config.Build.Git.Inject {
				// 如果启用了Git信息注入, 则替换链接器标志
				buildCmds[i] = fmt.Sprintf("\"%s\"", replaceGitPlaceholders(ctx.Config.Build.Git.Ldflags, ctx.VerMan))
			} else {
				// 否则使用默认链接器标志
				buildCmds[i] = fmt.Sprintf("\"%s\"", ctx.Config.Build.Compiler.Ldflags)
			}

		case "{{output}}": // 替换输出路径
			buildCmds[i] = outputPath
		case "{{if UseVendor}}-mod=vendor{{end}}": // 条件添加vendor标志
			if ctx.Config.Build.Source.UseVendor {
				buildCmds[i] = "-mod=vendor" // 添加vendor标志
			} else {
				buildCmds[i] = "-mod=readonly" // 添加readonly标志
			}
		case "{{mainFile}}": // 替换入口文件
			buildCmds[i] = ctx.Config.Build.Source.MainFile
		}
	}

	// 在输出目录下检查即将生成的可执行文件是否存在, 存在则删除
	if _, err := os.Stat(outputPath); err == nil {
		if err := os.Remove(outputPath); err != nil {
			return fmt.Errorf("删除 %s 失败: %v, 请手动删除该文件后重试", outputPath, err)
		}
	}

	// 获取环境变量
	envs := ctx.Env

	// 如果指定了环境变量, 则添加环境变量
	if len(ctx.Config.Env) > 0 {
		for k, v := range ctx.Config.Env {
			envs = append(envs, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// 获取Go代理
	GOPROXY := fmt.Sprintf("GOPROXY=%s", ctx.Config.Build.Compiler.Proxy)

	// 添加Go代理
	envs = append(envs, GOPROXY)

	// 检查是否启用CGO
	if ctx.Config.Build.Compiler.EnableCgo {
		envs = append(envs, "CGO_ENABLED=1")
	} else {
		envs = append(envs, "CGO_ENABLED=0")
	}

	// 执行构建命令
	if runtime.GOOS == "windows" {
		if buildErr := shellx.NewCmds(buildCmds).WithTimeout(ctx.Config.Build.TimeoutDuration).WithEnvs(envs).WithShell(shellx.ShellPowerShell).Exec(); buildErr != nil {
			return buildErr
		}
	} else {
		if buildErr := shellx.NewCmds(buildCmds).WithTimeout(ctx.Config.Build.TimeoutDuration).WithEnvs(envs).WithShell(shellx.ShellSh).Exec(); buildErr != nil {
			return buildErr
		}
	}

	// 如果启用了安装选项, 则执行安装
	if ctx.Config.Install.Install {
		if err := installExecutable(outputPath, ctx.Config); err != nil {
			return fmt.Errorf("安装失败: %w", err)
		}
		return nil
	}

	// 在buildSingle函数中添加zip打包逻辑
	if ctx.Config.Build.Output.Zip {
		// 检查输出路径是否存在, 不存在则跳过
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			return fmt.Errorf("编译后的可执行文件不存在: %w", err)
		}

		// 处理文件名
		baseName := strings.TrimSuffix(outputPath, ".exe") // 去除.exe后缀
		zipPath := fmt.Sprint(baseName, ".zip")            // 添加.zip后缀

		// 删除目标zip文件, 避免重复打包
		if err := os.RemoveAll(zipPath); err != nil {
			return fmt.Errorf("删除历史zip文件失败: %w", err)
		}

		// 打包zip文件
		if err := comprx.Pack(zipPath, outputPath); err != nil {
			return fmt.Errorf("压缩zip文件失败: %w", err)
		}

		// 删除原始文件
		if err := os.RemoveAll(outputPath); err != nil {
			return fmt.Errorf("删除编译生成的文件 %s 失败: %w", outputPath, err)
		}
	}
	return nil
}

// buildBatch 执行批量构建
//
// 参数:
//   - v: verman对象
//   - config: 配置对象
//
// 返回值:
//   - error: 错误信息
func buildBatch(v *verman.Info, config *gobConfig) error {
	var wg sync.WaitGroup                                  // 用于同步goroutine
	var printMutex sync.Mutex                              // 用于同步打印输出
	maxConcurrency := runtime.NumCPU()                     // 使用CPU核心数作为默认并发数
	concurrencyChan := make(chan struct{}, maxConcurrency) // 控制并发数量的信号量

	// 获取根环境变量
	rootEnvs := os.Environ()

	// 根环境变量长度
	rootEnvLen := len(rootEnvs)

	// 遍历平台
	for _, platform := range config.Build.Target.Platforms {
		// 遍历架构
		for _, arch := range config.Build.Target.Architectures {
			// 跳过不支持的darwin/386和darwin/arm组合
			if platform == "darwin" && (arch == "386" || arch == "arm") {
				continue
			}

			// 如果开启了仅构建当前平台, 则跳过其他平台
			if config.Build.Target.CurrentPlatformOnly {
				if platform != runtime.GOOS || arch != runtime.GOARCH {
					printMutex.Lock()
					// 仅在批量模式下打印跳过信息
					if config.Build.Target.Batch {
						globls.CL.Greenf("%s 跳过非当前平台: %s/%s\n", globls.PrintPrefix, platform, arch)
					}
					printMutex.Unlock()
					continue
				}
			}

			// 获取并发信号量
			concurrencyChan <- struct{}{}

			// 启动goroutine执行并行构建
			wg.Go(func() {
				defer func() {
					<-concurrencyChan // 释放并发信号量
				}()

				defer func() {
					if err := recover(); err != nil {
						fmt.Printf("%s panic: %v\nstack: %s\n", globls.PrintPrefix, err, debug.Stack())
					}
				}()

				// 拷贝根环境变量
				envs := make([]string, rootEnvLen)
				copy(envs, rootEnvs)

				// 设置平台和架构
				GOOS := fmt.Sprintf("GOOS=%s", platform)
				GOARCH := fmt.Sprintf("GOARCH=%s", arch)

				// 添加环境变量
				envs = append(envs, GOOS, GOARCH)

				// 构建上下文
				ctx := &BuildContext{
					VerMan:      v,        // VerMan对象
					Env:         envs,     // 环境变量
					SysPlatform: platform, // 平台
					SysArch:     arch,     // 架构
					Config:      config,   // 配置
				}

				// 直接调用构建函数并处理错误
				if buildErr := buildSingle(ctx); buildErr != nil {
					printMutex.Lock()
					globls.CL.Redf("%s build %s/%s ✗ %v\n", globls.PrintPrefix, platform, arch, buildErr)
					printMutex.Unlock()
				} else {
					printMutex.Lock()
					globls.CL.Greenf("%s build %s/%s ✓\n", globls.PrintPrefix, platform, arch)
					printMutex.Unlock()
				}
			})
		}
	}

	// 等待所有goroutine完成
	wg.Wait()
	return nil
}

// installExecutable 将可执行文件安装到指定路径或GOPATH/bin目录
//
// 参数:
//   - executablePath: 要安装的可执行文件路径
//   - c: 配置对象
//
// 返回值:
//   - error: 错误信息
func installExecutable(executablePath string, c *gobConfig) error {
	// 获取安装路径
	binDir := c.Install.InstallPath

	// 检查可执行文件是否存在
	if _, err := os.Stat(executablePath); os.IsNotExist(err) {
		return fmt.Errorf("可执行文件不存在: %s", executablePath)
	}

	// 检查安装目录是否存在, 不存在则创建
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("创建安装目录失败: %w", err)
	}

	// 构建目标路径
	targetPath := filepath.Join(binDir, filepath.Base(executablePath))

	// 检查目标文件是否已存在
	if _, err := os.Stat(targetPath); err == nil {
		if !c.Install.Force {
			return fmt.Errorf("文件已存在: %s, 请在配置文件中设置 [install] force = true 以强制覆盖", targetPath)
		}
		// 强制删除现有文件
		if err := os.Remove(targetPath); err != nil {
			return fmt.Errorf("删除现有文件失败: %w", err)
		}
	}

	// 移动文件到目标路径
	if err := os.Rename(executablePath, targetPath); err != nil {
		return fmt.Errorf("移动文件失败: %w", err)
	}

	// 打印安装成功信息
	globls.CL.Greenf("%s 已安装至: %s\n", globls.PrintPrefix, targetPath)

	return nil
}

// loadAndValidateConfig 加载并验证配置文件
// 参数:
// - config: 指向配置结构体的指针, 用于存储加载的配置
// - configFilePath: 配置文件的路径
//
// 返回值:
//
//	error: 如果加载或验证过程中出现错误, 则返回错误信息
func loadAndValidateConfig(config *gobConfig, configFilePath string) error {
	// 加载配置文件
	loadedConfig, err := loadConfig(configFilePath)
	if err != nil {
		return fmt.Errorf("加载构建文件 %s 失败: %v", configFilePath, err)
	}

	// 将加载的配置复制到传入的config指针
	*config = *loadedConfig

	// 如果启用了安装选项, 则处理安装路径
	if config.Install.Install {
		// 如果安装路径为空或者为 $GOPATH/bin, 则使用默认安装路径
		if config.Install.InstallPath == "" || strings.EqualFold(config.Install.InstallPath, "$GOPATH/bin") {
			config.Install.InstallPath = getDefaultInstallPath() // 获取默认安装路径
		} else {
			// 处理自定义路径
			standardizedPath := filepath.ToSlash(config.Install.InstallPath) // 标准化路径
			normalizedPath := strings.TrimSuffix(standardizedPath, "/")      // 去除末尾的斜杠

			// 检查路径有效性
			if _, err := os.Stat(normalizedPath); err != nil {
				return fmt.Errorf("自定义安装路径 %s 无效: %v", normalizedPath, err)
			}

			// 更新为标准化后的路径
			config.Install.InstallPath = normalizedPath
		}
	}

	return nil
}

// replaceGitPlaceholders 将链接器标志中的占位符替换为实际的Git元数据
//
// 参数:
//   - ldflags: 包含占位符的链接器标志字符串
//   - v: 包含Git元数据的结构体
//
// 返回值:
//   - string: 替换后的链接器标志字符串
func replaceGitPlaceholders(ldflags string, v *verman.Info) string {
	// 定义占位符映射关系
	placeholders := map[string]string{
		"{{AppName}}":       v.AppName,
		"{{GitVersion}}":    v.GitVersion,
		"{{GitCommit}}":     v.GitCommit,
		"{{GitCommitTime}}": v.GitCommitTime,
		"{{BuildTime}}":     v.BuildTime,
		"{{GitTreeState}}":  v.GitTreeState,
	}

	// 替换所有占位符
	for placeholder, value := range placeholders {
		ldflags = strings.ReplaceAll(ldflags, placeholder, value)
	}

	return ldflags
}

// listBuildTasks 列出可用的构建任务
//
// 返回值:
//   - error: 错误信息
func listBuildTasks() error {
	// 检查 gobf 目录是否存在
	gobfDir := "gobf"
	if _, err := os.Stat(gobfDir); os.IsNotExist(err) {
		return fmt.Errorf("gobf 目录不存在，请先运行 'gob init' 初始化构建配置")
	}

	// 读取 gobf 目录下的所有文件
	entries, err := os.ReadDir(gobfDir)
	if err != nil {
		return fmt.Errorf("读取 gobf 目录失败: %w", err)
	}

	// 收集所有 .toml 文件及其描述
	type taskInfo struct {
		name        string
		description string
	}
	var tasks []taskInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if taskName, ok := strings.CutSuffix(name, ".toml"); ok {
			// 尝试从配置文件中提取描述
			description := extractTaskDescription(filepath.Join(gobfDir, name))
			tasks = append(tasks, taskInfo{
				name:        taskName,
				description: description,
			})
		}
	}

	// 如果没有找到任何任务
	if len(tasks) == 0 {
		globls.CL.Yellowf("%s gobf 目录中没有找到 .toml 配置文件\n", globls.PrintPrefix)
		return nil
	}

	// 输出任务列表（使用 task 风格：星号开头）
	globls.CL.Greenf("%s 可用的构建任务：\n", globls.PrintPrefix)
	for _, task := range tasks {
		fmt.Printf("* %-20s %s\n", task.name, task.description)
	}

	// 输出使用提示
	globls.CL.Yellow("\nUsage: gob gobf/<task-name>.toml")
	globls.CL.Yellow("Usage: gob -run <task-name>")

	return nil
}

// extractTaskDescription 从配置文件中提取描述信息
//
// 参数:
//   - configPath: 配置文件路径
//
// 返回值:
//   - string: 描述信息
func extractTaskDescription(configPath string) string {
	// 读取配置文件的第一行
	file, err := os.Open(configPath)
	if err != nil {
		return "Build task"
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 如果第一行以 # 开头，去除 # 符号
		if strings.HasPrefix(line, "#") {
			description := strings.TrimPrefix(line, "#")
			return strings.TrimSpace(description)
		}
	}

	return "Build task"
}
