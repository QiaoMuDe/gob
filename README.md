# 🚀 GOB - Go Build Tool

[![Go Version](https://img.shields.io/badge/Go-1.24+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Gitee](https://img.shields.io/badge/Gitee-gob-red.svg)](https://gitee.com/MM-Q/gob.git)

**GOB** 是一个功能强大的 Golang 项目构建工具，旨在简化 Go 应用程序的构建、打包和安装流程。它支持跨平台编译、自定义安装路径、Git 元数据注入以及批量构建等功能，帮助开发者更高效地管理 Go 项目的构建过程。

## 📖 项目地址

🔗 [https://gitee.com/MM-Q/gob.git](https://gitee.com/MM-Q/gob.git)

## ✨ 功能特性

- 🌍 **跨平台构建** - 支持 Windows、Linux 和 macOS 等多个操作系统
- 🏗️ **多架构支持** - 支持 amd64、arm64 等多种硬件架构
- 📁 **自定义安装路径** - 可通过命令行标志指定安装路径，优先于 GOPATH 环境变量
- 🏷️ **Git 元数据注入** - 自动从 Git 仓库提取版本信息并注入到二进制文件中
- 📦 **批量构建** - 支持同时为多个平台和架构构建二进制文件
- 🗜️ **ZIP 打包** - 可将构建结果打包为 ZIP 文件以便分发
- ⚙️ **环境变量配置** - 灵活的环境变量设置，支持自定义编译环境
- 📚 **Vendor 支持** - 可使用 vendor 目录进行依赖管理
- 🎨 **颜色输出** - 支持彩色日志输出，提高可读性

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
├── gobf/                # 配置文件目录
│   ├── dev.toml         # 开发环境配置
│   └── release.toml     # 发布环境配置
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

### 基本构建

```bash
# 使用默认设置构建当前项目
gob
```

### 常用构建选项

```bash
# 指定输出目录和文件名
gob -o ./bin -n myapp

# 跨平台构建（Linux amd64）
export GOOS=linux GOARCH=amd64
gob

# 批量构建所有平台
gob --batch

# 构建并安装到自定义路径
gob --install --install-path /usr/local/bin

# 注入 Git 版本信息
gob --git

# 构建并打包为 ZIP 文件
gob --zip

# 生成默认配置文件
gob --generate-config
```

### 高级用法

```bash
# 使用 vendor 目录构建
gob --use-vendor

# 启用 CGO 构建
gob --cgo

# 设置代理并构建
gob --proxy https://goproxy.cn,direct

# 构建前运行测试
gob --test

# 强制覆盖已存在的文件
gob --force

# 启用彩色输出
gob --color
```

## 📚 命令行参数

### 基础参数

| 参数 | 缩写 | 描述 |
|------|------|------|
| `--output` | `-o` | 指定输出目录 |
| `--name` | `-n` | 指定输出文件名 |
| `--main` | `-m` | 指定入口文件 |
| `--env` | `-e` | 指定环境变量，格式：key=value |

### 构建选项

| 参数 | 缩写 | 描述 |
|------|------|------|
| `--use-vendor` | `-uv` | 在编译时使用 vendor 目录 |
| `--git` | `-g` | 在编译时注入 Git 信息 |
| `--cgo` | `-ec` | 启用 CGO |
| `--proxy` | `-p` | 设置 Go 代理 |
| `--test` | `-t` | 在构建前运行单元测试 |
| `--timeout` | | 设置编译超时时间 |

### 输出选项

| 参数 | 缩写 | 描述 |
|------|------|------|
| `--batch` | `-b` | 批量编译多平台 |
| `--current-platform-only` | `-cpo` | 仅编译当前平台 |
| `--zip` | `-z` | 打包输出文件为 ZIP |
| `--simple-name` | `-sn` | 使用简单名称 |

### 安装选项

| 参数 | 缩写 | 描述 |
|------|------|------|
| `--install` | `-i` | 安装编译后的二进制文件 |
| `--install-path` | `-ip` | 指定安装路径（优先于 GOPATH） |
| `--force` | `-f` | 执行强制操作 |

### 其他选项

| 参数 | 缩写 | 描述 |
|------|------|------|
| `--color` | `-c` | 启用颜色输出 |
| `--generate-config` | `-gcf` | 生成默认配置文件 |

## ⚙️ 配置文件

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

GOB 支持在链接器标志中使用以下命名字符串占位符，用于注入 Git 元数据和应用信息：

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

### 环境变量设置

```bash
# 设置 Go 代理
export GOPROXY=https://goproxy.cn,direct

# 设置私有模块
export GOPRIVATE=gitee.com/your-org/*

# 跨平台编译
export GOOS=linux GOARCH=amd64
```

## 🔧 故障排除

### 常见问题

**Q: 构建失败，提示找不到 Go 命令**
```bash
# 确保 Go 已正确安装并在 PATH 中
go version
```

**Q: 跨平台构建失败**
```bash
# 检查目标平台是否支持
go tool dist list
```

**Q: Git 信息注入失败**
```bash
# 确保在 Git 仓库中执行
git status
```

**Q: 权限不足无法安装**
```bash
# 使用 sudo 或指定用户目录
gob --install --install-path ~/bin
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