# 🚀 GOB - Go Build Tool

[![Go Version](https://img.shields.io/badge/Go-1.24+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Gitee](https://img.shields.io/badge/Gitee-gob-red.svg)](https://gitee.com/MM-Q/gob.git)

**GOB** 是一个功能强大的 Golang 项目构建和任务编排工具，旨在简化 Go 应用程序的构建、打包、安装和任务执行流程。它支持跨平台编译、自定义安装路径、Git 元数据注入、批量构建以及任务编排等功能，帮助开发者更高效地管理 Go 项目的构建过程。

## 📖 项目地址

🔗 [https://gitee.com/MM-Q/gob.git](https://gitee.com/MM-Q/gob.git)

## ✨ 功能特性

- 🌍 **跨平台构建** - 支持 Windows、Linux 和 macOS 等多个操作系统
- 🏗️ **多架构支持** - 支持 amd64、arm64 等多种硬件架构
- 📁 **配置文件驱动** - 通过 TOML 配置文件管理所有构建参数
- 📦 **任务编排** - 支持任务依赖关系定义和自动执行
- 🏷️ **Git 元数据注入** - 自动从 Git 仓库提取版本信息并注入到二进制文件中
- 📦 **批量构建** - 支持同时为多个平台和架构构建二进制文件
- 🗜️ **ZIP 打包** - 可将构建结果打包为 ZIP 文件以便分发
- ⚙️ **环境变量配置** - 灵活的环境变量设置，支持自定义编译环境
- � **Vendor 支持** - 可使用 vendor 目录进行依赖管理
- 🎨 **颜色输出** - 支持彩色日志输出，提高可读性
- 🚀 **快捷任务** - 通过 `--run` 快捷方式运行预定义的构建任务
- 🔧 **变量替换** - 支持全局变量和任务变量，以及命令执行变量
- 📝 **命令显示** - 可配置是否显示执行的命令，便于调试

## 📋 系统要求

- Go 1.24 或更高版本
- 支持 Windows、macOS、Linux

## 📁 项目结构

```
gob/
├── main.go              # 主入口文件
├── go.mod               # Go 模块文件
├── go.sum               # 依赖校验文件
├── build.py             # Python 构建脚本
├── LICENSE              # 许可证文件
├── README.md            # 项目说明文档
├── gobf/                # 构建配置文件目录
│   ├── dev.toml         # 开发环境配置
│   └── release.toml     # 发布环境配置
├── task.toml            # 任务编排配置文件
├── internal/            # 内部包目录
│   └── cmd/             # 命令行相关代码
└── vendor/              # 依赖包目录
```

## 🛠️ 安装方法

### 方式一：源码安装

```bash
# 克隆仓库
git clone https://gitee.com/MM-Q/gob.git
cd gob

# 构建并安装
python3 build.py -s -ai -f
```

### 方式二：Go Install

```bash
go install gitee.com/MM-Q/gob@latest
```

## 🚀 快速开始

### 初始化项目

```bash
# 初始化 gob 构建配置（生成 gobf/ 目录）
gob --init
```

### 任务编排

```bash
# 初始化任务配置文件
gob task --init

# 列出所有可用任务
gob task --list

# 运行指定任务
gob task --run deploy

# 使用自定义任务文件
gob task --run deploy --file custom.toml
```

### 基本构建

```bash
# 使用默认配置文件（gob.toml）构建
gob

# 使用指定的配置文件构建
gob gobf/dev.toml

# 使用快捷方式运行构建任务
gob --run dev
gob --run release
```

### 查看可用任务

```bash
# 列出所有可用的构建任务
gob --list
```

### 生成默认配置文件

```bash
# 生成默认配置文件（gob.toml）
gob --generate-config
```

## 📚 命令行参数

### 全局参数

| 参数 | 缩写 | 描述 |
|------|------|------|
| `--init` | `-i` | 初始化gob构建文件（生成 gobf/ 目录） |
| `--name` | `-n` | 指定生成的项目名称 |
| `--main` | `-m` | 指定入口文件，默认为main.go |
| `--generate-config` | `-gcf` | 生成默认配置文件（gob.toml） |
| `--force` | `-f` | 强制操作（覆盖已存在文件） |
| `--list` | `-l` | 列出可用的构建任务 |
| `--run` | | 运行指定的构建任务（自动在 gobf/ 目录查找） |

### 任务子命令

| 命令 | 描述 |
|------|------|
| `task` | 任务编排工具 |

#### 任务子命令参数

| 参数 | 缩写 | 描述 |
|------|------|------|
| `--init` | `-i` | 初始化任务配置文件 |
| `--list` | `-l` | 列出所有可用任务 |
| `--run` | `-r` | 运行指定任务 |
| `--file` | | 指定任务配置文件路径 |
| `--force` | `-f` | 强制覆盖已存在文件 |

### 使用说明

**重要：** 所有构建参数必须通过配置文件指定，不再支持命令行参数。

## ⚙️ 配置文件

GOB 使用 TOML 格式的配置文件来管理所有构建参数。配置文件通常位于 `gobf/` 目录下，例如 `gobf/dev.toml`、`gobf/release.toml`。

### 构建配置文件结构

```toml
# 配置文件描述（第一行注释将显示在任务列表中）
# 开发环境构建配置

[build]
# 源代码配置
[build.source]
main_file = "main.go"
use_vendor = false

# 编译器配置
[build.compiler]
ldflags = "-s -w"
enable_cgo = false
proxy = "https://goproxy.cn,direct"

# 输出配置
[build.output]
dir = "output"
name = "gob"
simple_name = false
zip = false

# 目标平台配置
[build.target]
platforms = ["windows", "linux", "darwin"]
architectures = ["amd64", "arm64"]
batch = false
current_platform_only = true

# Git 配置
[build.git]
inject = true
ldflags = "-X 'gitee.com/MM-Q/verman.appName={{AppName}}' -X 'gitee.com/MM-Q/verman.gitVersion={{GitVersion}}'"

# 命令配置
[build.command]
build = ["go", "build", "-trimpath", "-ldflags", "{{ldflags}}", "-o", "{{output}}", "{{if UseVendor}}-mod=vendor{{end}}", "{{mainFile}}"]

# 超时配置（秒）
timeout = 300

# 安装配置
[install]
install = false
install_path = ""
force = false

# UI 配置
[build.ui]
color = true

# 环境变量
[env]
# 自定义环境变量
# KEY = "value"
```

### 任务配置文件结构

```toml
# 全局配置
[global]
# 环境变量
envs = {}
# 任务变量
vars = {}
# 工作目录
work_dir = "."
# 超时时间, 单位秒
timeout = '30s'
# 是否显示输出
show_output = true
# 是否显示执行的命令
show_cmd = false
# 任务执行失败时是否退出程序, true=退出, false=继续执行但打印错误
exit_on_error = true

# 任务定义
[task.clean]
# 任务描述
desc = "清理项目"
# 命令列表
cmds = [
    "echo 清理临时文件...",
    "echo 清理完成"
]

[task.build]
desc = "构建项目"
cmds = [
    "echo 开始构建...",
    "echo 编译中...",
    "echo 构建完成"
]
depends_on = ["clean"]

[task.test]
desc = "运行测试"
cmds = [
    "echo 运行单元测试...",
    "echo 运行集成测试...",
    "echo 测试通过"
]
depends_on = ["build"]

[task.deploy]
desc = "部署应用"
cmds = [
    "echo 连接服务器...",
    "echo 上传文件...",
    "echo 部署完成"
]
depends_on = ["test"]

[task.run]
desc = "运行完整流程"
cmds = [
    "echo 启动应用...",
    "echo 应用已启动"
]
depends_on = ["clean", "build", "test", "deploy"]
```

### 变量替换

GOB 支持在任务配置中使用变量替换：

#### 全局变量

```toml
[global.vars]
app_name = "myapp"
version = "1.0.0"

[task.build]
cmds = [
    "echo 构建 {{global.vars.app_name}} v{{global.vars.version}}..."
]
```

#### 任务变量

```toml
[task.build]
vars = { build_mode = "debug" }
cmds = [
    "echo 构建模式: {{task.build.vars.build_mode}}"
]
```

#### 命令执行变量

使用 `@` 前缀执行命令并将结果作为变量值：

```toml
[global.vars]
current_time = "@date +%Y-%m-%d"
git_commit = "@git rev-parse --short HEAD"

[task.build]
cmds = [
    "echo 构建时间: {{global.vars.current_time}}",
    "echo Git提交: {{global.vars.git_commit}}"
]
```

### 常用配置示例

#### 1. 开发环境配置（仅当前平台）

```toml
# 开发环境 - 快速构建当前平台
[build.source]
main_file = "main.go"
use_vendor = true

[build.output]
dir = "bin"
name = "myapp-dev"
simple_name = true

[build.target]
current_platform_only = true

[build.git]
inject = true
```

#### 2. 发布环境配置（跨平台批量构建）

```toml
# 发布环境 - 多平台批量构建
[build.source]
main_file = "main.go"

[build.output]
dir = "output/release"
name = "myapp"
zip = true

[build.target]
platforms = ["windows", "linux", "darwin"]
architectures = ["amd64", "arm64"]
batch = true

[build.git]
inject = true
```

#### 3. 安装配置

```toml
# 安装配置 - 构建后自动安装
[build.source]
main_file = "main.go"

[build.output]
dir = "output"
name = "myapp"
simple_name = true

[build.target]
current_platform_only = true

[install]
install = true
install_path = "/usr/local/bin"
force = true
```

### 编译命令模板占位符

GOB 支持在编译命令模板中使用以下占位符，用于动态生成 `go build` 命令：

| 占位符 | 描述 |
|--------|------|
| `{{ldflags}}` | 链接器标志，对应 `--ldflags` 选项 |
| `{{output}}` | 输出路径，对应 `--output` 选项 |
| `{{if UseVendor}}-mod=vendor{{end}}` | 条件包含 `-vendor` 标志，基于 `use_vendor` 配置 |
| `{{mainFile}}` | 入口文件路径，对应 `--main` 选项 |

#### 配置示例

在 `gob.toml` 中自定义构建命令模板：

```toml
[build]
build_command = [
    "go", "build", "-trimpath", 
    "-ldflags", "{{ldflags}}", 
    "-o", "{{output}}", 
    "{{if UseVendor}}-mod=vendor{{end}}", 
    "{{mainFile}}"
]
```

### Git 链接器标志占位符

GOB 支持在链接器标志中使用以下名字字符串占位符，用于注入 Git 元数据和应用信息：

| 占位符 | 描述 |
|--------|------|
| `{{AppName}}` | 应用程序名称 |
| `{{GitVersion}}` | Git 版本标签 |
| `{{GitCommit}}` | Git 提交哈希 |
| `{{GitCommitTime}}` | Git 提交时间 |
| `{{BuildTime}}` | 构建时间 |
| `{{GitTreeState}}` | Git 树状态（clean/dirty） |

#### 自定义 Git 链接器标志

在 `gob.toml` 中自定义 Git 链接器标志：

```toml
[build]
git_ldflags = "-X main.version={{GitVersion}} -X main.commit={{GitCommit}}"
```

#### 默认配置

```bash
"-X 'gitee.com/MM-Q/verman.appName={{AppName}}' \
 -X 'gitee.com/MM-Q/verman.gitVersion={{GitVersion}}' \
 -X 'gitee.com/MM-Q/verman.gitCommit={{GitCommit}}' \
 -X 'gitee.com/MM-Q/verman.gitCommitTime={{GitCommitTime}}' \
 -X 'gitee.com/MM-Q/verman.buildTime={{BuildTime}}' \
 -X 'gitee.com/MM-Q/verman.gitTreeState={{GitTreeState}}' \
 -s -w"
```

## 💡 使用技巧

### 最佳实践

**1. 使用多个配置文件**

为不同的环境创建独立的配置文件：
- `gobf/dev.toml` - 开发环境
- `gobf/test.toml` - 测试环境
- `gobf/release.toml` - 生产环境

**2. 使用任务编排**

创建任务依赖关系，自动化复杂流程：
```toml
[task.build]
desc = "构建应用"
cmds = ["go build -o app ."]

[task.test]
desc = "运行测试"
cmds = ["go test ./..."]
depends_on = ["build"]

[task.deploy]
desc = "部署应用"
cmds = ["echo 部署到服务器"]
depends_on = ["test"]
```

**4. 使用变量**

利用变量减少重复配置：
```toml
[global.vars]
app_name = "myapp"
build_dir = "dist"

[task.build]
cmds = ["go build -o {{global.vars.build_dir}}/{{global.vars.app_name}} ."]
```

**5. 使用命令显示**

启用命令显示功能，便于调试：
```toml
[global]
show_cmd = true
```

**6. 使用快捷方式**

```bash
# 列出所有可用任务
gob task --list

# 使用快捷方式运行任务
gob task --run deploy
```

**5. 配置文件描述**

在配置文件的第一行添加注释作为描述：
```toml
# 开发环境 - 快速构建当前平台
```

这样在运行 `gob task --list` 时会显示该描述。

**6. 批量构建和安装**

批量构建和安装选项不能同时使用。如果需要构建并安装，请先构建当前平台，再单独安装。

### 环境变量设置

虽然大部分配置通过配置文件管理，但以下环境变量仍然有用：

```bash
# 设置 Go 代理
export GOPROXY=https://goproxy.cn,direct

# 设置私有模块
export GOPRIVATE=gitee.com/your-org/*

# 跨平台编译（全局设置，配置文件中的设置会覆盖此设置）
export GOOS=linux GOARCH=amd64
```

## 🔧 故障排除

### 常见问题

**Q: 构建失败，提示找不到 Go 命令**
```bash
# 确保 Go 已正确安装并在 PATH 中
go version
```

**Q: 配置文件不存在**
```bash
# 初始化 gob 构建配置
gob --init

# 初始化任务配置
gob task --init

# 或生成默认配置文件
gob --generate-config
```

**Q: 跨平台构建失败**
```bash
# 检查目标平台是否支持
go tool dist list

# 确保在配置文件中正确设置平台和架构
```

**Q: Git 信息注入失败**
```bash
# 确保在 Git 仓库中执行
git status

# 检查配置文件中的 [build.git] 设置
```

**Q: 权限不足无法安装**
```bash
# 在配置文件的 [install] 部分设置自定义安装路径
# 例如：install_path = "~/bin"
```

**Q: 任务执行失败**
```bash
# 检查任务配置文件语法
gob task --list

# 查看任务依赖关系
gob task --run <task-name>
```

## 🤝 贡献

欢迎提交 Issue 和 Pull Request 来改进这个项目！

### 开发环境设置

```bash
# 克隆项目
git clone https://gitee.com/MM-Q/gob.git
cd gob

# 安装依赖
go mod tidy

# 运行测试
go test ./...

# 构建项目
go build -o gob main.go
```

## 📄 许可证

本项目采用 MIT 许可证 - 详情请参见 [LICENSE](LICENSE) 文件

## 👨‍💻 作者

**M乔木** - *项目维护者*

- Gitee: [@MM-Q](https://gitee.com/MM-Q)

## 🙏 致谢

感谢所有为这个项目做出贡献的开发者！

---

<div align="center">

**如果这个项目对你有帮助，请给它一个 ⭐️**

</div>