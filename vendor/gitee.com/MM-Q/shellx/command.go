// Package shellx 定义了shell命令执行库的核心数据类型。
// 本文件定义了Command结构体, 封装了exec.Cmd并提供了丰富的命令执行功能。
//
// Command是命令执行对象的核心实现, 支持:
//   - 同步执行: Exec、ExecOutput、ExecStdout、ExecResult
//   - 异步执行: ExecAsync、Wait
//   - 进程控制: Kill、Signal、IsRunning、GetPID
//   - 执行状态管理: IsExecuted（确保命令只执行一次）
//   - 完整的执行结果: Result对象包含输出、错误、时间、退出码等信息
package shellx

import (
	"fmt"
	"os"
	"os/exec"
	"sync/atomic"
	"syscall"
	"time"
)

// Command 命令对象
type Command struct {
	cmd     *exec.Cmd   // 底层的 exec.Cmd 对象
	execOne atomic.Bool // 用于确保每个命令只执行一次
}

// Cmd 获取底层的 exec.Cmd 对象
//
// 返回:
//   - *exec.Cmd: 底层的 exec.Cmd 对象
func (c *Command) Cmd() *exec.Cmd {
	return c.cmd
}

// Exec 执行命令(阻塞)
//
// 返回:
//   - error: 错误信息
func (c *Command) Exec() error {
	if !c.execOne.CompareAndSwap(false, true) {
		return fmt.Errorf("command has already been executed")
	}
	return c.cmd.Run()
}

// ExecOutput 执行命令并返回合并后的输出(阻塞)
//
// 返回:
//   - []byte: 命令输出
//   - error: 错误信息
//
// 注意:
//   - 由于需要捕获默认的stdout和stderr合并输出, 内部已经设置了WithStdout(os.Stdout)和WithStderr(os.Stderr)
func (c *Command) ExecOutput() ([]byte, error) {
	if !c.execOne.CompareAndSwap(false, true) {
		return nil, fmt.Errorf("command has already been executed")
	}
	return c.cmd.CombinedOutput()
}

// ExecStdout 执行命令并返回标准输出(阻塞)
//
// 返回:
//   - []byte: 标准输出
//   - error: 错误信息
func (c *Command) ExecStdout() ([]byte, error) {
	if !c.execOne.CompareAndSwap(false, true) {
		return nil, fmt.Errorf("command has already been executed")
	}
	return c.cmd.Output()
}

// ExecResult 执行命令并返回完整的执行结果(阻塞)
//
// 使用示例:
//
//	result, err := cmd.ExecResult()
//	if err != nil {
//	    // 处理错误情况
//	    log.Printf("Command failed: %v", err)
//	    return
//	}
//	// 处理成功情况
//	fmt.Println(string(result.Output()))
//
// 返回:
//   - *Result: 执行结果对象, 包含输出、时间、退出码等信息
//   - error: 执行过程中的错误信息
func (c *Command) ExecResult() (*Result, error) {
	if !c.execOne.CompareAndSwap(false, true) {
		return nil, fmt.Errorf("command has already been executed")
	}

	// 命令执行开始时间
	startTime := time.Now()

	// 执行命令
	output, err := c.cmd.CombinedOutput()

	// 命令执行结束时间
	endTime := time.Now()

	// 获取命令的退出码
	var exitCode int
	if err != nil {
		exitCode = -1
	}

	// 创建Result对象
	result := &Result{
		startTime: startTime,              // 命令开始执行时间
		endTime:   endTime,                // 命令执行结束时间
		duration:  endTime.Sub(startTime), // 命令执行耗时
		output:    output,                 // 命令输出
		success:   err == nil,             // 命令执行是否成功
		exitCode:  exitCode,               // 命令退出码
	}

	return result, err
}

// ExecAsync 异步执行命令(非阻塞)
//
// 返回:
//   - error: 错误信息
func (c *Command) ExecAsync() error {
	if !c.execOne.CompareAndSwap(false, true) {
		return fmt.Errorf("command has already been executed")
	}
	return c.cmd.Start()
}

// Wait 等待命令执行完成(仅在异步执行时有效)
//
// 返回:
//   - error: 错误信息
func (c *Command) Wait() error {
	return c.cmd.Wait()
}

// Kill 杀死当前命令的进程
//
// 返回:
//   - error: 错误信息
func (c *Command) Kill() error {
	if c.cmd.Process == nil {
		return fmt.Errorf("no process to kill")
	}
	return c.cmd.Process.Kill()
}

// Signal 向当前进程发送信号
//
// 参数:
//   - sig: 信号类型
//
// 返回:
//   - error: 错误信息
func (c *Command) Signal(sig os.Signal) error {
	if c.cmd.Process == nil {
		return fmt.Errorf("no process to signal")
	}
	return c.cmd.Process.Signal(sig)
}

// IsRunning 检查进程是否还在运行
//
// 返回:
//   - bool: 是否在运行
func (c *Command) IsRunning() bool {
	if c.cmd.Process == nil {
		return false
	}

	if c.cmd.ProcessState != nil {
		return false // 进程已结束
	}

	// 尝试发送信号0检查进程是否存在
	err := c.cmd.Process.Signal(syscall.Signal(0))
	return err == nil
}

// GetPID 获取进程ID
//
// 返回:
//   - int: 进程ID, 如果进程不存在返回0
func (c *Command) GetPID() int {
	if c.cmd.Process == nil {
		return 0
	}
	return c.cmd.Process.Pid
}

// IsExecuted 检查命令是否已经执行过
//
// 返回:
//   - bool: 是否已执行
func (c *Command) IsExecuted() bool {
	return c.execOne.Load()
}

// ########################################
// 便捷函数
// ########################################

// ExecStr 执行命令(阻塞)
//
// 参数:
//   - cmdStr: 命令字符串
//
// 返回:
//   - error: 错误信息
func ExecStr(cmdStr string) error {
	return NewCmdStr(cmdStr).WithStdout(os.Stdout).WithStderr(os.Stderr).Build().Exec()
}

// Exec 执行命令(阻塞)
//
// 函数:
//   - name: 命令名
//   - args: 命令参数
//
// 返回:
//   - error: 错误信息
func Exec(name string, args ...string) error {
	return NewCmd(name, args...).WithStdout(os.Stdout).WithStderr(os.Stderr).Build().Exec()
}

// ExecOutputStr 执行命令并返回合并后的输出(阻塞)
//
// 参数:
//   - cmdStr: 命令字符串
//
// 返回:
//   - []byte: 输出
//   - error: 错误信息
//
// 注意:
//   - 由于需要捕获默认的stdout和stderr合并输出, 内部已经设置了WithStdout(os.Stdout)和WithStderr(os.Stderr)
func ExecOutputStr(cmdStr string) ([]byte, error) {
	return NewCmdStr(cmdStr).Build().ExecOutput()
}

// ExecOutput 执行命令并返回合并后的输出(阻塞)
//
// 函数:
//   - name: 命令名
//   - args: 命令参数
//
// 返回:
//   - []byte: 输出
//   - error: 错误信息
//
// 注意:
//   - 由于需要捕获默认的stdout和stderr合并输出, 内部已经设置了WithStdout(os.Stdout)和WithStderr(os.Stderr)
func ExecOutput(name string, args ...string) ([]byte, error) {
	return NewCmd(name, args...).Build().ExecOutput()
}
