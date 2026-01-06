package cmd

import (
	"fmt"
	"os"
	"time"

	"gitee.com/MM-Q/gob/internal/globls"
	"github.com/pelletier/go-toml/v2"
)

// gobConfig 表示gob构建工具的完整配置结构
// 对应gob.toml配置文件的结构
type gobConfig struct {
	Build   BuildConfig       `toml:"build" comment:"构建配置"`
	Install InstallConfig     `toml:"install" comment:"安装配置"`
	Env     map[string]string `toml:"env" comment:"--env, -e: 环境变量配置"` // 默认值为空映射
}

// BuildConfig 表示构建相关的配置项
// 对应gob.toml中的[build]部分
type BuildConfig struct {
	Output   OutputConfig   `toml:"output" comment:"输出配置"`    // 输出配置
	Source   SourceConfig   `toml:"source" comment:"源码配置"`    // 源码配置
	Git      GitConfig      `toml:"git" comment:"Git配置"`      // Git配置
	Compiler CompilerConfig `toml:"compiler" comment:"编译器配置"` // 编译器配置
	Target   TargetConfig   `toml:"target" comment:"目标平台配置"`  // 目标平台配置
	Command  CommandConfig  `toml:"command" comment:"命令配置"`   // 命令配置
	UI       UIConfig       `toml:"ui" comment:"UI配置"`        // UI配置

	TimeoutDuration time.Duration `toml:"-"` // 内部使用的Duration类型，不导出到TOML
}

// OutputConfig 表示输出相关的配置项
// 对应gob.toml中的[build.output]部分
type OutputConfig struct {
	Dir    string `toml:"dir" comment:"--output, -o: 指定输出目录"`                      // 默认值为"output"
	Name   string `toml:"name" comment:"--name, -n: 指定输出文件名"`                      // 默认值为"gob"
	Simple bool   `toml:"simple" comment:"--simple-name, -sn: 使用简单名称（不包含平台和架构信息）"` // 默认值为false
	Zip    bool   `toml:"zip" comment:"--zip, -z: 将输出文件打包为zip"`                    // 默认值为false
}

// SourceConfig 表示源码相关的配置项
// 对应gob.toml中的[build.source]部分
type SourceConfig struct {
	MainFile  string `toml:"main_file" comment:"--main, -m: 指定入口文件"`                 // 默认值为"main.go"
	UseVendor bool   `toml:"use_vendor" comment:"--use-vendor, -uv: 在编译时使用vendor目录"` // 默认值为false
}

// GitConfig 表示Git相关的配置项
// 对应gob.toml中的[build.git]部分
type GitConfig struct {
	Inject  bool   `toml:"inject" comment:"--git, -g: 在编译时注入git信息"`                                                                                                                                       // 默认值为false
	Ldflags string `toml:"ldflags" comment:"指定包含Git信息的链接器标志, 支持占位符: {{AppName}} (应用名称)、{{GitVersion}} (Git版本)、{{GitCommit}} (提交哈希)、{{GitCommitTime}} (提交时间)、{{BuildTime}} (构建时间)、{{GitTreeState}} (树状态)"` // 默认值为globls.DefaultGitLDFlags
}

// CompilerConfig 表示编译器相关的配置项
// 对应gob.toml中的[build.compiler]部分
type CompilerConfig struct {
	EnableCgo bool   `toml:"enable_cgo" comment:"--enable-cgo, -ec: 启用CGO"`             // 默认值为false
	Ldflags   string `toml:"ldflags" comment:"指定链接器标志"`                                 // 默认值为"-s -w"
	Proxy     string `toml:"proxy" comment:"--proxy, -p: 设置Go代理"`                       // 默认值为"https://goproxy.cn,https://goproxy.io,direct"
	SkipCheck bool   `toml:"skip_check" comment:"--skip-check, -sc: 跳过构建前检查"`           // 默认值为false
	Timeout   string `toml:"timeout" comment:"--timeout: 构建超时时间(支持单位: ns/us/ms/s/m/h)"` // 默认值为30s
}

// TargetConfig 表示目标平台相关的配置项
// 对应gob.toml中的[build.target]部分
type TargetConfig struct {
	Batch               bool     `toml:"batch" comment:"--batch, -b: 批量编译模式"`                                    // 默认值为false
	CurrentPlatformOnly bool     `toml:"current_platform_only" comment:"--current-platform-only, -cpo: 仅编译当前平台"` // 默认值为false
	Platforms           []string `toml:"platforms" comment:"支持的目标平台列表，多个平台用逗号分隔"`                                // 默认值为["darwin", "linux", "windows"]
	Architectures       []string `toml:"architectures" comment:"支持的目标架构列表，多个架构用逗号分隔"`                            // 默认值为["amd64", "arm64"]
}

// CommandConfig 表示命令相关的配置项
// 对应gob.toml中的[build.command]部分
type CommandConfig struct {
	Build []string `toml:"build" comment:"编译命令模板，支持占位符: {{ldflags}} (链接器标志)、{{output}} (输出路径)、{{if UseVendor}}-mod=vendor{{end}} (条件包含vendor)、{{mainFile}} (入口文件), 多个命令用逗号分隔"` // 默认值为globls.GoBuildCmd.Cmds
}

// UIConfig 表示UI相关的配置项
// 对应gob.toml中的[build.ui]部分
type UIConfig struct {
	Color bool `toml:"color" comment:"--color, -c: 启用颜色输出"` // 默认值为false
}

// InstallConfig 表示安装相关的配置项
// 对应gob.toml中的[install]部分
type InstallConfig struct {
	Install     bool   `toml:"install" comment:"--install, -i: 安装编译后的二进制文件"`       // 默认值为false
	InstallPath string `toml:"install_path" comment:"--install-path, -ip: 指定安装路径"` // 默认值为"$GOPATH/bin"
	Force       bool   `toml:"force" comment:"--force, -f: 强制安装（覆盖已存在文件）"`         // 默认值为false
}

// loadConfig 从指定路径加载TOML配置文件并解析为Config结构体
//
// 参数:
//   - filePath: TOML配置文件的路径
//
// 返回:
//   - 解析后的Config结构体指针和可能的错误
func loadConfig(filePath string) (*gobConfig, error) {
	// 创建默认配置结构体
	config := getDefaultConfig()

	// 如果文件不存在, 则返回默认配置
	if info, err := os.Stat(filePath); os.IsNotExist(err) {
		return config, nil
	} else if info.IsDir() {
		return nil, fmt.Errorf("file '%s' is a directory", filePath)
	}

	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// 解析TOML内容到配置结构体
	if err := toml.Unmarshal(content, config); err != nil {
		// 提取TOML解析错误的详细位置信息
		if decodeErr, ok := err.(*toml.DecodeError); ok {
			row, col := decodeErr.Position() // 获取行和列信息
			return nil, fmt.Errorf("TOML解析错误 (行 %d, 列 %d): %v", row, col, decodeErr.Error())
		}
		return nil, fmt.Errorf("加载配置文件 %s 失败: %w", filePath, err)
	}

	// 解析timeout标志设置内部使用的timeoutDuration字段
	var parseErr error
	config.Build.TimeoutDuration, parseErr = time.ParseDuration(config.Build.Compiler.Timeout)
	if parseErr != nil {
		return nil, fmt.Errorf("解析timeout标志失败: %w", parseErr)
	}

	return config, nil
}

// applyConfigFlags 将命令行标志的值应用到配置结构体
//
// 参数值:
//   - config: 要应用标志的配置结构体指针
func applyConfigFlags(config *gobConfig) {
	// 将命令行标志的值设置到配置结构体
	// UI配置
	config.Build.UI.Color = colorFlag.Get() // 是否启用颜色输出

	// 输出配置
	config.Build.Output.Zip = zipFlag.Get()           // 是否启用zip打包
	config.Build.Output.Simple = simpleNameFlag.Get() // 是否启用简单名称
	config.Build.Output.Dir = outputFlag.Get()        // 输出目录
	config.Build.Output.Name = nameFlag.Get()         // 输出文件名

	// 源码配置
	config.Build.Source.UseVendor = vendorFlag.Get() // 是否启用vendor模式
	config.Build.Source.MainFile = mainFlag.Get()    // 主入口文件

	// Git配置
	config.Build.Git.Inject = gitFlag.Get() // 是否启用Git信息注入

	// 编译器配置
	config.Build.Compiler.EnableCgo = cgoFlag.Get()            // 是否启用cgo
	config.Build.Compiler.Proxy = proxyFlag.Get()              // 代理
	config.Build.Compiler.SkipCheck = skipCheckFlag.Get()      // 是否跳过构建前检查
	config.Build.Compiler.Timeout = timeoutFlag.Get().String() // 设置超时时间

	// 目标平台配置
	config.Build.Target.Batch = batchFlag.Get()                             // 是否启用批量构建
	config.Build.Target.CurrentPlatformOnly = currentPlatformOnlyFlag.Get() // 是否仅编译当前平台

	// 安装配置
	config.Install.Install = installFlag.Get()         // 是否启用安装
	config.Install.InstallPath = installPathFlag.Get() // 安装路径
	config.Install.Force = forceFlag.Get()             // 是否启用强制操作

	// 环境变量
	config.Env = envFlag.Get() // 环境变量

	// 设置默认值
	config.Build.Target.Platforms = globls.DefaultPlatforms // 设置默认支持的平台
	config.Build.Target.Architectures = globls.DefaultArchs // 设置默认支持的架构
	config.Build.Command.Build = globls.GoBuildCmd.Cmds     // 设置默认的编译命令
	config.Build.Compiler.Ldflags = globls.DefaultLDFlags   // 链接器标志
	config.Build.Git.Ldflags = globls.DefaultGitLDFlags     // Git链接器标志

	config.Build.TimeoutDuration = timeoutFlag.Get() // 设置内部使用的Duration类型

	// 处理添加环境变量
	for k, v := range envFlag.Get() {
		config.Env[k] = v
	}
}

// getDefaultConfig 获取配置的默认值
//
// 返回值:
//   - *gobConfig: 包含所有默认配置值的结构体指针
func getDefaultConfig() *gobConfig {
	// 创建配置结构体
	defaultConfig := &gobConfig{}

	// UI配置
	defaultConfig.Build.UI.Color = colorFlag.GetDefault() // 是否启用颜色输出

	// 输出配置
	defaultConfig.Build.Output.Zip = zipFlag.GetDefault()           // 是否启用zip打包
	defaultConfig.Build.Output.Simple = simpleNameFlag.GetDefault() // 是否启用简单名称
	defaultConfig.Build.Output.Dir = outputFlag.GetDefault()        // 输出目录
	defaultConfig.Build.Output.Name = nameFlag.GetDefault()         // 输出文件名

	// 源码配置
	defaultConfig.Build.Source.UseVendor = vendorFlag.GetDefault() // 是否启用vendor模式
	defaultConfig.Build.Source.MainFile = mainFlag.GetDefault()    // 主入口文件

	// Git配置
	defaultConfig.Build.Git.Inject = gitFlag.GetDefault() // 是否启用Git信息注入

	// 编译器配置
	defaultConfig.Build.Compiler.EnableCgo = cgoFlag.GetDefault()            // 是否启用cgo
	defaultConfig.Build.Compiler.Proxy = proxyFlag.GetDefault()              // 代理
	defaultConfig.Build.Compiler.SkipCheck = skipCheckFlag.GetDefault()      // 是否跳过构建前检查
	defaultConfig.Build.Compiler.Timeout = timeoutFlag.GetDefault().String() // 设置超时时间

	// 目标平台配置
	defaultConfig.Build.Target.Batch = batchFlag.GetDefault()                             // 是否启用批量构建
	defaultConfig.Build.Target.CurrentPlatformOnly = currentPlatformOnlyFlag.GetDefault() // 是否仅编译当前平台

	// 安装配置
	defaultConfig.Install.Install = installFlag.GetDefault()         // 是否启用安装
	defaultConfig.Install.InstallPath = installPathFlag.GetDefault() // 安装路径
	defaultConfig.Install.Force = forceFlag.GetDefault()             // 是否启用强制操作

	// 环境变量
	defaultConfig.Env = envFlag.GetDefault() // 环境变量

	// 设置默认值
	defaultConfig.Build.Target.Platforms = globls.DefaultPlatforms // 设置默认支持的平台
	defaultConfig.Build.Target.Architectures = globls.DefaultArchs // 设置默认支持的架构
	defaultConfig.Build.Command.Build = globls.GoBuildCmd.Cmds     // 设置默认的编译命令
	defaultConfig.Build.Compiler.Ldflags = globls.DefaultLDFlags   // 链接器标志
	defaultConfig.Build.Git.Ldflags = globls.DefaultGitLDFlags     // Git链接器标志

	defaultConfig.Build.TimeoutDuration = timeoutFlag.GetDefault() // 设置内部使用的Duration类型

	// 处理添加环境变量
	for k, v := range envFlag.GetDefault() {
		defaultConfig.Env[k] = v
	}

	// 返回配置结构体
	return defaultConfig
}

// generateDefaultConfig 生成默认的gob.toml配置文件
//
// 参数值:
//   - config: 默认配置结构体指针
func generateDefaultConfig(config *gobConfig) error {
	// 检查gob.toml文件是否已存在
	if _, err := os.Stat(globls.GobBuildFile); err == nil {
		// 如果没启用--force, 则返回错误
		if !forceFlag.Get() {
			return fmt.Errorf("配置文件 %s 已存在，使用 --%s/-%s 强制覆盖", globls.GobBuildFile, forceFlag.LongName(), forceFlag.ShortName())
		}
	}

	// 设置默认安装路径
	config.Install.InstallPath = "$GOPATH/bin"

	// 设置默认的输出路径
	config.Build.Output.Dir = globls.DefaultOutputDir

	// 设置默认的入口文件
	config.Build.Source.MainFile = globls.DefaultMainFile

	// 创建文件
	file, err := os.Create(globls.GobBuildFile)
	if err != nil {
		return fmt.Errorf("创建gob.toml失败: %v", err)
	}
	defer func() { _ = file.Close() }()

	// 使用toml.Marshal序列化配置
	data, err := toml.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化gob.toml失败: %v", err)
	}

	// 写入文件
	// 先写入配置文件注释
	comment := []byte(globls.ConfigFileHeaderComment)
	if _, err := file.Write(comment); err != nil {
		return fmt.Errorf("写入注释失败: %v", err)
	}

	// 再写入配置数据
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("写入gob.toml失败: %v", err)
	}

	// 写入示例的ENV配置
	if _, err := file.Write([]byte(globls.EnvExample)); err != nil {
		return fmt.Errorf("写入示例配置失败: %v", err)
	}

	return nil
}
