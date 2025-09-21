// Package shellx 提供了一个功能完善、易于使用的Go语言shell命令执行库。
//
// 本库基于Go标准库的os/exec包进行封装，提供了更加友好的API和丰富的功能，
// 支持同步和异步命令执行、输入输出重定向、超时控制、上下文管理、
// 多种shell类型支持等功能，并提供类型安全的API和友好的链式调用接口。
//
// 主要特性：
//   - 支持三种命令创建方式：NewCmd(可变参数)、NewCmds(切片)、NewCmdStr(字符串解析)
//   - 链式调用API，支持流畅的方法链
//   - 完整的错误处理和类型安全
//   - 支持多种shell类型（sh、bash、cmd、powershell、pwsh等）
//   - 同步和异步执行支持
//   - 命令执行状态管理和进程控制
//   - 输入输出重定向和环境变量设置
//   - 超时控制和上下文取消
//   - 并发安全的设计
//   - 跨平台兼容（Windows、Linux、macOS）
//
// 核心组件：
//   - Builder: 命令构建器，提供链式调用API
//   - Command: 命令执行对象，封装exec.Cmd并提供额外功能
//   - Result: 命令执行结果，包含输出、错误、时间等信息
//   - ShellType: Shell类型枚举，支持多种shell
//
// 基本用法：
//
//	import "gitee.com/MM-Q/shellx"
//
//	// 方式1：使用可变参数创建命令
//	cmd := shellx.NewCmd("ls", "-la").
//		WithWorkDir("/tmp").
//		WithTimeout(30 * time.Second).
//		WithShell(shellx.ShellBash).
//		Build()
//
//	// 方式2：使用字符串创建命令
//	cmd := shellx.NewCmdStr(`echo "hello world"`).
//		WithEnv("MY_VAR", "value").
//		Build()
//
//	// 同步执行
//	err := cmd.Exec()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// 获取输出
//	output, err := cmd.ExecOutput()
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(string(output))
//
//	// 获取完整结果
//	result := cmd.ExecResult()
//	fmt.Printf("Exit Code: %d\n", result.Code())
//	fmt.Printf("Success: %t\n", result.Success())
//	fmt.Printf("Duration: %v\n", result.Duration())
//	fmt.Printf("Output: %s\n", result.Output())
//
//	// 异步执行
//	err = cmd.ExecAsync()
//	if err != nil {
//		log.Fatal(err)
//	}
//	// 等待完成
//	err = cmd.Wait()
//
// 高级用法：
//
//	// 设置标准输入输出
//	var stdout, stderr bytes.Buffer
//	stdin := strings.NewReader("input data")
//
//	cmd := shellx.NewCmd("cat").
//		WithStdin(stdin).
//		WithStdout(&stdout).
//		WithStderr(&stderr).
//		Build()
//
//	// 使用上下文控制
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
//
//	cmd := shellx.NewCmd("long-running-command").
//		WithContext(ctx).
//		Build()
//
//	// 进程控制
//	cmd.ExecAsync()
//	pid := cmd.GetPID()
//	isRunning := cmd.IsRunning()
//	cmd.Kill() // 或 cmd.Signal(syscall.SIGTERM)
//
// 命令解析：
//
//	// 支持复杂的命令字符串解析，包括引号处理
//	cmd := shellx.NewCmdStr(`git commit -m "Initial commit with 'quotes'"`)
//	// 解析结果：["git", "commit", "-m", "Initial commit with 'quotes'"]
//
// Shell类型：
//
//	// 支持多种shell类型
//	shellx.ShellSh         // sh shell
//	shellx.ShellBash       // bash shell
//	shellx.ShellCmd        // Windows cmd
//	shellx.ShellPowerShell // Windows PowerShell
//	shellx.ShellPwsh       // PowerShell Core
//	shellx.ShellNone       // 直接执行，不使用shell
//	shellx.ShellDefault    // 根据操作系统自动选择
//
// 注意事项：
//   - 每个Command对象只能执行一次，重复执行会返回错误
//   - Builder是并发安全的，可以在多个goroutine中安全使用
//   - 命令执行会继承父进程的环境变量，可通过WithEnv添加额外变量
//   - 超时设置仅在支持的Go版本中有效
//   - 异步执行需要调用Wait()等待完成或使用Kill()终止
package shellx
