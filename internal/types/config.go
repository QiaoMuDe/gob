package types

import (
	"time"
)

// GobConfig 表示gob构建工具的完整配置结构
// 对应gob.toml配置文件的结构
type GobConfig struct {
	Build   BuildConfig       `toml:"build" comment:"构建配置"`
	Install InstallConfig     `toml:"install" comment:"安装配置"`
	Env     map[string]string `toml:"env" comment:"环境变量配置"` // 默认值为空映射
}

// BuildConfig 表示构建相关的配置项
// 对应gob.toml中的[build]部分
type BuildConfig struct {
	Output    OutputConfig    `toml:"output" comment:"输出配置"`             // 输出配置
	Source    SourceConfig    `toml:"source" comment:"源码配置"`             // 源码配置
	Git       GitConfig       `toml:"git" comment:"Git配置"`               // Git配置
	Compiler  CompilerConfig  `toml:"compiler" comment:"编译器配置"`          // 编译器配置
	Target    TargetConfig    `toml:"target" comment:"目标平台配置"`           // 目标平台配置
	Command   CommandConfig   `toml:"command" comment:"命令配置"`            // 命令配置
	UI        UIConfig        `toml:"ui" comment:"UI配置"`                 // UI配置
	WorkDir   string          `toml:"work_dir" comment:"构建工作目录，默认为当前目录"` // 构建工作目录
	PreBuild  PreBuildConfig  `toml:"pre_build" comment:"构建前执行配置"`       // 构建前执行配置
	PostBuild PostBuildConfig `toml:"post_build" comment:"构建后执行配置"`      // 构建后执行配置

	TimeoutDuration time.Duration `toml:"-"` // 内部使用的Duration类型，不导出到TOML
}

// OutputConfig 表示输出相关的配置项
// 对应gob.toml中的[build.output]部分
type OutputConfig struct {
	Dir    string `toml:"dir" comment:"输出目录"`                  // 默认值为"output"
	Name   string `toml:"name" comment:"输出文件名"`                // 默认值为"gob"
	Simple bool   `toml:"simple" comment:"使用简单名称（不包含平台和架构信息）"` // 默认值为false
	Zip    bool   `toml:"zip" comment:"将输出文件打包为zip"`           // 默认值为false
}

// SourceConfig 表示源码相关的配置项
// 对应gob.toml中的[build.source]部分
type SourceConfig struct {
	MainFile  string `toml:"main_file" comment:"入口文件"`            // 默认值为"main.go"
	UseVendor bool   `toml:"use_vendor" comment:"在编译时使用vendor目录"` // 默认值为false
}

// GitConfig 表示Git相关的配置项
// 对应gob.toml中的[build.git]部分
type GitConfig struct {
	Inject  bool   `toml:"inject" comment:"在编译时注入git信息"`                                                                                                                                                  // 默认值为false
	Ldflags string `toml:"ldflags" comment:"指定包含Git信息的链接器标志, 支持占位符: {{AppName}} (应用名称)、{{GitVersion}} (Git版本)、{{GitCommit}} (提交哈希)、{{GitCommitTime}} (提交时间)、{{BuildTime}} (构建时间)、{{GitTreeState}} (树状态)"` // 默认值为DefaultGitLDFlags
}

// CompilerConfig 表示编译器相关的配置项
// 对应gob.toml中的[build.compiler]部分
type CompilerConfig struct {
	EnableCgo bool   `toml:"enable_cgo" comment:"启用CGO"`                     // 默认值为false
	Ldflags   string `toml:"ldflags" comment:"指定链接器标志"`                      // 默认值为"-s -w"
	Proxy     string `toml:"proxy" comment:"设置Go代理"`                         // 默认值为"https://goproxy.cn,https://goproxy.io,direct"
	SkipCheck bool   `toml:"skip_check" comment:"跳过构建前检查"`                   // 默认值为false
	Timeout   string `toml:"timeout" comment:"构建超时时间(支持单位: ns/us/ms/s/m/h)"` // 默认值为60s
}

// TargetConfig 表示目标平台相关的配置项
// 对应gob.toml中的[build.target]部分
type TargetConfig struct {
	Batch               bool     `toml:"batch" comment:"批量编译模式"`                      // 默认值为false
	CurrentPlatformOnly bool     `toml:"current_platform_only" comment:"仅编译当前平台"`     // 默认值为false
	Platforms           []string `toml:"platforms" comment:"支持的目标平台列表，多个平台用逗号分隔"`     // 默认值为["darwin", "linux", "windows"]
	Architectures       []string `toml:"architectures" comment:"支持的目标架构列表，多个架构用逗号分隔"` // 默认值为["amd64", "arm64"]
}

// CommandConfig 表示命令相关的配置项
// 对应gob.toml中的[build.command]部分
type CommandConfig struct {
	Build []string `toml:"build" comment:"编译命令模板，支持占位符: {{ldflags}} (链接器标志)、{{output}} (输出路径)、{{if UseVendor}}-mod=vendor{{end}} (条件包含vendor)、{{mainFile}} (入口文件), 多个命令用逗号分隔"` // 默认值为GoBuildCmd.Cmds
}

// UIConfig 表示UI相关的配置项
// 对应gob.toml中的[build.ui]部分
type UIConfig struct {
	Color bool `toml:"color" comment:"启用颜色输出"` // 默认值为false
}

// PreBuildConfig 表示构建前执行的配置项
// 对应gob.toml中的[build.pre_build]部分
type PreBuildConfig struct {
	Enabled     bool     `toml:"enabled" comment:"是否启用构建前命令"`                                   // 是否启用构建前命令
	Commands    []string `toml:"commands" comment:"构建前执行的命令列表"`                                 // 构建前执行的命令列表
	ExitOnError bool     `toml:"exit_on_error" comment:"命令执行失败时是否退出程序，true=退出，false=继续执行但打印错误"` // 错误处理策略
}

// PostBuildConfig 表示构建后执行的配置项
// 对应gob.toml中的[build.post_build]部分
type PostBuildConfig struct {
	Enabled     bool     `toml:"enabled" comment:"是否启用构建后命令"`                                   // 是否启用构建后命令
	Commands    []string `toml:"commands" comment:"构建后执行的命令列表"`                                 // 构建后执行的命令列表
	ExitOnError bool     `toml:"exit_on_error" comment:"命令执行失败时是否退出程序，true=退出，false=继续执行但打印错误"` // 错误处理策略
}

// InstallConfig 表示安装相关的配置项
// 对应gob.toml中的[install]部分
type InstallConfig struct {
	Install     bool   `toml:"install" comment:"安装编译后的二进制文件"` // 默认值为false
	InstallPath string `toml:"install_path" comment:"指定安装路径"` // 默认值为"$GOPATH/bin"
	Force       bool   `toml:"force" comment:"强制安装（覆盖已存在文件）"` // 默认值为false
}
