package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gitee.com/MM-Q/gob/internal/types"
	"github.com/pelletier/go-toml/v2"
)

// LoadConfig 从指定路径加载TOML配置文件并解析为Config结构体
//
// 参数:
//   - filePath: TOML配置文件的路径
//
// 返回:
//   - 解析后的Config结构体指针和可能的错误
func LoadConfig(filePath string) (*types.GobConfig, error) {
	// 创建默认配置结构体
	config := GetDefaultConfig()

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

// GetDefaultConfig 获取配置的默认值
//
// 返回值:
//   - *gobConfig: 包含所有默认配置值的结构体指针
func GetDefaultConfig() *types.GobConfig {
	// 解析timeout
	timeoutDuration, err := time.ParseDuration("60s")
	if err != nil {
		timeoutDuration = 60 * time.Second
	}

	return &types.GobConfig{
		Build: types.BuildConfig{
			Output: types.OutputConfig{
				Dir:    types.DefaultOutputDir, // 默认输出目录
				Name:   types.DefaultAppName,   // 默认应用名称
				Simple: false,                  // 默认不使用简单模式
				Zip:    false,                  // 默认不压缩输出
			},
			Source: types.SourceConfig{
				MainFile:  types.DefaultMainFile, // 默认入口文件
				UseVendor: false,                 // 默认不使用vendor目录
			},
			Git: types.GitConfig{
				Inject:  false,                   // 默认不注入Git信息
				Ldflags: types.DefaultGitLDFlags, // 默认Git链接器标志
			},
			Compiler: types.CompilerConfig{
				EnableCgo: false,                // 默认不启用CGO
				Ldflags:   types.DefaultLDFlags, // 默认链接器标志
				Proxy:     types.DefaultGoProxy, // 默认Go代理
				SkipCheck: false,                // 默认不跳过Go模块检查
				Timeout:   "60s",                // 默认编译超时时间
			},
			Target: types.TargetConfig{
				Batch:               false,                  // 默认不批量编译
				CurrentPlatformOnly: false,                  // 默认不仅编译当前平台
				Platforms:           types.DefaultPlatforms, // 默认支持的目标平台
				Architectures:       types.DefaultArchs,     // 默认支持的目标架构
			},
			Command: types.CommandConfig{
				Build: types.GoBuildCmd.Cmds, // 默认编译命令模板
			},
			UI: types.UIConfig{
				Color: false, // 默认不启用颜色输出
			},
			WorkDir: ".", // 默认当前目录
			PreBuild: types.PreBuildConfig{
				Enabled:     false,      // 默认不启用构建前命令
				Commands:    []string{}, // 默认空命令列表
				ExitOnError: true,       // 默认遇到错误时退出
			},
			PostBuild: types.PostBuildConfig{
				Enabled:     false,      // 默认不启用构建后命令
				Commands:    []string{}, // 默认空命令列表
				ExitOnError: true,       // 默认遇到错误时退出
			},
			TimeoutDuration: timeoutDuration, // 默认编译超时时间
		},
		Install: types.InstallConfig{
			Install:     false,         // 默认不安装编译后的二进制文件
			InstallPath: "$GOPATH/bin", // 默认安装路径
			Force:       false,         // 默认不强制安装（覆盖已存在文件）
		},
		Env: make(map[string]string), // 默认环境变量
	}
}

// GenerateDefaultConfig 生成默认的gob.toml配置文件
//
// 参数值:
//   - f: 是否强制覆盖已存在的配置文件
//
// 返回值:
//   - error: 错误信息，如果生成成功则返回nil
func GenerateDefaultConfig(f bool) error {
	// 获取默认配置
	config := GetDefaultConfig()

	// 检查gob.toml文件是否已存在
	if _, err := os.Stat(types.GobBuildFile); err == nil {
		// 如果没有启用f, 则返回错误
		if !f {
			return fmt.Errorf("配置文件 %s 已存在，使用 --force/-f 强制覆盖", types.GobBuildFile)
		}
	}

	// 创建文件
	file, err := os.Create(types.GobBuildFile)
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
	comment := []byte(types.ConfigFileHeaderComment)
	if _, err := file.Write(comment); err != nil {
		return fmt.Errorf("写入注释失败: %v", err)
	}

	// 再写入配置数据
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("写入gob.toml失败: %v", err)
	}

	// 写入示例的ENV配置
	if _, err := file.Write([]byte(types.EnvExample)); err != nil {
		return fmt.Errorf("写入示例配置失败: %v", err)
	}

	return nil
}

// FindConfigByPrefix 根据前缀查找配置文件
//
// 参数:
//   - prefix: 前缀字符串（至少两个字符）
//   - configDir: 配置文件目录（默认为 "gobf"）
//
// 返回值:
//   - string: 匹配的配置文件名（不含路径）
//   - error: 错误信息
func FindConfigByPrefix(prefix, configDir string) (string, error) {
	// 检查前缀长度至少为两个字符
	if len(prefix) < 2 {
		return "", fmt.Errorf("配置名称至少需要两个字符")
	}

	// 如果没有指定配置目录，使用默认值
	if configDir == "" {
		configDir = "gobf"
	}

	// 读取配置目录下的所有文件
	entries, err := os.ReadDir(configDir)
	if err != nil {
		return "", fmt.Errorf("读取配置目录失败: %w", err)
	}

	// 查找匹配的配置文件
	var matchedFile string
	prefixLower := strings.ToLower(prefix) // 转换为小写，实现不区分大小写

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// 检查是否为 .toml 文件
		if filepath.Ext(name) != ".toml" {
			continue
		}

		// 去除 .toml 后缀并转换为小写
		baseName := strings.TrimSuffix(name, ".toml")
		baseNameLower := strings.ToLower(baseName)

		// 检查是否以前缀匹配（不区分大小写）
		if strings.HasPrefix(baseNameLower, prefixLower) {
			matchedFile = name
			break
		}
	}

	// 如果没有找到匹配的文件
	if matchedFile == "" {
		return "", fmt.Errorf("没有找到以 '%s' 开头的配置文件", prefix)
	}

	return matchedFile, nil
}
