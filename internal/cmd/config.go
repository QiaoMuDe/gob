package cmd

import (
	"fmt"
	"os"

	"gitee.com/MM-Q/gob/internal/globls"
	"github.com/pelletier/go-toml/v2"
)

// gobConfig 表示gob构建工具的完整配置结构
// 对应gob.toml配置文件的结构
type gobConfig struct {
	//Title   string            `toml:"title" comment:"gob 构建工具配置文件 - 此文件包含所有可用的构建配置选项，与命令行参数对应"` //
	Build   BuildConfig       `toml:"build"`
	Install InstallConfig     `toml:"install"`
	Env     map[string]string `toml:"env" comment:"--env, -e: 环境变量配置"` // 默认值为空映射
}

// BuildConfig 表示构建相关的配置项
// 对应gob.toml中的[build]部分
type BuildConfig struct {
	OutputDir           string   `toml:"output_dir" comment:"--output, -o: 指定输出目录"`                              // 默认值为"output"
	OutputName          string   `toml:"output_name" comment:"--name, -n: 指定输出文件名"`                              // 默认值为"gob"
	MainFile            string   `toml:"main_file" comment:"--main, -m: 指定入口文件"`                                 // 默认值为"main.go"
	Ldflags             string   `toml:"ldflags" comment:"--ldflags, -l: 指定链接器标志"`                               // 默认值为"-s -w"
	GitLdflags          string   `toml:"git_ldflags" comment:"--git-ldflags, -gl: 指定包含Git信息的链接器标志"`              // 默认值为globls.DefaultGitLDFlags
	UseVendor           bool     `toml:"use_vendor" comment:"--use-vendor, -uv: 在编译时使用vendor目录"`                 // 默认值为false
	InjectGitInfo       bool     `toml:"inject_git_info" comment:"--git, -g: 在编译时注入git信息"`                       // 默认值为false
	SimpleName          bool     `toml:"simple_name" comment:"--simple-name, -sn: 使用简单名称（不包含平台和架构信息）"`           // 默认值为false
	Proxy               string   `toml:"proxy" comment:"--proxy, -p: 设置Go代理"`                                    // 默认值为"https://goproxy.cn,https://goproxy.io,direct"
	EnableCgo           bool     `toml:"enable_cgo" comment:"--enable-cgo, -ec: 启用CGO"`                          // 默认值为false
	ColorOutput         bool     `toml:"color_output" comment:"--color, -c: 启用颜色输出"`                             // 默认值为false
	BatchMode           bool     `toml:"batch_mode" comment:"--batch, -b: 批量编译模式"`                               // 默认值为false
	CurrentPlatformOnly bool     `toml:"current_platform_only" comment:"--current-platform-only, -cpo: 仅编译当前平台"` // 默认值为false
	ZipOutput           bool     `toml:"zip_output" comment:"--zip, -z: 将输出文件打包为zip"`                            // 默认值为false
	Platforms           []string `toml:"platforms" comment:"支持的目标平台列表，多个平台用逗号分隔"`                                // 默认值为["darwin", "linux", "windows"]
	Architectures       []string `toml:"architectures" comment:"支持的目标架构列表，多个架构用逗号分隔"`                            // 默认值为["amd64", "arm64"]
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
		return nil, fmt.Errorf("TOML解析失败: %v", err)
	}

	return config, nil
}

// applyConfigFlags 将命令行标志的值应用到配置结构体
//
// 参数值:
//   - config: 要应用标志的配置结构体指针
func applyConfigFlags(config *gobConfig) {
	// 将命令行标志的值设置到配置结构体
	config.Build.ColorOutput = colorFlag.Get()                       // 是否启用颜色输出
	config.Build.BatchMode = batchFlag.Get()                         // 是否启用批量构建
	config.Build.ZipOutput = zipFlag.Get()                           // 是否启用zip打包
	config.Build.CurrentPlatformOnly = currentPlatformOnlyFlag.Get() // 是否仅编译当前平台
	config.Build.UseVendor = vendorFlag.Get()                        // 是否启用vendor模式
	config.Build.EnableCgo = cgoFlag.Get()                           // 是否启用cgo
	config.Build.InjectGitInfo = gitFlag.Get()                       // 是否启用Git信息注入
	config.Build.SimpleName = simpleNameFlag.Get()                   // 是否启用简单名称
	config.Build.OutputDir = outputFlag.Get()                        // 输出目录
	config.Build.OutputName = nameFlag.Get()                         // 输出文件名
	config.Build.MainFile = mainFlag.Get()                           // 主入口文件
	config.Build.Ldflags = ldflagsFlag.Get()                         // 链接器标志
	config.Build.GitLdflags = gitLdflagsFlag.Get()                   // Git链接器标志
	config.Build.Proxy = proxyFlag.Get()                             // 代理
	config.Install.Install = installFlag.Get()                       // 是否启用安装
	config.Install.InstallPath = installPathFlag.Get()               // 安装路径
	config.Install.Force = forceFlag.Get()                           // 是否启用强制操作
	config.Env = envFlag.Get()                                       // 环境变量
	config.Build.Platforms = globls.DefaultPlatforms                 // 设置默认支持的平台
	config.Build.Architectures = globls.DefaultArchs                 // 设置默认支持的架构

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

	// 将命令行标志的值设置到配置结构体
	defaultConfig.Build.ColorOutput = colorFlag.GetDefault()                       // 是否启用颜色输出
	defaultConfig.Build.BatchMode = batchFlag.GetDefault()                         // 是否启用批量构建
	defaultConfig.Build.ZipOutput = zipFlag.GetDefault()                           // 是否启用zip打包
	defaultConfig.Build.CurrentPlatformOnly = currentPlatformOnlyFlag.GetDefault() // 是否仅编译当前平台
	defaultConfig.Build.UseVendor = vendorFlag.GetDefault()                        // 是否启用vendor模式
	defaultConfig.Build.EnableCgo = cgoFlag.GetDefault()                           // 是否启用cgo
	defaultConfig.Build.InjectGitInfo = gitFlag.GetDefault()                       // 是否启用Git信息注入
	defaultConfig.Build.SimpleName = simpleNameFlag.GetDefault()                   // 是否启用简单名称
	defaultConfig.Build.OutputDir = outputFlag.GetDefault()                        // 输出目录
	defaultConfig.Build.OutputName = nameFlag.GetDefault()                         // 输出文件名
	defaultConfig.Build.MainFile = mainFlag.GetDefault()                           // 主入口文件
	defaultConfig.Build.Ldflags = ldflagsFlag.GetDefault()                         // 链接器标志
	defaultConfig.Build.GitLdflags = gitLdflagsFlag.GetDefault()                   // Git链接器标志
	defaultConfig.Build.Proxy = proxyFlag.GetDefault()                             // 代理
	defaultConfig.Install.Install = installFlag.GetDefault()                       // 是否启用安装
	defaultConfig.Install.InstallPath = installPathFlag.GetDefault()               // 安装路径
	defaultConfig.Install.Force = forceFlag.GetDefault()                           // 是否启用强制操作
	defaultConfig.Env = envFlag.GetDefault()                                       // 环境变量
	defaultConfig.Build.Platforms = globls.DefaultPlatforms                        // 设置默认支持的平台
	defaultConfig.Build.Architectures = globls.DefaultArchs                        // 设置默认支持的架构

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
	config.Build.OutputDir = globls.DefaultOutputDir

	// 设置默认的入口文件
	config.Build.MainFile = globls.DefaultMainFile

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
