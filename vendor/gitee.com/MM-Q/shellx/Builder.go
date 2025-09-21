// Package shellx 定义了shell命令执行库的核心数据类型。
// 本文件定义了Builder结构体和相关构造函数，提供链式调用API来构建命令对象。
//
// Builder是命令构建器的核心实现，支持：
//   - 三种命令创建方式：NewCmd、NewCmds、NewCmdStr
//   - 链式调用设置：工作目录、环境变量、超时、上下文、标准输入输出、Shell类型
//   - 并发安全的读写操作
//   - 灵活的命令配置和构建
package shellx

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"
)

// Builder 命令构建器，提供链式调用
type Builder struct {
	shellType ShellType       // shell类型, 默认根据操作系统自动选择(windows使用cmd, 其他系统使用sh)
	raw       string          // 原始命令字符串
	name      string          // 命令名
	args      []string        // 命令参数
	ctx       context.Context // 上下文
	dir       string          // 工作目录
	env       []string        // 环境变量
	stdin     io.Reader       // 标准输入
	stdout    io.Writer       // 标准输出
	stderr    io.Writer       // 标准错误输出
	timeout   time.Duration   // 超时时间
	mu        sync.RWMutex    // 读写锁，用于保护命令构建器中的字段
}

// ShellType 获取shell类型
//
// 返回:
//   - ShellType: shell类型
func (b *Builder) ShellType() ShellType {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.shellType
}

// Raw 获取原始命令字符串
//
// 返回:
//   - string: 原始命令字符串
func (b *Builder) Raw() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.raw
}

// Name 获取命令名称
//
// 返回:
//   - string: 命令名称
func (b *Builder) Name() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.name
}

// Args 获取命令参数列表
//
// 返回:
//   - []string: 命令参数列表
func (b *Builder) Args() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	tempArgs := make([]string, len(b.args))
	copy(tempArgs, b.args)
	return tempArgs
}

// WorkDir 获取命令执行的工作目录
//
// 返回:
//   - string: 命令执行目录
func (b *Builder) WorkDir() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.dir
}

// Env 获取命令环境变量列表
//
// 返回:
//   - []string: 命令环境变量列表
func (b *Builder) Env() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	tempEnv := make([]string, len(b.env))
	copy(tempEnv, b.env)
	return tempEnv
}

// Timeout 获取命令执行超时时间
//
// 返回:
//   - time.Duration: 命令执行超时时间
func (b *Builder) Timeout() time.Duration {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.timeout
}

// NewCmd 创建新的命令构建器 (数组方式 - 可变参数)
//
// 参数：
//   - name: 命令名
//   - args: 命令参数列表
//
// 返回：
//   - *Builder: 命令构建器对象
func NewCmd(name string, args ...string) *Builder {
	if name == "" {
		panic("name cannot be empty")
	}

	return &Builder{
		name:      name,           // 命令名
		args:      args,           // 命令参数
		env:       os.Environ(),   // 默认继承父进程的环境变量
		shellType: ShellDefault,   // 默认根据操作系统自动选择shell
		mu:        sync.RWMutex{}, // 读写锁
	}
}

// NewCmds 创建新的命令构建器 (数组方式 - 切片参数，第一个元素为命令名)
//
// 参数：
//   - cmdArgs: 命令参数列表，第一个元素为命令名，后续元素为参数
//
// 返回：
//   - *Builder: 命令构建器对象
func NewCmds(cmdArgs []string) *Builder {
	if len(cmdArgs) == 0 {
		return &Builder{}
	}

	name := cmdArgs[0] // 第一个元素为命令名
	args := []string{} // 后续元素为参数
	if len(cmdArgs) > 1 {
		args = cmdArgs[1:]
	}

	return NewCmd(name, args...)
}

// NewCmdStr 创建新的命令构建器 (字符串方式)
//
// 参数：
//   - cmdStr: 命令字符串
//
// 返回：
//   - *Builder: 命令构建器对象
func NewCmdStr(cmdStr string) *Builder {
	// 使用命令解析器解析命令字符串
	cmds := ParseCmd(cmdStr)
	b := NewCmds(cmds)
	// 保存原始命令字符串
	b.raw = cmdStr
	return b
}

// WithWorkDir 设置命令的工作目录
//
// 参数：
//   - dir: 命令的工作目录
//
// 返回：
//   - *Builder: 命令构建器对象
func (b *Builder) WithWorkDir(dir string) *Builder {
	b.mu.Lock()
	defer b.mu.Unlock()

	if dir == "" {
		return b
	}

	b.dir = dir
	return b
}

// WithEnv 设置命令的环境变量
//
// 参数：
//   - key: 环境变量的键
//   - value: 环境变量的值
//
// 返回：
//   - *Builder: 命令构建器对象
//
// 注意:
//   - 该方法会验证key是否为空, 如果为空则忽略。
//   - 无需添加系统环境变量os.Environ(), 系统环境变量会自动继承.
func (b *Builder) WithEnv(key, value string) *Builder {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.env == nil {
		b.env = os.Environ()
	}

	if key == "" {
		return b
	}

	// 只有当key不为空时才添加环境变量
	b.env = append(b.env, fmt.Sprintf("%s=%s", key, value))
	return b
}

// WithEnvs 批量设置命令的环境变量
//
// 参数：
//   - envs: []string类型，环境变量列表，每个元素为"key=value"格式
//
// 返回：
//   - *Builder: 命令构建器对象
//
// 注意:
//   - 该方法会验证环境变量格式，只添加验证通过的环境变量。
//   - 无需添加系统环境变量os.Environ(), 系统环境变量会自动继承.
func (b *Builder) WithEnvs(envs []string) *Builder {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(envs) == 0 {
		return b
	}

	if b.env == nil {
		b.env = os.Environ()
	}

	// 验证环境变量格式，只添加验证通过的环境变量
	validEnvs := make([]string, 0, len(envs))
	for _, env := range envs {
		if err := validateEnvVar(env); err == nil {
			validEnvs = append(validEnvs, env)
		}
	}

	b.env = append(b.env, validEnvs...)
	return b
}

// WithTimeout 设置命令的超时时间
//
// 参数：
//   - timeout: time.Duration类型，命令执行的超时时间
//
// 返回：
//   - *Builder: 命令构建器对象
func (b *Builder) WithTimeout(timeout time.Duration) *Builder {
	b.mu.Lock()
	defer b.mu.Unlock()

	if timeout <= 0 {
		return b
	}

	b.timeout = timeout
	return b
}

// WithContext 设置命令的上下文
//
// 参数：
//   - ctx: context.Context类型，用于取消命令执行和超时控制
//
// 返回：
//   - *Builder: 命令构建器对象
func (b *Builder) WithContext(ctx context.Context) *Builder {
	b.mu.Lock()
	defer b.mu.Unlock()

	if ctx == nil {
		panic("context cannot be nil")
	}
	b.ctx = ctx
	return b
}

// WithStdin 设置命令的标准输入
//
// 参数：
//   - stdin: io.Reader类型，用于提供命令的标准输入
//
// 返回：
//   - *Builder: 命令构建器对象
func (b *Builder) WithStdin(stdin io.Reader) *Builder {
	b.mu.Lock()
	defer b.mu.Unlock()

	if stdin == nil {
		panic("stdin cannot be nil")
	}

	b.stdin = stdin
	return b
}

// WithStdout 设置命令的标准输出
//
// 参数：
//   - stdout: io.Writer类型，用于接收命令的标准输出
//
// 返回：
//   - *Builder: 命令构建器对象
func (b *Builder) WithStdout(stdout io.Writer) *Builder {
	b.mu.Lock()
	defer b.mu.Unlock()

	if stdout == nil {
		panic("stdout cannot be nil")
	}

	b.stdout = stdout
	return b
}

// WithStderr 设置命令的标准错误输出
//
// 参数：
//   - stderr: io.Writer类型，用于接收命令的标准错误输出
//
// 返回：
//   - *Builder: 命令构建器对象
func (b *Builder) WithStderr(stderr io.Writer) *Builder {
	b.mu.Lock()
	defer b.mu.Unlock()

	if stderr == nil {
		panic("stderr cannot be nil")
	}

	b.stderr = stderr
	return b
}

// WithShell 设置命令的shell类型
//
// 参数：
//   - shell: ShellType类型，表示要使用的shell类型
//
// 返回：
//   - *Builder: 命令构建器对象
func (b *Builder) WithShell(shell ShellType) *Builder {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.shellType = shell
	return b
}

// Build 构建并返回命令对象
//
// 返回:
//   - *Command: 构建的命令对象
func (b *Builder) Build() *Command {
	b.mu.Lock()
	defer b.mu.Unlock()

	var cmd *exec.Cmd
	if b.ctx != nil {
		// 使用上下文创建命令对象
		if b.shellType != ShellNone {
			// 使用shell执行命令
			cmd = exec.CommandContext(b.ctx, b.shellType.String(), b.shellType.shellFlags(), getCmdStr(b))
		} else {
			// 不使用shell执行命令
			cmd = exec.CommandContext(b.ctx, b.name, b.args...)
		}

	} else {
		// 创建命令对象
		if b.shellType != ShellNone {
			// 使用shell执行命令
			cmd = exec.Command(b.shellType.String(), b.shellType.shellFlags(), getCmdStr(b))
		} else {
			// 不使用shell执行命令
			cmd = exec.Command(b.name, b.args...)
		}
	}

	// 设置命令属性
	cmd.Dir = b.dir           // 设置工作目录
	cmd.Env = b.env           // 设置环境变量
	cmd.Stdin = b.stdin       // 设置标准输入
	cmd.Stdout = b.stdout     // 设置标准输出
	cmd.Stderr = b.stderr     // 设置错误输出
	cmd.WaitDelay = b.timeout // 设置超时时间

	// 创建包装的命令对象
	c := &Command{cmd: cmd}
	c.execOne.Store(false) // 初始化执行状态为false

	// 返回包装的命令对象
	return c
}
