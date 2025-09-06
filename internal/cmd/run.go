// run.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"gitee.com/MM-Q/gob/internal/globls"
	"gitee.com/MM-Q/qflag"
	"gitee.com/MM-Q/verman"
)

// BuildContext 构建上下文，封装构建所需的所有参数
type BuildContext struct {
	VerMan      *verman.VerMan // verman对象
	Env         []string       // 环境变量
	SysPlatform string         // 系统平台
	SysArch     string         // 系统架构
	Config      *gobConfig     // 配置对象
}

// Run 运行 gob 构建工具
func Run() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("panic: %v\nstack: %s\n", err, debug.Stack())
			os.Exit(1)
		}
	}()

	// 记录构建开始时间
	startTime := time.Now()
	defer func() {
		// 获取构建耗时
		duration := time.Since(startTime)
		// 格式化耗时为秒并保留两位小数
		globls.CL.Greenf("本次构建耗时 %.2fs\n", duration.Seconds())
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
		globls.CL.Greenf("已生成默认配置文件: %s\n", globls.GobBuildFile)
		os.Exit(0)
	}

	// 获取非标志参数0作为gob.toml的文件路径
	configFilePath := filepath.Clean(qflag.Arg(0))

	// 创建配置结构体
	config := &gobConfig{}

	// 如果命令行参数0为空, 则使用默认配置文件路径
	if configFilePath == "" || configFilePath == "." {
		configFilePath = globls.GobBuildFile
	}

	// 执行主逻辑: 检查gob.toml文件是否存在, 如果存在就读取配置,不存在则通过命令行参数获取配置
	if _, statErr := os.Stat(configFilePath); statErr == nil {
		// 如果存在, 则通过loadAndValidateConfig函数读取配置
		if err := loadAndValidateConfig(config, configFilePath); err != nil {
			globls.CL.PrintErrorf("%v\n", err)
			os.Exit(1)
		}
		// 默认关闭颜色输出
		globls.CL.SetColor(config.Build.ColorOutput)
		// 输出加载模式
		globls.CL.Greenf("BuildFile: %s\n", configFilePath)

	} else {
		// 如果不存在，则将命令行标志的值设置到配置结构体
		applyConfigFlags(config)
		// 默认关闭颜色输出
		globls.CL.SetColor(config.Build.ColorOutput)
		// 输出加载模式
		globls.CL.Green("CLI args")
	}

	// 获取verman对象
	v := verman.Get()

	// 第一阶段：执行检查和准备阶段
	globls.CL.Green("开始构建准备")
	if err := checkBaseEnv(config); err != nil {
		globls.CL.PrintErrorf("%v\n", err)
		os.Exit(1)
	}

	// 如果启用了测试选项，则运行单元测试
	if testFlag.Get() {
		globls.CL.Green("开始运行单元测试")
		if err := runTests(config.Build.TimeoutDuration); err != nil {
			globls.CL.PrintErrorf("%v\n", err)
			os.Exit(1)
		}
	}

	// 检查批量构建和安装选项是否同时启用
	if config.Build.BatchMode && config.Install.Install {
		globls.CL.PrintError("不能同时使用批量构建和安装选项")
		os.Exit(1)
	}

	// 检查安装和zip选项是否同时启用
	if config.Install.Install && config.Build.ZipOutput {
		globls.CL.PrintError("不能同时使用安装和zip选项")
		os.Exit(1)
	}

	// 第二阶段: 根据参数获取git信息
	if config.Build.InjectGitInfo {
		globls.CL.Green("获取Git元数据")
		if err := getGitMetaData(config.Build.TimeoutDuration, v, config); err != nil {
			globls.CL.PrintErrorf("Git信息获取失败: %v\n", err)
			os.Exit(1)
		}
	}

	// 第三阶段: 执行构建命令
	globls.CL.Green("开始构建")
	if config.Build.BatchMode {
		// 批量构建
		if err := buildBatch(v, config); err != nil {
			globls.CL.PrintError(err.Error())
			os.Exit(1)
		}
	} else {
		// 单个构建
		ctx := &BuildContext{
			VerMan:      v,
			Env:         os.Environ(),
			SysPlatform: runtime.GOOS,
			SysArch:     runtime.GOARCH,
			Config:      config,
		}
		if err := buildSingle(ctx); err != nil {
			globls.CL.PrintError(err.Error())
			os.Exit(1)
		}
	}
}

// buildSingle 执行单个平台和架构的构建
//
// 参数:
//   - ctx: 构建上下文，包含所有构建所需的参数
//
// 返回值:
//   - error: 错误信息
func buildSingle(ctx *BuildContext) error {
	// 获取构建命令 - 创建副本避免修改全局模板
	buildCmds := make([]string, len(ctx.Config.Build.BuildCommand))
	copy(buildCmds, ctx.Config.Build.BuildCommand)

	// 生成输出路径
	outputPath := filepath.Join(ctx.Config.Build.OutputDir, genOutputName(ctx.Config.Build.OutputName, ctx.Config.Build.SimpleName, ctx.VerMan.GitVersion, ctx.SysPlatform, ctx.SysArch))

	// 动态替换命令中的占位符
	for i, cmd := range buildCmds {
		switch cmd {
		case "{{ldflags}}": // 替换链接器标志
			if ctx.Config.Build.InjectGitInfo {
				// 如果启用了Git信息注入，则替换链接器标志
				buildCmds[i] = replaceGitPlaceholders(ctx.Config.Build.GitLdflags, ctx.VerMan)
			} else {
				// 否则使用默认链接器标志
				buildCmds[i] = ctx.Config.Build.Ldflags
			}

		case "{{output}}": // 替换输出路径
			buildCmds[i] = outputPath
		case "{{if UseVendor}}-mod=vendor{{end}}": // 条件添加vendor标志
			if ctx.Config.Build.UseVendor {
				buildCmds[i] = "-mod=vendor" // 添加vendor标志
			} else {
				buildCmds[i] = "-mod=readonly" // 添加readonly标志
			}
		case "{{mainFile}}": // 替换入口文件
			buildCmds[i] = ctx.Config.Build.MainFile
		}
	}

	// 在输出目录下检查即将生成的可执行文件是否存在，存在则删除
	if _, err := os.Stat(outputPath); err == nil {
		if err := os.Remove(outputPath); err != nil {
			// 退出并打印提示让用户手动删除
			globls.CL.PrintErrorf("删除 %s 失败: %v\n请手动删除该文件后重试\n", outputPath, err)
			os.Exit(1)
		}
	}

	// 获取环境变量
	envs := ctx.Env

	// 如果指定了环境变量，则添加环境变量
	if len(ctx.Config.Env) > 0 {
		for k, v := range ctx.Config.Env {
			envs = append(envs, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// 获取Go代理
	GOPROXY := fmt.Sprintf("GOPROXY=%s", ctx.Config.Build.Proxy)

	// 添加Go代理
	envs = append(envs, GOPROXY)

	// 检查是否启用CGO
	if ctx.Config.Build.EnableCgo {
		envs = append(envs, "CGO_ENABLED=1")
	} else {
		envs = append(envs, "CGO_ENABLED=0")
	}

	// 执行构建命令
	if result, buildErr := runCmd(ctx.Config.Build.TimeoutDuration, buildCmds, envs); buildErr != nil {
		return fmt.Errorf("build %s/%s ✗ Command: %s Error: %v Output: %s", ctx.SysPlatform, ctx.SysArch, buildCmds, buildErr, result)
	}

	// 构建成功
	globls.CL.Greenf("build %s/%s ✓\n", ctx.SysPlatform, ctx.SysArch)

	// 如果启用了安装选项，则执行安装
	if ctx.Config.Install.Install {
		if err := installExecutable(outputPath, ctx.Config); err != nil {
			return fmt.Errorf("安装失败: %w", err)
		}
		return nil
	}

	// 在buildSingle函数中添加zip打包逻辑
	if ctx.Config.Build.ZipOutput {
		// 检查输出路径是否存在, 不存在则跳过
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			return fmt.Errorf("编译后的可执行文件不存在: %w", err)
		}

		// 处理文件名
		baseName := strings.TrimSuffix(outputPath, ".exe") // 去除.exe后缀
		zipPath := fmt.Sprint(baseName, ".zip")            // 添加.zip后缀

		// 调用CreateZip函数创建zip文件
		if err := createZip(zipPath, outputPath); err != nil {
			return fmt.Errorf("zip %s/%s ✗ Error: %w", ctx.SysPlatform, ctx.SysArch, err)
		}
		globls.CL.Greenf("zip %s/%s ✓\n", ctx.SysPlatform, ctx.SysArch)

		// 删除原始文件
		if _, err := os.Stat(outputPath); err == nil {
			if err := os.Remove(outputPath); err != nil {
				return fmt.Errorf("删除编译生成的文件 %s 失败: %w", outputPath, err)
			}
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
func buildBatch(v *verman.VerMan, config *gobConfig) error {
	var wg sync.WaitGroup                                  // 用于同步goroutine
	var printMutex sync.Mutex                              // 用于同步打印输出
	maxConcurrency := runtime.NumCPU()                     // 使用CPU核心数作为默认并发数
	concurrencyChan := make(chan struct{}, maxConcurrency) // 控制并发数量的信号量

	// 获取根环境变量
	rootEnvs := os.Environ()

	// 根环境变量长度
	rootEnvLen := len(rootEnvs)

	// 遍历平台
	for _, platform := range config.Build.Platforms {
		// 遍历架构
		for _, arch := range config.Build.Architectures {
			// 跳过不支持的darwin/386和darwin/arm组合
			if platform == "darwin" && (arch == "386" || arch == "arm") {
				continue
			}

			// 如果开启了仅构建当前平台，则跳过其他平台
			if config.Build.CurrentPlatformOnly {
				if platform != runtime.GOOS || arch != runtime.GOARCH {
					printMutex.Lock()
					globls.CL.PrintOkf("跳过非当前平台: %s/%s\n", platform, arch)
					printMutex.Unlock()
					continue
				}
			}

			// 增加等待组计数
			wg.Add(1)

			// 获取并发信号量
			concurrencyChan <- struct{}{}

			// 启动goroutine执行并行构建
			go func(p, a string) {
				defer func() {
					wg.Done()         // 完成后减少等待组计数
					<-concurrencyChan // 释放并发信号量
				}()

				defer func() {
					if err := recover(); err != nil {
						fmt.Printf("panic: %v\nstack: %s\n", err, debug.Stack())
					}
				}()

				// 拷贝根环境变量
				envs := make([]string, rootEnvLen)
				copy(envs, rootEnvs)

				// 设置平台和架构
				GOOS := fmt.Sprintf("GOOS=%s", p)
				GOARCH := fmt.Sprintf("GOARCH=%s", a)

				// 添加环境变量
				envs = append(envs, GOOS, GOARCH)

				// 构建上下文
				ctx := &BuildContext{
					VerMan:      v,
					Env:         envs,
					SysPlatform: p,
					SysArch:     a,
					Config:      config,
				}

				// 直接调用构建函数并处理错误
				if buildErr := buildSingle(ctx); buildErr != nil {
					printMutex.Lock()
					globls.CL.PrintError(buildErr)
					printMutex.Unlock()
				}
			}(platform, arch)
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

	// 打印安装信息
	globls.CL.PrintOk("开始安装")

	// 检查安装目录是否存在，不存在则创建
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("创建安装目录失败: %w", err)
	}

	// 构建目标路径
	targetPath := filepath.Join(binDir, filepath.Base(executablePath))

	// 检查目标文件是否已存在
	if _, err := os.Stat(targetPath); err == nil {
		if !c.Install.Force {
			return fmt.Errorf("文件已存在: %s, 使用--%s/-%s强制覆盖", targetPath, forceFlag.LongName(), forceFlag.ShortName())
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
	globls.CL.PrintOkf("已安装至: %s\n", targetPath)

	return nil
}

// loadAndValidateConfig 加载并验证配置文件
// 参数:
// - config: 指向配置结构体的指针，用于存储加载的配置
// - configFilePath: 配置文件的路径
//
// 返回值:
//
//	error: 如果加载或验证过程中出现错误，则返回错误信息
func loadAndValidateConfig(config *gobConfig, configFilePath string) error {
	// 加载配置文件
	loadedConfig, err := loadConfig(configFilePath)
	if err != nil {
		return fmt.Errorf("加载构建文件 %s 失败: %v", configFilePath, err)
	}

	// 将加载的配置复制到传入的config指针
	*config = *loadedConfig

	// 如果启用了安装选项，则处理安装路径
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
//
//	ldflags - 包含占位符的链接器标志字符串
//	v - 包含Git元数据的结构体
//
// 返回值:
//
//	替换后的链接器标志字符串
func replaceGitPlaceholders(ldflags string, v *verman.VerMan) string {
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

// runTests 运行单元测试
//
// 参数:
//   - timeout: 每个命令的超时时间
//
// 返回值:
//   - error: 错误信息
func runTests(timeout time.Duration) error {
	// 清理测试缓存
	globls.CL.Green("清理测试缓存")
	result, err := runCmd(timeout, globls.GoCleanTestCacheCmd.Cmds, os.Environ())
	if err != nil {
		return fmt.Errorf("%s:\n%s\n%w", globls.GoCleanTestCacheCmd.Name, string(result), err)
	}

	// 执行go test命令
	globls.CL.Green("开始执行单元测试")
	result, err = runCmd(timeout, globls.GoTestCmd.Cmds, os.Environ())
	if err != nil {
		return fmt.Errorf("%s:\n%s\n%w", globls.GoTestCmd.Name, string(result), err)
	}
	return nil
}
