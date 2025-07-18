package cmd

import (
	"os"

	"gitee.com/MM-Q/gob/internal/globls"
	"github.com/pelletier/go-toml/v2"
)

// Config 表示gob构建工具的完整配置结构
// 对应gob.toml配置文件的结构
type Config struct {
	Build   BuildConfig       `toml:"build"`
	Install InstallConfig     `toml:"install"`
	Env     map[string]string `toml:"env"`
}

// BuildConfig 表示构建相关的配置项
// 对应gob.toml中的[build]部分
type BuildConfig struct {
	OutputDir           string `toml:"output_dir"`            // 输出目录，对应--output标志
	OutputName          string `toml:"output_name"`           // 输出文件名，对应--name标志
	MainFile            string `toml:"main_file"`             // 入口文件路径，对应--main标志
	Ldflags             string `toml:"ldflags"`               // 链接器标志，对应--ldflags标志
	UseVendor           bool   `toml:"use_vendor"`            // 是否使用vendor目录，对应--use-vendor标志
	InjectGitInfo       bool   `toml:"inject_git_info"`       // 是否注入git信息，对应--git标志
	SimpleName          bool   `toml:"simple_name"`           // 是否使用简单名称，对应--simple-name标志
	Proxy               string `toml:"proxy"`                 // Go代理地址，对应--proxy标志
	EnableCgo           bool   `toml:"enable_cgo"`            // 是否启用CGO，对应--enable-cgo标志
	ColorOutput         bool   `toml:"color_output"`          // 是否启用颜色输出，对应--color标志
	BatchMode           bool   `toml:"batch_mode"`            // 是否批量编译模式，对应--batch标志
	CurrentPlatformOnly bool   `toml:"current_platform_only"` // 是否仅编译当前平台，对应--current-platform-only标志
	ZipOutput           bool   `toml:"zip_output"`            // 是否打包为zip文件，对应--zip标志
}

// InstallConfig 表示安装相关的配置项
// 对应gob.toml中的[install]部分
type InstallConfig struct {
	Install     bool   `toml:"install"`      // 是否安装二进制文件，对应--install标志
	InstallPath string `toml:"install_path"` // 安装路径，对应--install-path标志
	Force       bool   `toml:"force"`        // 是否强制安装，对应--force标志
}

// loadConfig 从指定路径加载TOML配置文件并解析为Config结构体
//
// 参数:
//   - filePath: TOML配置文件的路径
//
// 返回:
//   - 解析后的Config结构体指针和可能的错误
func loadConfig(filePath string) (*Config, error) {
	// 创建默认配置
	config := &Config{
		Build: BuildConfig{
			OutputDir:  globls.DefaultOutputDir,
			OutputName: globls.DefaultAppName,
			MainFile:   globls.DefaultMainFile,
			Ldflags:    globls.DefaultLDFlags,
			Proxy:      globls.DefaultGoProxy,
		},
		Install: InstallConfig{
			InstallPath: getDefaultInstallPath(),
		},
		Env: make(map[string]string),
	}

	// 如果文件不存在，返回默认配置
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return config, nil
	}

	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// 解析TOML内容到配置结构体
	if err := toml.Unmarshal(content, config); err != nil {
		return nil, err
	}

	return config, nil
}
