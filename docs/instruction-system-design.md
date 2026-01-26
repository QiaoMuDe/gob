# GOB 指令系统设计方案

## 1. 概述

本文档描述了GOB工具的内置指令系统设计方案，旨在统一构建和任务编排功能，提供强大而灵活的流程控制能力。

### 1.1 设计目标

- 统一构建和任务编排的概念模型
- 提供强大的流程控制和条件执行能力
- 保持与系统命令的兼容性，防止命名冲突
- 支持扩展性，便于添加新功能
- 保持向后兼容性

### 1.2 核心思想

通过内置指令系统，将Go构建功能完全整合到任务编排框架中，实现统一的执行模型。所有功能都通过任务配置和指令组合来实现，提供一致的语法和执行体验。

## 2. 指令系统架构

### 2.1 指令格式规范

```
基本格式: @gob-<category>-<action> [参数...]

示例:
@gob-file-exists path/to/file
@gob-dir-mkdir path/to/dir
@gob-if-condition then-command else-command
@gob-go-version
```

### 2.2 指令分类设计

#### 2.2.1 文件操作类 (@gob-file-*)
```
@gob-file-exists <path>        # 检查文件是否存在，返回true/false
@gob-file-copy <src> <dst>     # 复制文件
@gob-file-move <src> <dst>     # 移动文件
@gob-file-remove <path>         # 删除文件
@gob-file-size <path>          # 获取文件大小
@gob-file-read <path>          # 读取文件内容
@gob-file-write <path> <content> # 写入文件内容
```

#### 2.2.2 目录操作类 (@gob-dir-*)
```
@gob-dir-exists <path>         # 检查目录是否存在，返回true/false
@gob-dir-mkdir <path>          # 创建目录（支持递归）
@gob-dir-remove <path>          # 删除目录（支持递归）
@gob-dir-list <path>           # 列出目录内容
@gob-dir-empty <path>          # 检查目录是否为空
```

#### 2.2.3 条件控制类 (@gob-if-*)
```
@gob-if <condition> <then-cmd> [else-cmd]    # 条件执行
@gob-when <condition> <cmd>                  # 条件满足时执行
@gob-unless <condition> <cmd>                # 条件不满足时执行
@gob-switch <var> <case1> <cmd1> <case2> <cmd2> [default-cmd] # 多条件分支
```

#### 2.2.4 Go专用类 (@gob-go-*)
```
@gob-go-version               # 获取Go版本
@gob-go-mod-tidy             # 执行go mod tidy
@gob-go-mod-download         # 执行go mod download
@gob-go-vet                 # 执行go vet
@gob-go-test [args...]        # 执行go test
@gob-go-build [args...]      # 执行go build
@gob-go-install [args...]     # 执行go install
@gob-go-run [args...]        # 执行go run
```

#### 2.2.5 环境变量类 (@gob-env-*)
```
@gob-env-get <name>          # 获取环境变量
@gob-env-set <name> <value>   # 设置环境变量
@gob-env-list               # 列出所有环境变量
@gob-env-exists <name>       # 检查环境变量是否存在
```

#### 2.2.6 变量操作类 (@gob-var-*)
```
@gob-var-get <name>          # 获取变量
@gob-var-set <name> <value>   # 设置变量
@gob-var-append <name> <value> # 追加到变量
@gob-var-prepend <name> <value> # 前置添加到变量
@gob-var-increment <name>     # 变量值递增
@gob-var-decrement <name>     # 变量值递减
```

#### 2.2.7 构建信息类 (@gob-build-*)
```
@gob-build-version           # 获取构建版本
@gob-build-time             # 获取构建时间
@gob-build-os               # 获取目标操作系统
@gob-build-arch             # 获取目标架构
```

#### 2.2.8 Git信息类 (@gob-git-*)
```
@gob-git-commit             # 获取Git提交哈希
@gob-git-branch            # 获取Git分支
@gob-git-tag               # 获取Git标签
@gob-git-status            # 获取Git状态
@gob-git-remote            # 获取Git远程仓库URL
```

#### 2.2.9 系统信息类 (@gob-sys-*)
```
@gob-sys-os                # 获取当前操作系统
@gob-sys-arch              # 获取当前架构
@gob-sys-homedir           # 获取用户主目录
@gob-sys-workdir           # 获取当前工作目录
```

#### 2.2.10 字符串处理类 (@gob-str-*)
```
@gob-str-lower <string>     # 转换为小写
@gob-str-upper <string>     # 转换为大写
@gob-str-trim <string>      # 去除首尾空格
@gob-str-replace <old> <new> <string> # 字符串替换
@gob-str-join <sep> <str1> <str2> ... # 字符串连接
@gob-str-split <sep> <string> # 字符串分割
@gob-str-format <template> <args...> # 字符串格式化
```

#### 2.2.11 数学运算类 (@gob-math-*)
```
@gob-math-add <num1> <num2>  # 加法
@gob-math-sub <num1> <num2>  # 减法
@gob-math-mul <num1> <num2>  # 乘法
@gob-math-div <num1> <num2>  # 除法
@gob-math-mod <num1> <num2>  # 取模
@gob-math-max <num1> <num2>  # 最大值
@gob-math-min <num1> <num2>  # 最小值
```

### 2.3 指令参数解析

#### 2.3.1 基本参数解析
- 支持空格分隔的参数
- 支持引号包裹的参数（包含空格）
- 支持转义字符

#### 2.3.2 特殊参数格式
```
# 变量引用
@gob-var-set name {{global.vars.app_name}}

# 环境变量引用
@gob-env-set GOPATH {{env.GOPATH}}

# 命令输出引用
@gob-var-set commit @gob-git-commit

# 嵌套指令
@gob-if '@gob-file-exists go.mod' '@gob-go-mod-tidy' 'echo No go.mod found'
```

## 3. 指令系统实现架构

### 3.1 核心组件

#### 3.1.1 指令解析器 (InstructionParser)
```go
type InstructionParser struct {
    registry map[string]InstructionHandler
    options  *ParserOptions
}

type ParserOptions struct {
    DebugMode    bool
    StrictMode  bool
    MaxDepth     int
    Timeout     time.Duration
}
```

#### 3.1.2 指令处理器接口 (InstructionHandler)
```go
type InstructionHandler interface {
    Handle(args []string, context *TaskExecutionContext) (string, error)
    GetHelp() string
    GetCategory() string
    GetAction() string
    Validate(args []string) error
}
```

#### 3.1.3 指令注册表 (InstructionRegistry)
```go
type InstructionRegistry struct {
    handlers map[string]InstructionHandler
    categories map[string][]string
    mutex    sync.RWMutex
}
```

#### 3.1.4 任务执行上下文 (TaskExecutionContext)
```go
type TaskExecutionContext struct {
    TaskName     string
    GlobalConfig *GlobalConfig
    TaskConfig   *TaskConfig
    AllTasks     map[string]*TaskConfig
    Envs         []string
    Vars         map[string]string
    WorkDir      string
    Timeout      time.Duration
    Depth        int // 防止无限递归
}
```

### 3.2 指令执行流程

```
1. 指令识别
   - 使用正则表达式识别指令模式: @gob-<category>-<action>
   - 提取类别、操作和参数

2. 指令验证
   - 检查指令是否在白名单中
   - 验证参数数量和格式
   - 检查执行深度限制

3. 指令解析
   - 解析参数（支持引号和转义）
   - 处理变量替换
   - 处理嵌套指令

4. 指令执行
   - 查找并调用对应的处理器
   - 传递执行上下文
   - 捕获执行结果

5. 结果处理
   - 处理返回值
   - 更新上下文状态
   - 错误处理和恢复
```

### 3.3 防冲突策略

#### 3.3.1 命名空间隔离
- 使用`@gob-`前缀明确标识内置指令
- 系统命令保持原样，不进行特殊处理
- 提供转义机制`@@gob-`表示字面值

#### 3.3.2 指令白名单
- 维护一个允许的指令列表
- 支持配置文件自定义白名单
- 提供安全模式限制危险指令

#### 3.3.3 优先级处理
- 指令优先级高于系统命令
- 支持强制执行系统命令的语法
- 提供指令覆盖机制

## 4. 指令使用示例

### 4.1 基本使用示例

```toml
# 检查环境和依赖
[task.check]
desc = "检查环境和依赖"
cmds = [
    "@gob-go-version",
    "@gob-if '@gob-file-exists go.mod' 'echo Go模块检查通过' 'echo 错误: 不存在go.mod文件 && exit 1'",
    "@gob-if '@gob-file-exists {{main_file}}' 'echo 入口文件检查通过' 'echo 错误: 入口文件不存在 && exit 1'"
]

# 构建Go项目
[task.build]
desc = "构建Go项目"
cmds = [
    "@gob-dir-mkdir {{output_dir}}",
    "go build -trimpath -ldflags '{{ldflags}}' -o {{output_dir}}/{{app_name}} {{main_file}}"
]
depends_on = ["check"]

# 打包构建产物
[task.package]
desc = "打包构建产物"
cmds = [
    "@gob-if '{{zip}}' '@gob-file-remove {{output_dir}}/{{app_name}}.zip'",
    "@gob-if '{{zip}}' 'comprx pack {{output_dir}}/{{app_name}}.zip {{output_dir}}/{{app_name}}'",
    "@gob-if '{{zip}}' '@gob-file-remove {{output_dir}}/{{app_name}}'"
]
depends_on = ["build"]
```

### 4.2 高级使用示例

```toml
# 多平台构建
[task.build-all]
desc = "多平台构建"
cmds = [
    "@gob-var-set platforms 'linux,windows,darwin'",
    "@gob-var-set archs 'amd64,arm64'",
    "@gob-for-each platform '@gob-var-split , {{platforms}}' '@gob-for-each arch '@gob-var-split , {{archs}}' 'go build -o {{output_dir}}/{{app_name}}_{{platform}}_{{arch}} .'"
]

# 条件构建
[task.build-release]
desc = "发布版本构建"
cmds = [
    "@gob-var-set version '@gob-git-tag'",
    "@gob-if '@gob-str-eq {{version}} \"\"' 'echo 错误: 没有Git标签 && exit 1' 'echo 构建版本: {{version}}'",
    "@gob-dir-mkdir {{output_dir}}/release",
    "go build -ldflags '-X main.version={{version}}' -o {{output_dir}}/release/{{app_name}} ."
]

# 动态任务执行
[task.dynamic]
desc = "动态任务执行"
cmds = [
    "@gob-var-set task-file '@gob-env-get TASK_FILE'",
    "@gob-if '@gob-file-exists {{task-file}}' 'gob task --run {{task-file}}' 'echo 任务文件不存在: {{task-file}}'"
]
```

## 5. 向后兼容性

### 5.1 命令行接口兼容
- 保持现有命令行接口不变
- 内部将构建请求转换为任务执行
- 提供迁移指南和工具

### 5.2 配置文件兼容
- 支持现有配置文件格式
- 提供自动转换工具
- 逐步废弃旧格式

### 5.3 渐进式迁移
- 支持新旧格式混合使用
- 提供详细的迁移文档
- 设置过渡期和废弃计划

## 6. 扩展性设计

### 6.1 插件指令
- 支持第三方指令插件
- 提供插件开发SDK
- 维护插件市场

### 6.2 自定义指令
- 支持用户自定义指令
- 提供指令开发指南
- 支持指令组合

### 6.3 指令模板
- 支持指令模板定义
- 提供常用模板库
- 支持模板继承

## 7. 安全考虑

### 7.1 指令沙箱
- 限制指令执行权限
- 控制文件系统访问
- 监控资源使用

### 7.2 输入验证
- 严格验证指令参数
- 防止注入攻击
- 限制字符串长度

### 7.3 执行限制
- 设置执行超时
- 限制递归深度
- 控制并发数量

## 8. 实施计划

### 8.1 第一阶段：基础框架
- 实现指令解析器
- 设计指令处理器接口
- 创建指令注册表

### 8.2 第二阶段：核心指令
- 实现文件操作指令
- 实现条件控制指令
- 实现变量操作指令

### 8.3 第三阶段：专用指令
- 实现Go专用指令
- 实现Git信息指令
- 实现系统信息指令

### 8.4 第四阶段：集成测试
- 集成到任务系统
- 实现向后兼容
- 完善错误处理

### 8.5 第五阶段：文档和工具
- 编写用户文档
- 提供迁移工具
- 创建示例和教程

## 9. 总结

指令系统设计方案通过统一的命名空间和分类体系，有效解决了命令冲突问题，同时提供了强大而灵活的功能扩展能力。该设计既保持了向后兼容性，又为未来发展提供了良好的基础，是GOB工具架构演进的最佳选择。

通过将Go构建功能完全整合到任务编排框架中，我们实现了概念模型的统一，简化了代码架构，提高了维护效率，同时为用户提供了更加强大和灵活的构建和任务执行能力。