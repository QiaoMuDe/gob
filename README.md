# gob

Golang项目构建工具 - 支持自定义安装路径和跨平台构建的Go项目构建工具

## 项目介绍

`gob`是一个功能强大的Golang项目构建工具，旨在简化Go应用程序的构建、打包和安装流程。它支持跨平台编译、自定义安装路径、Git元数据注入以及批量构建等功能，帮助开发者更高效地管理Go项目的构建过程。

## 项目地址

[https://gitee.com/MM-Q/gob.git](https://gitee.com/MM-Q/gob.git)

## 功能特性

- **跨平台构建**：支持Windows、Linux和macOS等多个操作系统
- **多架构支持**：支持amd64、arm64等多种硬件架构
- **自定义安装路径**：可通过命令行标志指定安装路径，优先于GOPATH环境变量
- **Git元数据注入**：自动从Git仓库提取版本信息并注入到二进制文件中
- **批量构建**：支持同时为多个平台和架构构建二进制文件
- **ZIP打包**：可将构建结果打包为ZIP文件以便分发
- **环境变量配置**：灵活的环境变量设置，支持自定义编译环境
- **Vendor支持**：可使用vendor目录进行依赖管理
- **颜色输出**：支持彩色日志输出，提高可读性

## 安装方法

### 源码安装

```bash
# 克隆仓库
git clone https://gitee.com/MM-Q/gob.git
cd gob

# 构建并安装
python3 build.py -s -ai -f
```

## 使用示例

### 基本构建

```bash
# 使用默认设置构建当前项目
gob
```

### 指定输出目录和文件名

```bash
# 将构建结果输出到指定目录并使用自定义名称
gob -o ./bin -n myapp
```

### 跨平台构建

```bash
# 为Linux amd64架构构建
export GOOS=linux GOARCH=amd64
gob
```

### 批量构建

```bash
# 为所有支持的平台和架构构建
gob --batch
```

### 安装到自定义路径

```bash
# 构建并安装到指定路径
gob --install --install-path /usr/local/bin
```

### 注入Git元数据

```bash
# 构建时注入Git版本信息
gob --git
```

### 打包为ZIP文件

```bash
# 构建并打包为ZIP文件
gob --zip
```

## 命令行参数

| 参数 | 缩写 | 描述 |
|------|------|------|
| `--env` | `-e` | 指定环境变量，格式为: key=value |
| `--output` | `-o` | 指定输出目录 |
| `--name` | `-n` | 指定输出文件名 |
| `--main` | `-m` | 指定入口文件 |
| `--use-vendor` | `-uv` | 在编译时使用vendor目录 |
| `--git` | `-g` | 在编译时注入git信息 |
| `--simple-name` | `-sn` | 使用简单名称 |
| `--proxy` | `-p` | 设置go代理 |
| `--cgo` | `-ec` | 启用cgo |
| `--color` | `-c` | 启用颜色输出 |
| `--install` | `-i` | 安装编译后的二进制文件 |
| `--force` | `-f` | 执行强制操作 |
| `--batch` | `-b` | 批量编译 |
| `--current-platform-only` | `-cpo` | 仅编译当前平台 |
| `--zip` | `-z` | 打包输出文件为zip文件 |
| `--install-path` | `-ip` | 指定安装路径，优先于GOPATH环境变量 |
| `--generate-config` | `-gcf` | 生成默认配置文件 |
| `--test` | `-t` | 在构建前运行单元测试 |
| `--timeout` | 无 | 设置编译超时时间 |

## 编译命令模板占位符

`gob`支持在编译命令模板中使用以下占位符，用于动态生成go build命令：

| 占位符 | 描述 |
|--------|------|
| `{{ldflags}}` | 链接器标志，对应--ldflags选项 |
| `{{output}}` | 输出路径，对应--output选项 |
| `{{if UseVendor}}-mod=vendor{{end}}` | 条件包含-vendor标志，基于use_vendor配置 |
| `{{mainFile}}` | 入口文件路径，对应--main选项 |

### 配置示例

在`gob.toml`中自定义构建命令模板：
```toml
[build]
build_command = ["go", "build", "-trimpath", "-ldflags", "{{ldflags}}", "-o", "{{output}}", "{{if UseVendor}}-mod=vendor{{end}}", "{{mainFile}}"]
```

此示例展示了完整的构建命令模板配置，包含所有可用占位符。你可以根据项目需求调整命令参数和占位符组合。

## Git链接器标志占位符

`gob`支持在链接器标志中使用以下命名字符串占位符，用于注入Git元数据和应用信息：

| 占位符 | 描述 |
|--------|------|
| `{{AppName}}` | 应用程序名称 |
| `{{GitVersion}}` | Git版本标签 |
| `{{GitCommit}}` | Git提交哈希 |
| `{{GitCommitTime}}` | Git提交时间 |
| `{{BuildTime}}` | 构建时间 |
| `{{GitTreeState}}` | Git树状态（clean/dirty） |

### 配置示例

在`gob.toml`中自定义Git链接器标志：
```toml
[build]
git_ldflags = "-X main.version={{GitVersion}} -X main.commit={{GitCommit}}"
```

默认Git链接器标志配置：
```go
"-X 'gitee.com/MM-Q/verman.appName={{AppName}}' -X 'gitee.com/MM-Q/verman.gitVersion={{GitVersion}}' -X 'gitee.com/MM-Q/verman.gitCommit={{GitCommit}}' -X 'gitee.com/MM-Q/verman.gitCommitTime={{GitCommitTime}}' -X 'gitee.com/MM-Q/verman.buildTime={{BuildTime}}' -X 'gitee.com/MM-Q/verman.gitTreeState={{GitTreeState}}' -s -w"
```

## 许可证

本项目采用MIT许可证 - 详情请参见LICENSE文件

## 版权信息

Copyright (c) 2025 M乔木
