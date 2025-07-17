package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"gitee.com/MM-Q/gob/internal/globls"
	"gitee.com/MM-Q/verman"
)

// runCmd 执行指定系统命令，仅使用指定的环境变量
//
// 参数：
//   - args: 命令行参数切片，args[0] 为命令本身
//   - env: 完整的环境变量切片，形如 "KEY=VALUE"；传 nil 或空切片表示不额外设置
//
// 返回：
//   - result: 标准输出与标准错误合并后的内容
//   - err: 命令执行期间的任何错误
func runCmd(args []string, env []string) ([]byte, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("empty command")
	}
	cmd := exec.Command(args[0], args[1:]...)
	if len(env) > 0 {
		cmd.Env = env // 直接覆盖，不再继承系统环境
	}
	return cmd.CombinedOutput()
}

// genOutputName 生成输出文件名
//
// 参数：
//   - appName: 应用名
//   - useSimpleName: 是否使用简单的文件名
//   - version: 版本号
//
// 返回：
//   - 生成的输出文件名
//
// 注意：
//   - 简单模式：示例, `myapp`
//   - 完整模式：示例, `myapp_linux_amd64_1.0.0`
func genOutputName(appName string, useSimpleName bool, version string) string {
	// 获取系统平台
	sysPlatform := runtime.GOOS

	// 获取系统架构
	sysArch := runtime.GOARCH

	// 简单模式，不添加平台和版本信息
	if useSimpleName {
		switch sysPlatform {
		case "windows":
			return fmt.Sprint(strings.TrimSuffix(appName, ".exe"), ".exe")
		case "darwin":
			return fmt.Sprint(strings.TrimSuffix(appName, ".app"), ".app")
		default:
			return appName
		}
	}

	// 完整模式，添加平台和版本信息
	switch sysPlatform {
	case "windows":
		return fmt.Sprintf("%s_%s_%s_%s.exe", appName, sysPlatform, sysArch, version)
	case "darwin":
		return fmt.Sprintf("%s_%s_%s_%s.app", appName, sysPlatform, sysArch, version)
	default:
		return fmt.Sprintf("%s_%s_%s_%s", appName, sysPlatform, sysArch, version)
	}
}

// checkBaseEnv 检查基础环境以及格式化和静态检查
func checkBaseEnv(v *verman.VerMan) error {
	// 检查go环境
	if _, err := runCmd([]string{"go", "version"}, os.Environ()); err != nil {
		return fmt.Errorf("未找到go环境, 请先安装go环境: %w", err)
	}

	// 检查当前目录下是否存在go.mod
	if _, statErr := os.Stat("go.mod"); os.IsNotExist(statErr) {
		return fmt.Errorf("当前目录下不存在go.mod文件, 请先初始化go.mod文件, 或前往项目根目录执行: %w", statErr)
	}

	// 检查指定的入口文件是否存在
	if _, statErr := os.Stat(mainFlag.Get()); os.IsNotExist(statErr) {
		return fmt.Errorf("入口文件不存在: %w", statErr)
	}

	// 如果启用vendor模式，检查vendor目录是否存在
	if vendorFlag.Get() {
		if _, statErr := os.Stat("vendor"); os.IsNotExist(statErr) {
			return fmt.Errorf("当前路径下不存在vendor目录, 请先执行 go mod vendor 命令生成vendor目录: %w", statErr)
		}
	}

	// 定义用于判断选择检查模式的变量
	var checkMode bool

	// 检查系统中是否存在golangci-lint否则执行默认的处理命令
	if _, err := runCmd([]string{"golangci-lint", "version"}, os.Environ()); err != nil {
		checkMode = true
	}

	// 根据checkMode的值执行不同的处理命令
	var cmds []globls.CommandGroup
	if checkMode {
		cmds = append(cmds, globls.DefaultCheckCmds...)
	} else {
		cmds = append(cmds, globls.GolangciLintCheckCmds...)
	}

	// 获取环境变量
	env := os.Environ()

	// 设置Go代理
	env = append(env, fmt.Sprintf("GOPROXY=%s", proxyFlag.Get()))

	// 检查是否启用CGO
	if cgoFlag.Get() {
		env = append(env, "CGO_ENABLED=1")
	} else {
		env = append(env, "CGO_ENABLED=0")
	}

	// 遍历处理命令组
	for _, cmdGroup := range cmds {
		if result, runErr := runCmd(cmdGroup.Cmds, env); runErr != nil {
			return fmt.Errorf("执行 %s 失败: \n%s \n%w", cmdGroup.Cmds, string(result), runErr)
		}
	}

	// 检查输出目录是否存在，不存在则创建
	if _, err := os.Stat(outputFlag.Get()); os.IsNotExist(err) {
		if err := os.MkdirAll(outputFlag.Get(), os.ModePerm); err != nil {
			return fmt.Errorf("创建输出目录失败: %w", err)
		}
	}

	// 在输出目录下检查即将生成的可执行文件是否存在，存在则删除
	if _, err := os.Stat(genOutputName(nameFlag.Get(), simpleNameFlag.Get(), v.GitVersion)); err == nil {
		if err := os.Remove(genOutputName(nameFlag.Get(), simpleNameFlag.Get(), v.GitVersion)); err != nil {
			return fmt.Errorf("删除历史构建的可执行文件失败: %w", err)
		}
	}

	return nil
}

// getGitMetaData 获取git元数据
//
// 参数：
//   - v: verman.VerMan 结构体指针，用于存储获取到的git元数据
//
// 返回值：
//   - error: 错误信息，如果获取成功则返回nil
func getGitMetaData(v *verman.VerMan) error {
	// 定义命令和对应字段的映射
	commands := []struct {
		cmd   globls.CommandGroup
		field *string
	}{
		{globls.GitVersionCmd, &v.GitVersion},
		{globls.GitCommitHashCmd, &v.GitCommit},
		{globls.GitCommitTimeCmd, &v.GitCommitTime},
	}

	// 处理常规git信息
	for _, item := range commands {
		result, err := runCmd(item.cmd.Cmds, os.Environ())
		if err != nil {
			return fmt.Errorf("%s: \n\t%s \n%w", item.cmd.Name, string(result), err)
		}
		// 设置字段值，并去除首尾空格
		*item.field = strings.TrimSpace(string(result))
	}

	// 特殊处理git树状态
	result, err := runCmd(globls.GitTreeStatusCmd.Cmds, os.Environ())
	if err != nil {
		return fmt.Errorf("%s: \n\t%s \n%w", globls.GitTreeStatusCmd.Name, string(result), err)
	}

	// 根据git树状态设置GitTreeState字段
	if strings.TrimSpace(string(result)) == "" {
		v.GitTreeState = "clean"
	} else {
		v.GitTreeState = "dirty"
	}

	// 设置appName字段
	v.AppName = nameFlag.Get()

	return nil
}
