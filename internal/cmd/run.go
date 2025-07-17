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

// buildSingle 执行单个平台构建
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

	// 设置参数: 链接器标志
	buildCmds[3] = ldflags

	// 设置参数: 输出路径
	outputPath := filepath.Join(outputDir, genOutputName(v.AppName, simpleNameFlag.Get(), v.GitVersion, sysPlatform, sysArch))
	buildCmds[5] = outputPath

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

	// 在buildSingle函数中添加zip打包逻辑
	if zipFlag.Get() {
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
