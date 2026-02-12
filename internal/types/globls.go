package types

// 用于打印信息的前缀
const (
	PrintPrefix = "gob:"
)

// GitMetaData 用于存储Git相关元数据
type GitMetaData struct {
	AppName       string // 应用程序名称
	GitVersion    string // git版本号
	GitCommit     string // git提交哈希值
	GitCommitTime string // git提交时间
	BuildTime     string // 构建时间
	GitTreeState  string // git树状态
}

// DefaultPlatforms 默认支持的平台
var DefaultPlatforms = []string{"darwin", "linux", "windows"}

// DefaultArchs 默认支持的架构
var DefaultArchs = []string{"amd64", "arm64"}

// 默认配置
const (
	// DefaultGoProxy 默认的Go代理
	DefaultGoProxy = "https://goproxy.cn,https://goproxy.io,direct"

	// DefaultOutputDir 默认输出目录
	DefaultOutputDir = "output"

	// DefaultMainFile 默认入口文件
	DefaultMainFile = "main.go"

	// DefaultAppName 默认应用程序名称
	DefaultAppName = "myapp"

	// DefaultLDFlags 默认链接器标志
	DefaultLDFlags = "-s -w"

	// DefaultGitLDFlags 默认启用的Git元数据链接器标志
	DefaultGitLDFlags = "-X 'gitee.com/MM-Q/verman.appName={{AppName}}' -X 'gitee.com/MM-Q/verman.gitVersion={{GitVersion}}' -X 'gitee.com/MM-Q/verman.gitCommit={{GitCommit}}' -X 'gitee.com/MM-Q/verman.gitCommitTime={{GitCommitTime}}' -X 'gitee.com/MM-Q/verman.buildTime={{BuildTime}}' -X 'gitee.com/MM-Q/verman.gitTreeState={{GitTreeState}}' -s -w"

	// ConfigFileHeaderComment 配置文件头注释
	ConfigFileHeaderComment = "# gob 构建工具配置文件 \n# 项目地址: https://gitee.com/MM-Q/gob.git\n\n"

	// EnvExample 环境变量示例
	EnvExample = "# 示例:\n# GOOS = \"linux\"\n# GOARCH = \"amd64\"\n# CGO_ENABLED = \"1\"\n"
)

// 定义命令结构体类型
type CommandGroup struct {
	Name string
	Cmds []string
}

// 定义默认执行检查期间的命令切片
var DefaultCheckCmds = []CommandGroup{
	{"go fmt 格式化", []string{"go", "fmt", "./..."}},
	{"go vet 静态检查", []string{"go", "vet", "./..."}},
}

// 获取git版本号的命令
var GitVersionCmd = CommandGroup{
	"获取git版本号",
	[]string{"git", "describe", "--tags", "--always", "--dirty"},
}

// 获取git提交哈希值的命令
var GitCommitHashCmd = CommandGroup{
	"获取git提交哈希值",
	[]string{"git", "rev-parse", "--short", "HEAD"},
}

// 获取git提交时间的命令
var GitCommitTimeCmd = CommandGroup{
	"获取git提交时间",
	[]string{"git", "log", "-1", "--format=%cd", "--date=iso"},
}

// 获取git树状态的命令
var GitTreeStatusCmd = CommandGroup{
	"获取git树状态",
	[]string{"git", "status", "--porcelain"},
}

// 编译命令 - 包含条件编译选项和入口文件
var GoBuildCmd = CommandGroup{
	"编译GO程序",
	[]string{"go", "build", "-trimpath", "-ldflags", "{{ldflags}}", "-o", "{{output}}", "{{if UseVendor}}-mod=vendor{{end}}", "{{mainFile}}"},
}

// git rev-parse --is-inside-work-tree 用于判断当前目录是否为git仓库
var GitIsInsideWorkTreeCmd = CommandGroup{
	"判断当前目录是否为git仓库",
	[]string{"git", "rev-parse", "--is-inside-work-tree"},
}

// 执行清理 go 测试缓存的命令
var GoCleanTestCacheCmd = CommandGroup{
	"清理 go 测试缓存",
	[]string{"go", "clean", "-testcache"},
}

// 执行 go 测试的命令
var GoTestCmd = CommandGroup{
	"执行 go 测试",
	[]string{"go", "test", "-race", "./..."},
}

const (
	// 定义gob.toml配置文件
	GobBuildFile = "gob.toml"
)
