// run.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gitee.com/MM-Q/gob/internal/globls"
	"gitee.com/MM-Q/verman"
)

// Run 运行 gob 构建工具
func Run() {
	// 默认关闭颜色输出
	if !colorFlag.Get() {
		globls.CL.SetNoColor(true)
	}

	defer func() {
		if err := recover(); err != nil {
			// 打印错误信息和堆栈并退出
			buf := make([]byte, 1024)
			for {
				n := runtime.Stack(buf, false)
				if n < len(buf) {
					buf = buf[:n]
					break
				}
				buf = make([]byte, 2*len(buf))
			}
			globls.CL.PrintErrf("panic: %v\nsstack: %s\n", err, buf)
			os.Exit(1)
		}
	}()

	globls.CL.PrintDbg("gob 构建工具")

	// 获取verman对象
	v := verman.Get()

	// 第一阶段：执行检查和准备阶段
	if err := checkBaseEnv(); err != nil {
		globls.CL.PrintErr(err.Error())
		os.Exit(1)
	}

	// 检查批量构建和安装选项是否同时启用
	if batchFlag.Get() && installFlag.Get() {
		globls.CL.PrintErr("错误: 不能同时使用批量构建和安装选项")
		os.Exit(1)
	}

	// 检查安装和zip选项是否同时启用
	if installFlag.Get() && zipFlag.Get() {
		globls.CL.PrintErr("错误: 不能同时使用安装和zip选项")
		os.Exit(1)
	}

	// 第二阶段：根据参数获取git信息
	if gitFlag.Get() {
		if err := getGitMetaData(v); err != nil {
			globls.CL.PrintErr(err.Error())
			os.Exit(1)
		}
	}

	// 第三阶段：设置构建命令参数
	// 获取链接器标志
	var ldflags string
	ldflags = ldflagsFlag.Get()
	if gitFlag.Get() {
		// 添加git信息
		ldflags = fmt.Sprintf(globls.DefaultGitLDFlags, v.AppName, v.GitVersion, v.GitCommit, v.GitCommitTime, v.BuildTime, v.GitTreeState)
	}

	// 获取输出目录
	outputDir := outputFlag.Get()

	// 第四阶段：执行构建
	if batchFlag.Get() {
		// 批量构建
		if err := buildBatch(v, ldflags, outputDir); err != nil {
			globls.CL.PrintErr(err.Error())
			os.Exit(1)
		}
	} else {
		// 单个构建
		if err := buildSingle(v, ldflags, outputDir, os.Environ(), runtime.GOOS, runtime.GOARCH); err != nil {
			globls.CL.PrintErr(err.Error())
			os.Exit(1)
		}
	}
}

// buildSingle 执行单个平台和架构的构建
//
// 参数:
//   - v: verman对象
//   - ldflags: 链接器标志
//   - outputDir: 输出目录
//   - env: 环境变量
//   - sysPlatform: 系统平台
//   - sysArch: 系统架构
//
// 返回值:
//   - error: 错误信息
func buildSingle(v *verman.VerMan, ldflags string, outputDir string, env []string, sysPlatform string, sysArch string) error {
	// 获取构建命令
	buildCmds := globls.GoBuildCmd.Cmds

	// 生成输出路径
	outputPath := filepath.Join(outputDir, genOutputName(v.AppName, simpleNameFlag.Get(), v.GitVersion, sysPlatform, sysArch))

	// 动态替换命令中的占位符
	for i, cmd := range buildCmds {
		switch cmd {
		case "{{ldflags}}": // 替换链接器标志
			buildCmds[i] = ldflags
		case "{{output}}": // 替换输出路径
			buildCmds[i] = outputPath
		}
	}

	// 在输出目录下检查即将生成的可执行文件是否存在，存在则删除
	if _, err := os.Stat(outputPath); err == nil {
		if err := os.Remove(outputPath); err != nil {
			return fmt.Errorf("删除历史构建的可执行文件失败: %w", err)
		}
	}

	// 如果指定了vendor，则添加-vendor标志
	if vendorFlag.Get() {
		buildCmds = append(buildCmds, "-mod=vendor")
	}

	// 添加入口文件
	buildCmds = append(buildCmds, mainFlag.Get())

	// 获取环境变量
	envs := env

	// 如果指定了环境变量，则添加环境变量
	if len(envFlag.Get()) > 0 {
		for k, v := range envFlag.Get() {
			envs = append(envs, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// 获取Go代理
	GOPROXY := fmt.Sprintf("GOPROXY=%s", proxyFlag.Get())

	// 添加Go代理
	envs = append(envs, GOPROXY)

	// 检查是否启用CGO
	if cgoFlag.Get() {
		envs = append(envs, "CGO_ENABLED=1")
	} else {
		envs = append(envs, "CGO_ENABLED=0")
	}

	// 执行构建命令
	if result, buildErr := runCmd(buildCmds, envs); buildErr != nil {
		return fmt.Errorf("执行 %s 失败: \n%s \n%v", globls.GoBuildCmd.Cmds, result, buildErr)
	}

	// 构建成功
	globls.CL.PrintOkf("build %s %s %s success\n", sysPlatform, sysArch, outputDir)

	// 如果启用了安装选项，则执行安装
	if installFlag.Get() {
		if err := installExecutable(outputPath); err != nil {
			return fmt.Errorf("安装失败: %w", err)
		}
	}

	// 在buildSingle函数中添加zip打包逻辑
	if zipFlag.Get() {
		// 检查输出路径是否存在, 不存在则跳过
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			return fmt.Errorf("编译后的可执行文件不存在: %w", err)
		}

		// 处理文件名
		baseName := strings.TrimSuffix(outputPath, filepath.Ext(filepath.Base(outputPath)))
		zipPath := fmt.Sprint(baseName, ".zip")

		// 调用CreateZip函数
		if err := createZip(zipPath, outputPath); err != nil {
			return fmt.Errorf("打包文件失败: %w", err)
		}
		globls.CL.PrintOkf("打包完成: %s\n", zipPath)

		// 删除原始文件
		if err := os.Remove(outputPath); err != nil {
			return fmt.Errorf("删除原始文件失败: %w", err)
		}
	}
	return nil
}

// buildBatch 执行批量构建
//
// 参数:
//   - v: verman对象
//   - ldflags: 链接器标志
//   - outputDir: 输出目录
//
// 返回值:
//   - error: 错误信息
func buildBatch(v *verman.VerMan, ldflags string, outputDir string) error {
	// 遍历平台
	for _, platform := range globls.DefaultPlatforms {
		// 遍历架构
		for _, arch := range globls.DefaultArchs {
			// 跳过不支持的darwin/386和darwin/arm组合
			if platform == "darwin" && (arch == "386" || arch == "arm") {
				continue
			}

			// 如果开启了仅构建当前平台，则跳过其他平台
			if currentPlatformOnlyFlag.Get() {
				if platform != runtime.GOOS || arch != runtime.GOARCH {
					globls.CL.PrintInff("跳过非当前平台: %s %s\n", platform, arch)
					continue
				}
			}

			// 设置环境变量
			envs := os.Environ()

			// 设置平台和架构
			GOOS := fmt.Sprintf("GOOS=%s", platform)
			GOARCH := fmt.Sprintf("GOARCH=%s", arch)

			// 添加环境变量
			envs = append(envs, GOOS, GOARCH)

			// 调用单个构建函数
			if buildErr := buildSingle(v, ldflags, outputDir, envs, platform, arch); buildErr != nil {
				globls.CL.PrintErrf("build %s %s %s error: %s\n", platform, arch, outputDir, buildErr)
				continue
			}
		}
	}
	return nil
}

// installExecutable 将可执行文件安装到指定路径或GOPATH/bin目录
//
// 参数:
//   - executablePath: 要安装的可执行文件路径
//
// 返回值:
//   - error: 错误信息
func installExecutable(executablePath string) error {
	// 确定安装目录
	installPath := installPathFlag.Get()
	var binDir string

	if installPath != "" {
		binDir = installPath
	} else {
		// 获取GOPATH环境变量
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			// 尝试获取用户主目录作为默认GOPATH
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("未设置GOPATH环境变量且无法获取用户主目录: %w", err)
			}
			gopath = filepath.Join(homeDir, "go")
		}
		binDir = filepath.Join(gopath, "bin")
	}

	// 检查可执行文件是否存在
	if _, err := os.Stat(executablePath); os.IsNotExist(err) {
		return fmt.Errorf("可执行文件不存在: %s", executablePath)
	}

	// 检查安装目录是否存在，不存在则创建
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("创建安装目录失败: %w", err)
	}

	// 构建目标路径
	targetPath := filepath.Join(binDir, filepath.Base(executablePath))

	// 检查目标文件是否已存在
	if _, err := os.Stat(targetPath); err == nil {
		if !forceFlag.Get() {
			return fmt.Errorf("文件已存在: %s, 使用--force强制覆盖", targetPath)
		}
		// 强制删除现有文件
		if err := os.Remove(targetPath); err != nil {
			return fmt.Errorf("删除现有文件失败: %w", err)
		}
		globls.CL.PrintInff("已删除现有文件: %s\n", targetPath)
	}

	// 移动文件到目标路径
	if err := os.Rename(executablePath, targetPath); err != nil {
		return fmt.Errorf("移动文件失败: %w", err)
	}

	// 打印安装成功信息
	globls.CL.PrintOkf("已将 %s 安装到 %s\n", executablePath, binDir)

	return nil
}
