package qflag

import (
	"os"
	"path/filepath"
)

// Root 全局根命令实例, 提供对全局标志和子命令的访问
// 用户可以通过 qflag.Root.String() 这样的方式直接创建全局标志
// 这是访问命令行功能的主要入口点, 推荐优先使用
var Root *Cmd

// init 包初始化函数, 直接创建全局根命令实例
func init() {
	// 使用一致的命令名生成逻辑
	cmdName := "myapp"
	if len(os.Args) > 0 {
		cmdName = filepath.Base(os.Args[0])
	}

	// 直接创建全局根命令实例
	Root = NewCmd(cmdName, "", ExitOnError)
}

// Parse 解析命令行参数
//
// 参数:
//   - args: 命令行参数列表
//
// 返回值:
//   - error: 解析失败时返回错误
//
// 功能说明:
//   - 实现types.Command接口
//   - 使用sync.Once确保只解析一次
//   - 调用解析器的Parse方法
//   - 递归解析所有子命令
func Parse() error {
	return Root.Parse(os.Args[1:])
}

// ParseOnly 仅解析当前命令, 不递归解析子命令
//
// 参数:
//   - args: 命令行参数列表
//
// 返回值:
//   - error: 解析失败时返回错误
//
// 功能说明:
//   - 实现types.Command接口
//   - 使用sync.Once确保只执行一次
//   - 调用解析器的ParseOnly方法
//   - 不处理子命令解析
func ParseOnly() error {
	return Root.ParseOnly(os.Args[1:])
}

// ParseAndRoute 解析并路由执行命令
//
// 参数:
//   - args: 命令行参数列表
//
// 返回值:
//   - error: 解析或执行失败时返回错误
//
// 功能说明:
//   - 实现types.Command接口
//   - 使用sync.Once确保只执行一次
//   - 调用解析器的ParseAndRoute方法
//   - 完整的解析和执行流程
func ParseAndRoute() error {
	return Root.ParseAndRoute(os.Args[1:])
}

// AddSubCmds 添加子命令到全局根命令
//
// 参数:
//   - cmd: 要添加的子命令实例
//
// 返回值:
//   - error: 添加子命令过程中遇到的错误, 如果没有错误则返回 nil
func AddSubCmds(cmds ...Command) error {
	return Root.AddSubCmds(cmds...)
}

// AddSubCmdFrom 从切片添加子命令
//
// 参数:
//   - cmds: 要添加的子命令实例切片
//
// 返回值:
//   - error: 添加子命令过程中遇到的错误, 如果没有错误则返回 nil
func AddSubCmdFrom(cmds []Command) error {
	return Root.AddSubCmdFrom(cmds)
}

// AddMutexGroup 添加互斥组到命令
//
// 参数:
//   - name: 互斥组名称, 用于错误提示和标识
//   - flags: 互斥组中的标志名称列表
//   - allowNone: 是否允许一个都不设置
//
// 功能说明:
//   - 创建新的互斥组并添加到命令配置中
//   - 互斥组中的标志最多只能有一个被设置
//   - 如果 allowNone 为 false, 则必须至少有一个标志被设置
//   - 使用写锁保护并发安全
//
// 注意事项:
//   - 标志名称必须是已注册的标志
//   - 互斥组名称在命令中应该唯一
//   - 如果组名已存在，返回错误
//
// 返回值:
//   - error: 添加失败时返回错误
func AddMutexGroup(name string, flags []string, allowNone bool) error {
	return Root.AddMutexGroup(name, flags, allowNone)
}

// AddRequiredGroup 添加必需组到命令
//
// 参数:
//   - name: 必需组名称, 用于错误提示和标识
//   - flags: 必需组中的标志名称列表
//
// 功能说明:
//   - 创建新的必需组并添加到命令配置中
//   - 必需组中的所有标志都必须被设置
//   - 使用写锁保护并发安全
//
// 注意事项:
//   - 标志名称必须是已注册的标志
//   - 必需组名称在命令中应该唯一
//   - 如果组名已存在，返回错误
//
// 返回值:
//   - error: 添加失败时返回错误
func AddRequiredGroup(name string, flags []string) error {
	return Root.AddRequiredGroup(name, flags)
}

// ApplyOpts 应用选项到全局根命令
//
// 参数:
//   - opts: 要应用的选项结构体实例
//
// 返回值:
//   - error: 应用选项过程中遇到的错误, 如果没有错误则返回 nil
//
// 功能说明:
//   - 将选项结构体的所有属性应用到全局根命令实例
//   - 支持部分配置（未设置的属性不会被修改）
//   - 使用写锁保护并发安全
func ApplyOpts(opts *CmdOpts) error {
	return Root.ApplyOpts(opts)
}
