package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gitee.com/MM-Q/gob/internal/types"
	"gitee.com/MM-Q/shellx"
	"gitee.com/MM-Q/verman"
)

// GenOutputName 生成输出文件名
//
// 参数：
//   - appName: 应用名
//   - useSimpleName: 是否使用简单的文件名
//   - version: 版本号
//   - sysPlatform: 系统平台
//   - sysArch: 系统架构
//   - isBatch: 是否为批量构建模式
//
// 返回：
//   - 生成的输出文件名
//
// 注意：
//   - 简单模式：示例, `myapp`
//   - 完整模式：示例, `myapp_linux_amd64_1.0.0`
func GenOutputName(appName string, useSimpleName bool, version string, sysPlatform string, sysArch string, isBatch bool) string {
	if useSimpleName && isBatch {
		CL.Yellow("使用批量构建时, 简单模式将失效")
	}

	// 简单模式: 不添加平台和版本信息
	if useSimpleName && !isBatch {
		switch sysPlatform {
		case "windows":
			return fmt.Sprintf("%s.exe", strings.TrimSuffix(appName, filepath.Ext(appName)))
		default:
			return appName
		}
	}

	// 完整模式: 添加平台和版本信息
	switch sysPlatform {
	case "windows":
		return fmt.Sprintf("%s_%s_%s_%s.exe", strings.TrimSuffix(appName, filepath.Ext(appName)), sysPlatform, sysArch, version)
	default:
		return fmt.Sprintf("%s_%s_%s_%s", appName, sysPlatform, sysArch, version)
	}
}

// CheckBaseEnv 检查基础环境以及格式化和静态检查
//
// 参数:
//   - config: 配置结构体
//
// 返回值:
//   - error: 错误信息
func CheckBaseEnv(config *types.GobConfig) error {
	// 检查go环境
	if err := shellx.NewCmds([]string{"go", "env"}).WithTimeout(config.Build.TimeoutDuration).Exec(); err != nil {
		return err
	}

	// 检查当前目录下是否存在go.mod
	if _, statErr := os.Stat("go.mod"); os.IsNotExist(statErr) {
		return fmt.Errorf("当前目录下不存在go.mod文件, 请先初始化go.mod文件, 或前往项目根目录执行: %w", statErr)
	}

	// 检查指定的入口文件是否存在
	if _, statErr := os.Stat(config.Build.Source.MainFile); os.IsNotExist(statErr) {
		return fmt.Errorf("入口文件不存在: %w", statErr)
	}

	// 如果启用vendor模式，检查vendor目录是否存在
	if config.Build.Source.UseVendor {
		if _, statErr := os.Stat("vendor"); os.IsNotExist(statErr) {
			return fmt.Errorf("当前路径下不存在vendor目录, 请先执行 go mod vendor 命令生成vendor目录: %w", statErr)
		}
	}

	// 如果启用了跳过检查选项，则跳过代码检查
	if config.Build.Compiler.SkipCheck {
		CL.Yellowf("%s 已启用 'skip_check' 选项，跳过代码检查\n", types.PrintPrefix)
	} else {
		// 定义用于判断选择检查模式的变量
		var checkMode bool

		// 检查系统中是否存在golangci-lint否则执行默认的处理命令
		if err := shellx.NewCmds([]string{"golangci-lint", "version"}).WithTimeout(config.Build.TimeoutDuration).Exec(); err != nil {
			checkMode = true
		}

		// 根据checkMode的值执行不同的处理命令
		var cmds []types.CommandGroup
		if checkMode {
			cmds = append(cmds, types.DefaultCheckCmds...)
		} else {
			cmds = append(cmds, types.GolangciLintCheckCmds...)
		}

		// 设置Go代理(如果配置了代理)
		var envs []string
		if config.Build.Compiler.Proxy != "" {
			envs = append(envs, fmt.Sprintf("GOPROXY=%s", config.Build.Compiler.Proxy))
		}

		// 遍历处理命令组
		for _, cmdGroup := range cmds {
			if result, runErr := shellx.NewCmds(cmdGroup.Cmds).WithTimeout(config.Build.TimeoutDuration).WithEnvs(envs).ExecOutput(); runErr != nil {
				// 如果存在输出，则打印
				if len(result) > 0 {
					return fmt.Errorf("执行 %s 失败: %s%s", cmdGroup.Cmds, string(result), runErr)
				}

				// 如果没有输出，则打印错误
				return fmt.Errorf("执行 %s 失败: %w", cmdGroup.Cmds, runErr)
			}
		}
	}

	// 创建输出目录(如果不存在)
	if err := os.MkdirAll(config.Build.Output.Dir, os.ModePerm); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	return nil
}

// GetGitMetaData 获取git元数据
//
// 参数：
//   - timeout: 每个命令的超时时间
//   - v: verman.Info 结构体指针，用于存储获取到的git元数据
//   - c: gobConfig 结构体指针，用于获取配置信息
//
// 返回值：
//   - error: 错误信息，如果获取成功则返回nil
func GetGitMetaData(timeout time.Duration, v *verman.Info, c *types.GobConfig) error {
	// 检查Git是否安装
	if err := shellx.NewCmds([]string{"git", "--version"}).WithTimeout(timeout).Exec(); err != nil {
		return fmt.Errorf("未检测到Git, 请先安装Git并确保其在PATH中: %w", err)
	}

	// 检查当前目录是否为git仓库
	if result, err := shellx.NewCmds(types.GitIsInsideWorkTreeCmd.Cmds).WithTimeout(timeout).ExecOutput(); err != nil {
		if strings.Contains(string(result), "not a git repository") {
			return fmt.Errorf("当前目录不是Git仓库, 请先执行`git init`初始化仓库: %w", err)
		}
		return fmt.Errorf("检查Git仓库状态失败: %w", err)
	}

	// 定义命令和对应字段的映射
	commands := []struct {
		cmd   types.CommandGroup
		field *string
	}{
		{types.GitVersionCmd, &v.GitVersion},
		{types.GitCommitHashCmd, &v.GitCommit},
		{types.GitCommitTimeCmd, &v.GitCommitTime},
	}

	// 处理常规git信息
	for _, item := range commands {
		cmdResult, runErr := shellx.NewCmds(item.cmd.Cmds).WithTimeout(timeout).ExecOutput()
		if runErr != nil {
			return fmt.Errorf("%s: \n\t%s \n%w", item.cmd.Name, string(cmdResult), runErr)
		}
		// 设置字段值，并去除首尾空格
		*item.field = strings.TrimSpace(string(cmdResult))
	}

	// 特殊处理git树状态
	result, err := shellx.NewCmds(types.GitTreeStatusCmd.Cmds).WithTimeout(timeout).ExecOutput()
	if err != nil {
		return fmt.Errorf("%s: \n\t%s \n%w", types.GitTreeStatusCmd.Name, string(result), err)
	}

	// 根据git树状态设置GitTreeState字段
	if strings.TrimSpace(string(result)) == "" {
		v.GitTreeState = "clean"
	} else {
		v.GitTreeState = "dirty"
	}

	// 设置appName字段
	v.AppName = c.Build.Output.Name

	return nil
}

// GetDefaultInstallPath 返回默认安装路径（多级回退策略）
// 优先级: GOPATH/bin > 用户主目录/go/bin > 当前工作目录/bin
//
// 返回值:
//   - string: 计算得到的默认安装路径（确保返回非空字符串）
func GetDefaultInstallPath() string {
	// 1. 优先使用GOPATH/bin
	if gopath := os.Getenv("GOPATH"); gopath != "" {
		return filepath.Join(gopath, "bin")
	}

	// 2. 尝试获取用户主目录/go/bin
	if homeDir, err := os.UserHomeDir(); err == nil {
		return filepath.Join(homeDir, "go", "bin")
	}

	// 3. 使用当前工作目录/bin（保底策略）
	if currentDir, err := os.Getwd(); err == nil {
		return filepath.Join(currentDir, "bin")
	}

	// 所有获取失败时返回相对路径（理论上不会执行到此处）
	return filepath.Join(".", "bin")
}
