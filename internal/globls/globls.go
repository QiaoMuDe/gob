package globls

import "gitee.com/MM-Q/colorlib"

// GitMetaData 用于存储Git相关元数据
type GitMetaData struct {
	AppName       string // 应用程序名称
	GitVersion    string // git版本号
	GitCommit     string // git提交哈希值
	GitCommitTime string // git提交时间
	BuildTime     string // 构建时间
	GitTreeState  string // git树状态
}

// DefaultPlatformMap 默认支持的平台
var DefaultPlatformMap = map[string]bool{
	"windows": true,
	"darwin":  true,
	"linux":   true,
}

// DefaultArchMap 默认支持的架构
var DefaultArchMap = map[string]bool{
	"386":    true,
	"amd64":  true,
	"x86_64": true,
	"x64":    true,
	"x86":    true,
	"arm":    true,
	"arm64":  true,
}

var (
	// CL 颜色实例
	CL = colorlib.GetCL()
)

// 默认配置
const (
	// DefaultGoProxy 默认的Go代理
	DefaultGoProxy = "https://goproxy.cn,https://goproxy.io,direct"

	// DefaultOutputDir 默认输出目录
	DefaultOutputDir = "./output"

	// DefaultMainFile 默认入口文件
	DefaultMainFile = "./main.go"

	// DefaultAppName 默认应用程序名称
	DefaultAppName = "myapp"

	// DefaultLDFlags 默认链接器标志
	DefaultLDFlags = "-s -w"

	// DefaultGitLDFlags 默认启用的Git元数据链接器标志
	DefaultGitLDFlags = "-X 'gitee.com/MM-Q/verman.appName={app_name}' -X 'gitee.com/MM-Q/verman.gitVersion={git_version}' -X 'gitee.com/MM-Q/verman.gitCommit={git_commit}' -X 'gitee.com/MM-Q/verman.gitCommitTime={commit_time}' -X 'gitee.com/MM-Q/verman.buildTime={build_time}' -X 'gitee.com/MM-Q/verman.gitTreeState={tree_state}' -s -w"
)
