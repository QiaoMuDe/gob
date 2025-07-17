// run.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

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
	if err := checkBaseEnv(v); err != nil {
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

	// 设置参数: 链接器标志和输出目录
	globls.GoBuildCmd.Cmds[3] = ldflags
	globls.GoBuildCmd.Cmds[5] = filepath.Join(outputDir, genOutputName(nameFlag.Get(), simpleNameFlag.Get(), v.GitVersion))

	// 如果指定了vendor，则添加-vendor标志
	if vendorFlag.Get() {
		globls.GoBuildCmd.Cmds = append(globls.GoBuildCmd.Cmds, "-mod=vendor")
	}

	// 添加入口文件
	globls.GoBuildCmd.Cmds = append(globls.GoBuildCmd.Cmds, mainFlag.Get())

	// 获取环境变量
	envs := os.Environ()

	// 如果指定了环境变量，则添加环境变量
	if len(envFlag.Get()) > 0 {
		for k, v := range envFlag.Get() {
			envs = append(envs, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// 添加Go代理
	envs = append(envs, fmt.Sprintf("GOPROXY=%s", proxyFlag.Get()))

	// 检查是否启用CGO
	if cgoFlag.Get() {
		envs = append(envs, "CGO_ENABLED=1")
	} else {
		envs = append(envs, "CGO_ENABLED=0")
	}

	// 第四阶段：执行构建命令
	if result, buildErr := runCmd(globls.GoBuildCmd.Cmds, envs); buildErr != nil {
		globls.CL.PrintErrf("执行 %s 失败: \n%s \n%v", globls.GoBuildCmd.Cmds, result, buildErr)
		os.Exit(1)
	}
}
