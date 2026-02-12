<div align="center">

# 🚀 QFlag

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8E6?style=flat&logo=go)](https://golang.org/) [![License](https://img.shields.io/badge/License-MIT-green?style=flat)](LICENSE) [![MIT License](https://img.shields.io/badge/License-MIT-green?style=flat)](https://opensource.org/licenses/MIT) [![Gitee](https://img.shields.io/badge/Gitee-qflag-red?style=flat)](https://gitee.com/MM-Q/qflag) [![GitHub](https://img.shields.io/badge/GitHub-qflag-black?style=flat)](https://github.com/QiaoMuDe/qflag) [![Ask DeepWiki](https://deepwiki.com/badge.svg?style=flat)](https://deepwiki.com/QiaoMuDe/qflag)

*泛型设计 • 自动路由 • 类型安全 • 并发安全 • 自动补全 • 子命令管理*

</div>

---

## 📖 项目简介

QFlag 是一个专为 Go 语言设计的命令行参数解析库, 提供了丰富的功能和优雅的 API, 帮助开发者快速构建专业的命令行工具。它支持多种标志类型、子命令、环境变量绑定、自动补全等高级特性, 同时保持简单易用的设计理念。

### ✨ 核心特性

- 🎯 **类型安全** - 支持多种标志类型, 确保类型安全
- 🚀 **高性能** - 优化的解析算法, 快速高效
- 🛡️ **错误处理** - 结构化的错误类型, 便于调试和处理
- 🌍 **国际化** - 支持中文和英文双语
- 🔄 **环境变量** - 自动绑定环境变量
- 📝 **自动补全** - 生成 Bash 和 PowerShell 补全脚本
- 🎨 **帮助生成** - 自动生成专业的帮助文档
- 🔗 **互斥标志** - 支持标志互斥组
- ✅ **必需标志** - 支持标志必需组
- 🌳 **子命令** - 完整的子命令支持

---

## 🔗 项目地址

该项目同时托管在 Gitee 和 GitHub 上，您可以选择合适的平台访问：

| 平台               | 地址                                                        | 描述                       |
| ------------------ | ----------------------------------------------------------- | -------------------------- |
| 🔴**Gitee**  | [gitee.com/MM-Q/qflag](https://gitee.com/MM-Q/qflag)           | 国内访问更快，主要开发仓库 |
| ⚫**GitHub** | [github.com/QiaoMuDe/qflag](https://github.com/QiaoMuDe/qflag) | 国际化平台，同步更新       |

---

## 📦 安装指南

使用 `go get` 命令安装：

```bash
go get -u gitee.com/MM-Q/qflag
```

然后在代码中导入：

```go
import "gitee.com/MM-Q/qflag"
```

---

## 💡 使用示例

### 🚀 全局根命令 (推荐) 

QFlag 提供了全局根命令 `qflag.Root`, 这是最简单、最直接的使用方式。**推荐优先使用**全局根命令作为命令行工具的入口点。

```go
package main

import (
    "fmt"
    "gitee.com/MM-Q/qflag"
)

func main() {
    // 直接使用全局根命令创建标志
    name := qflag.Root.String("name", "n", "用户名", "guest")
    age := qflag.Root.Int("age", "a", "年龄", 18)
    verbose := qflag.Root.Bool("verbose", "v", "详细模式", false)
    
    // 配置全局命令
    qflag.Root.SetDesc("示例应用程序")
    qflag.Root.SetVersion("1.0.0")
    
    // 解析命令行参数
    if err := qflag.Parse(); err != nil {
        fmt.Printf("解析错误: %v\n", err)
        return
    }
    
    // 使用参数
    fmt.Printf("用户名: %s\n", name.Get())
    fmt.Printf("年龄: %d\n", age.Get())
    fmt.Printf("详细模式: %t\n", verbose.Get())
    fmt.Printf("非标志参数: %v\n", qflag.Root.Args())
}
```

#### 全局根命令的优势

- 🎯 **简单易用**: 无需手动创建命令实例, 直接使用
- 🚀 **零配置**: 自动使用可执行文件名作为命令名
- 🔧 **功能完整**: 支持所有 QFlag 的高级功能
- 📦 **统一入口**: 所有操作都通过 `qflag.Root` 访问

#### 全局根命令支持的便捷函数

```go
// 解析函数
qflag.Parse()          // 解析命令行参数
qflag.ParseOnly()       // 仅解析当前命令
qflag.ParseAndRoute()   // 解析并路由到子命令

// 子命令管理
qflag.AddSubCmds(cmd1, cmd2)           // 添加子命令
qflag.AddSubCmdFrom([]Command{cmd1, cmd2}) // 从切片添加子命令

// 互斥组
qflag.AddMutexGroup("format", []string{"json", "xml"}, false)

// 必需组
qflag.AddRequiredGroup("connection", []string{"host", "port"})
```

### 基础用法

```go
package main

import (
    "fmt"
    "os"
    "gitee.com/MM-Q/qflag"
)

func main() {
    // 创建命令
    cmd := qflag.NewCmd("myapp", "m", qflag.ContinueOnError)
    cmd.SetDesc("我的应用程序")
    cmd.SetVersion("1.0.0")

    // 使用便捷方法创建标志
    nameFlag := cmd.String("name", "n", "用户名", "guest")
    verboseFlag := cmd.Bool("verbose", "v", "详细输出", false)

    // 解析参数
    if err := cmd.Parse(os.Args[1:]); err != nil {
        fmt.Printf("参数解析错误: %v\n", err)
        os.Exit(1)
    }

    // 使用参数
    fmt.Printf("用户名: %s\n", nameFlag.GetStr())
    fmt.Printf("详细模式: %v\n", verboseFlag.IsSet())
}
```

### 高级用法

#### 使用全局根命令的子命令支持

```go
package main

import (
    "fmt"
    "gitee.com/MM-Q/qflag"
)

func main() {
    // 配置全局根命令
    qflag.Root.SetDesc("命令行工具")
    qflag.Root.SetVersion("1.0.0")
    
    // 创建全局标志
    verbose := qflag.Root.Bool("verbose", "v", "详细输出", false)
    
    // 创建子命令
    listCmd := qflag.NewCmd("list", "ls", qflag.ContinueOnError)
    listCmd.SetDesc("列出所有项目")
    listCmd.Bool("all", "a", "显示所有项目", false)
    
    addCmd := qflag.NewCmd("add", "a", qflag.ContinueOnError)
    addCmd.SetDesc("添加新项目")
    addCmd.String("name", "n", "项目名称", "")
    
    // 添加子命令到全局根命令
    qflag.AddSubCmds(listCmd, addCmd)
    
    // 解析并路由
    if err := qflag.ParseAndRoute(); err != nil {
        fmt.Printf("错误: %v\n", err)
        return
    }
    
    // 如果是根命令执行
    if qflag.Root.NArg() == 0 {
        fmt.Printf("详细模式: %t\n", verbose.Get())
        fmt.Println("使用 'help' 查看可用命令")
    }
}
```

#### 传统子命令支持

```go
package main

import (
    "fmt"
    "os"
    "gitee.com/MM-Q/qflag"
)

func main() {
    // 创建根命令
    rootCmd := qflag.NewCmd("cli", "", qflag.ContinueOnError)
    rootCmd.SetDesc("命令行工具")

    // 使用便捷方法创建子命令
    listCmd := qflag.NewCmd("list", "ls", qflag.ContinueOnError)
    listCmd.SetDesc("列出所有项目")
    listCmd.Bool("all", "a", "显示所有项目", false)
    
    addCmd := qflag.NewCmd("add", "a", qflag.ContinueOnError)
    addCmd.SetDesc("添加新项目")
    addCmd.String("name", "n", "项目名称", "")
    
    // 添加子命令
    rootCmd.AddSubCmds(listCmd, addCmd)

    // 解析并路由
    if err := rootCmd.ParseAndRoute(os.Args[1:]); err != nil {
        fmt.Printf("错误: %v\n", err)
        os.Exit(1)
    }
}
```

#### 便捷方法创建标志

```go
package main

import (
    "fmt"
    "os"
    "time"
    "gitee.com/MM-Q/qflag"
)

func main() {
    cmd := qflag.NewCmd("server", "s", qflag.ContinueOnError)

    // 使用便捷方法创建多个标志
    cmd.String("host", "h", "主机地址", "localhost")
    cmd.Uint("port", "p", "端口号", 8080)
    cmd.Duration("timeout", "t", "超时时间", time.Second*30)
    cmd.Bool("debug", "d", "调试模式", false)

    fmt.Printf("成功添加 %d 个标志\n", len(cmd.Flags()))
}
```

#### 使用全局根命令的互斥标志组

```go
package main

import (
    "fmt"
    "gitee.com/MM-Q/qflag"
)

func main() {
    // 配置全局根命令
    qflag.Root.SetDesc("格式转换工具")
    
    // 使用全局根命令创建互斥标志
    jsonFlag := qflag.Root.Bool("json", "j", "JSON 格式", false)
    xmlFlag := qflag.Root.Bool("xml", "x", "XML 格式", false)
    yamlFlag := qflag.Root.Bool("yaml", "y", "YAML 格式", false)

    // 添加互斥组到全局根命令
    qflag.AddMutexGroup("format", []string{"json", "xml", "yaml"}, false)

    // 解析参数
    if err := qflag.Parse(); err != nil {
        fmt.Printf("错误: %v\n", err)
        return
    }

    // 使用参数
    if jsonFlag.Get() {
        fmt.Println("使用 JSON 格式")
    } else if xmlFlag.Get() {
        fmt.Println("使用 XML 格式")
    } else if yamlFlag.Get() {
        fmt.Println("使用 YAML 格式")
    }
}
```

#### 传统互斥标志组

```go
package main

import (
    "fmt"
    "os"
    "gitee.com/MM-Q/qflag"
)

func main() {
    cmd := qflag.NewCmd("converter", "", qflag.ContinueOnError)

    // 使用便捷方法创建互斥标志
    jsonFlag := cmd.Bool("json", "j", "JSON 格式", false)
    xmlFlag := cmd.Bool("xml", "x", "XML 格式", false)
    yamlFlag := cmd.Bool("yaml", "y", "YAML 格式", false)

    // 创建互斥标志组
    formatGroup := qflag.NewMutexGroup("format", "输出格式", true)
    formatGroup.AddFlags(jsonFlag, xmlFlag, yamlFlag)
    cmd.AddMutexGroup(formatGroup)

    // 解析参数
    if err := cmd.Parse(os.Args[1:]); err != nil {
        fmt.Printf("错误: %v\n", err)
        return
    }
}
```

#### 使用全局根命令的必需标志组

```go
package main

import (
    "fmt"
    "gitee.com/MM-Q/qflag"
)

func main() {
    // 配置全局根命令
    qflag.Root.SetDesc("数据库连接工具")
    
    // 使用全局根命令创建必需标志
    hostFlag := qflag.Root.String("host", "h", "主机地址", "")
    portFlag := qflag.Root.Uint("port", "p", "端口号", 0)

    // 添加必需组到全局根命令
    qflag.AddRequiredGroup("connection", []string{"host", "port"})

    // 解析参数
    if err := qflag.Parse(); err != nil {
        fmt.Printf("错误: %v\n", err)
        return
    }

    // 使用参数
    fmt.Printf("连接到 %s:%d\n", hostFlag.Get(), portFlag.Get())
}
```

#### 传统必需标志组

```go
package main

import (
    "fmt"
    "os"
    "gitee.com/MM-Q/qflag"
)

func main() {
    cmd := qflag.NewCmd("db-connect", "", qflag.ContinueOnError)

    // 使用便捷方法创建必需标志
    hostFlag := cmd.String("host", "h", "主机地址", "")
    portFlag := cmd.Uint("port", "p", "端口号", 0)
    usernameFlag := cmd.String("username", "u", "用户名", "")
    passwordFlag := cmd.String("password", "P", "密码", "")

    // 添加必需组
    cmd.AddRequiredGroup("connection", []string{"host", "port"})
    cmd.AddRequiredGroup("auth", []string{"username", "password"})

    // 解析参数
    if err := cmd.Parse(os.Args[1:]); err != nil {
        fmt.Printf("错误: %v\n", err)
        return
    }

    // 使用参数
    fmt.Printf("连接到 %s:%d\n", hostFlag.Get(), portFlag.Get())
    fmt.Printf("用户: %s\n", usernameFlag.Get())
}
```

#### 环境变量绑定

```go
package main

import (
    "fmt"
    "os"
    "gitee.com/MM-Q/qflag"
)

func main() {
    cmd := qflag.NewCmd("app", "", qflag.ContinueOnError)
    
    // 设置环境变量前缀
    cmd.SetEnvPrefix("MYAPP")

    // 使用便捷方法创建并绑定环境变量的标志
    dbFlag := cmd.String("database", "d", "数据库地址", "localhost")
    dbFlag.BindEnv("DATABASE_URL")
    
    // 解析参数
    if err := cmd.Parse(os.Args[1:]); err != nil {
        fmt.Printf("错误: %v\n", err)
        return
    }

    fmt.Printf("数据库地址: %s\n", dbFlag.GetStr())
}
```

---

## 📚 API 文档概述

QFlag 提供了简洁而强大的 API, 主要包含以下核心组件: 

### 🚀 全局根命令 (推荐使用方式) 

QFlag 提供了全局根命令 `qflag.Root`, 这是最简单、最直接的使用方式。**推荐优先使用**全局根命令作为命令行工具的入口点。

#### 全局根命令的优势

- 🎯 **简单易用**: 无需手动创建命令实例, 直接使用
- 🚀 **零配置**: 自动使用可执行文件名作为命令名
- 🔧 **功能完整**: 支持所有 QFlag 的高级功能
- 📦 **统一入口**: 所有操作都通过 `qflag.Root` 访问

#### 全局根命令支持的便捷函数

```go
// 解析函数
qflag.Parse()          // 解析命令行参数
qflag.ParseOnly()       // 仅解析当前命令
qflag.ParseAndRoute()   // 解析并路由到子命令

// 子命令管理
qflag.AddSubCmds(cmd1, cmd2)           // 添加子命令
qflag.AddSubCmdFrom([]Command{cmd1, cmd2}) // 从切片添加子命令

// 互斥组
qflag.AddMutexGroup("format", []string{"json", "xml"}, false)
```

#### 全局根命令的使用方式

```go
// 直接使用全局根命令创建标志
name := qflag.Root.String("name", "n", "用户名", "guest")
age := qflag.Root.Int("age", "a", "年龄", 18)
verbose := qflag.Root.Bool("verbose", "v", "详细模式", false)

// 配置全局命令
qflag.Root.SetDesc("示例应用程序")
qflag.Root.SetVersion("1.0.0")

// 解析命令行参数
if err := qflag.Parse(); err != nil {
    fmt.Printf("解析错误: %v\n", err)
    return
}
```

### 核心概念

- **Command (命令) ** - 命令行工具的核心, 支持标志管理、参数解析、子命令等功能
- **Flag (标志) ** - 命令行参数的抽象, 支持多种数据类型
- **MutexGroup (互斥组) ** - 确保组内只有一个标志被设置
- **RequiredGroup (必需组) ** - 确保组内所有标志都被设置

### 便捷方法

Command 类型提供了丰富的便捷方法来创建各种类型的标志, 无需手动创建和添加标志: 

```go
// 字符串和布尔标志
nameFlag := cmd.String("name", "n", "用户名", "guest")
verboseFlag := cmd.Bool("verbose", "v", "详细输出", false)

// 数值类型标志
portFlag := cmd.Uint("port", "p", "端口号", 8080)
timeoutFlag := cmd.Duration("timeout", "t", "超时时间", time.Second*30)

// 集合类型标志
filesFlag := cmd.StringSlice("files", "f", "文件列表", []string{})
tagsFlag := cmd.IntSlice("tags", "", "标签列表", []int{})
```

### 详细的 API 文档

完整的 API 文档和示例代码, 请参考项目中的 `examples/` 目录和源代码中的注释。

---

## 🎯 支持的功能

### 标志功能

- ✅ 短标志名 (单字符, 如 `-v`) 
- ✅ 长标志名 (多字符, 如 `--verbose`) 
- ✅ 默认值设置
- ✅ 必需标志验证
- ✅ 标志描述
- ✅ 环境变量绑定
- ✅ 标志值验证
- ✅ 枚举值限制
- ✅ 切片类型支持

### 命令功能

- ✅ 子命令支持
- ✅ 命令别名
- ✅ 命令描述
- ✅ 版本信息
- ✅ 帮助信息
- ✅ 示例展示
- ✅ 注意事项

### 高级功能

- ✅ 互斥标志组
- ✅ 必需标志组
- ✅ 自动补全脚本生成
- ✅ 环境变量前缀
- ✅ 错误处理策略
- ✅ 中文/英文双语支持
- ✅ Logo 文本设置
- ✅ 使用语法自定义

---

## ⚙️ 配置选项说明

### 错误处理策略

QFlag 提供了三种错误处理策略: 

- **ContinueOnError** - 遇到错误继续解析
- **ExitOnError** - 遇到错误立即退出
- **ReturnOnError** - 遇到错误返回错误

### 命令配置

通过 `CmdConfig` 可以自定义命令的各种行为, 包括版本信息、语言设置、环境变量前缀等。详细的配置选项请参考源代码中的注释。

---

## 📁 项目结构

```
qflag/
├── internal/              # 内部实现
│   ├── builtin/          # 内置标志 (help, version, completion) 
│   ├── cmd/             # 命令实现
│   ├── completion/       # 自动补全脚本生成
│   ├── flag/            # 标志类型实现
│   ├── parser/          # 参数解析器
│   ├── registry/        # 注册表实现
│   ├── testutils/       # 测试工具
│   ├── types/           # 类型定义
│   └── utils/          # 工具函数
├── examples/            # 使用示例
├── exports.go           # 公共 API 导出
├── qflag.go            # 全局根命令和便捷函数
├── go.mod              # Go 模块文件
└── README.md           # 项目文档
```

---

## 🧪 测试说明

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/cmd

# 运行测试并显示覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 测试覆盖

项目使用全面的测试套件, 包括: 
- 单元测试
- 集成测试
- 边界条件测试
- 错误处理测试

---

## 📄 许可证和贡献指南

### 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

### 贡献指南

我们欢迎任何形式的贡献！

#### 贡献方式

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

#### 代码规范

- 遵循 Go 语言代码规范
- 添加适当的注释
- 编写测试用例
- 确保所有测试通过
- 更新相关文档

#### 问题反馈

如果您发现 bug 或有功能建议, 请: 
- 搜索现有的 issues
- 创建新的 issue, 详细描述问题
- 提供复现步骤和环境信息

---

## 📞 联系方式和相关链接

### 相关资源

- 📦 **仓库地址**: [https://gitee.com/MM-Q/qflag.git](https://gitee.com/MM-Q/qflag.git)
- 📖 **文档**: [项目文档](https://gitee.com/MM-Q/qflag)
- 🐛 **问题反馈**: [Issues](https://gitee.com/MM-Q/qflag/issues)
- 💬 **讨论**: [Discussions](https://gitee.com/MM-Q/qflag/discussions)

### 联系方式

- 📧 **邮箱**: [提交 Issue](https://gitee.com/MM-Q/qflag/issues)

---

## 🙏 致谢

感谢所有为本项目做出贡献的开发者！

---

<div align="center">

**[⬆ 返回顶部](#qflag)**

Made with ❤️ by QFlag Team

</div>