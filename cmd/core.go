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

	"gitee.com/MM-Q/comprx"
	"gitee.com/MM-Q/gob/internal/types"
	"gitee.com/MM-Q/gob/internal/utils"
	"gitee.com/MM-Q/shellx"
	"gitee.com/MM-Q/verman"
)

// executeCommands 执行命令列表
//
// 参数:
//   - commands: 要执行的命令列表
//   - exitOnError: 命令执行失败时是否退出程序
//   - config: 配置对象
//   - envs: 环境变量列表
//
// 返回值:
//   - error: 错误信息
func executeCommands(commands []string, exitOnError bool, config *types.GobConfig, envs []string) error {
	if len(commands) == 0 {
		return nil
	}

	// 准备环境变量
	var cmdEnvs []string
	if len(envs) > 0 {
		cmdEnvs = append(cmdEnvs, envs...)
	}

	// 添加配置文件中的环境变量
	if len(config.Env) > 0 {
		for k, v := range config.Env {
			cmdEnvs = append(cmdEnvs, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// 获取工作目录
	workDir := config.Build.WorkDir
	if workDir == "" {
		workDir = "."
	}

	// 执行每个命令
	for _, cmd := range commands {
		// 检查命令是否为空
		if strings.TrimSpace(cmd) == "" {
			continue
		}

		// 执行命令
		var err error
		if runtime.GOOS == "windows" {
			err = shellx.NewCmdStr(cmd).WithEnvs(cmdEnvs).WithWorkDir(workDir).WithShell(shellx.ShellPowerShell).Exec()
		} else {
			err = shellx.NewCmdStr(cmd).WithEnvs(cmdEnvs).WithWorkDir(workDir).WithShell(shellx.ShellSh).Exec()
		}

		if err != nil {
			if exitOnError {
				return fmt.Errorf("执行命令 '%s' 失败: %w", cmd, err)
			} else {
				// 打印错误但继续执行
				utils.CL.Redf("%s 执行命令 '%s' 失败: %v\n", types.PrintPrefix, cmd, err)
				continue
			}
		}
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
func buildSingle(ctx *types.BuildContext) error {
	// 1. 执行构建前命令
	if ctx.Config.Build.PreBuild.Enabled {
		if err := executeCommands(ctx.Config.Build.PreBuild.Commands, ctx.Config.Build.PreBuild.ExitOnError, ctx.Config, ctx.Env); err != nil {
			return fmt.Errorf("构建前命令执行失败: %w", err)
		}
	}

	// 2. 获取构建命令 - 创建副本避免修改全局模板
	buildCmds := make([]string, len(ctx.Config.Build.Command.Build))
	copy(buildCmds, ctx.Config.Build.Command.Build)

	// 生成输出路径
	outputPath := filepath.Join(ctx.Config.Build.Output.Dir, utils.GenOutputName(ctx.Config.Build.Output.Name, ctx.Config.Build.Output.Simple, ctx.VerMan.GitVersion, ctx.SysPlatform, ctx.SysArch, ctx.Config.Build.Target.Batch))

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

	// 3. 执行构建命令
	if runtime.GOOS == "windows" {
		if buildErr := shellx.NewCmds(buildCmds).WithTimeout(ctx.Config.Build.TimeoutDuration).WithEnvs(envs).WithShell(shellx.ShellPowerShell).Exec(); buildErr != nil {
			return buildErr
		}
	} else {
		if buildErr := shellx.NewCmds(buildCmds).WithTimeout(ctx.Config.Build.TimeoutDuration).WithEnvs(envs).WithShell(shellx.ShellSh).Exec(); buildErr != nil {
			return buildErr
		}
	}

	// 4. 执行构建后命令
	if ctx.Config.Build.PostBuild.Enabled {
		if err := executeCommands(ctx.Config.Build.PostBuild.Commands, ctx.Config.Build.PostBuild.ExitOnError, ctx.Config, ctx.Env); err != nil {
			return fmt.Errorf("构建后命令执行失败: %w", err)
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
func buildBatch(v *verman.Info, config *types.GobConfig) error {
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
						utils.CL.Greenf("%s 跳过非当前平台: %s/%s\n", types.PrintPrefix, platform, arch)
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
						fmt.Printf("%s panic: %v\nstack: %s\n", types.PrintPrefix, err, debug.Stack())
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
				ctx := &types.BuildContext{
					VerMan:      v,        // VerMan对象
					Env:         envs,     // 环境变量
					SysPlatform: platform, // 平台
					SysArch:     arch,     // 架构
					Config:      config,   // 配置
				}

				// 直接调用构建函数并处理错误
				if buildErr := buildSingle(ctx); buildErr != nil {
					printMutex.Lock()
					utils.CL.Redf("%s build %s/%s ✗ %v\n", types.PrintPrefix, platform, arch, buildErr)
					printMutex.Unlock()
				} else {
					printMutex.Lock()
					utils.CL.Greenf("%s build %s/%s ✓\n", types.PrintPrefix, platform, arch)
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
func installExecutable(executablePath string, c *types.GobConfig) error {
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
	utils.CL.Greenf("%s 已安装至: %s\n", types.PrintPrefix, targetPath)

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
func loadAndValidateConfig(config *types.GobConfig, configFilePath string) error {
	// 加载配置文件
	loadedConfig, err := types.LoadConfig(configFilePath)
	if err != nil {
		return fmt.Errorf("加载构建文件 %s 失败: %v", configFilePath, err)
	}

	// 将加载的配置复制到传入的config指针
	*config = *loadedConfig

	// 如果启用了安装选项, 则处理安装路径
	if config.Install.Install {
		// 如果安装路径为空或者为 $GOPATH/bin, 则使用默认安装路径
		if config.Install.InstallPath == "" || strings.EqualFold(config.Install.InstallPath, "$GOPATH/bin") {
			config.Install.InstallPath = utils.GetDefaultInstallPath() // 获取默认安装路径
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
		utils.CL.Yellowf("%s gobf 目录中没有找到 .toml 配置文件\n", types.PrintPrefix)
		return nil
	}

	// 输出任务列表（使用 task 风格：星号开头）
	utils.CL.Greenf("%s 可用的构建任务：\n", types.PrintPrefix)
	for _, task := range tasks {
		fmt.Printf("%s %-20s %s\n", utils.CL.Syellow("*"), utils.CL.Scyan(task.name), task.description)
	}

	// 输出使用提示
	utils.CL.Yellow("\nUsage: gob gobf/<task-name>.toml")
	utils.CL.Yellow("Usage: gob -run <task-name>")

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
