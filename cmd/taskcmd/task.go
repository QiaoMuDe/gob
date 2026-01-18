package taskcmd

import (
	"fmt"
	"os"

	"gitee.com/MM-Q/gob/internal/types"
	"gitee.com/MM-Q/gob/internal/utils"
	"gitee.com/MM-Q/qflag"
	"gitee.com/MM-Q/shellx"
)

// getTaskConfigPath 获取任务配置文件路径
// 优先使用 --file 参数指定的路径，否则使用默认路径
// 如果文件不存在且未指定 --file，则尝试备用文件名
//
// 返回值:
//   - string: 配置文件路径
//   - error: 错误信息
func getTaskConfigPath() (string, error) {
	// 确定配置文件路径
	configPath := types.TaskConfigFileName
	if fileFlag.Get() != "" {
		configPath = fileFlag.Get()
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 尝试备用文件名
		if fileFlag.Get() == "" { // 只有在未指定文件时才尝试备用文件名
			configPath = types.AltTaskConfigFileName
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				return "", fmt.Errorf("任务配置文件不存在")
			}
		} else {
			return "", fmt.Errorf("指定的任务配置文件不存在: %s", configPath)
		}
	}

	return configPath, nil
}

// run 任务命令主入口
// 根据传入的标志执行相应的操作：初始化配置、列出任务或运行指定任务
//
// 参数:
//   - cmd: 命令对象
//
// 返回值:
//   - error: 错误信息
func run(cmd *qflag.Cmd) error {
	// 处理 --init 参数
	if initFlag.Get() {
		return initTaskConfig()
	}

	// 处理 --list 参数
	if listFlag.Get() {
		return listTasks()
	}

	// 处理 --run 参数
	taskName := runFlag.Get()
	if taskName != "" {
		return runTask(taskName)
	}

	// 如果没有指定参数，显示帮助信息
	cmd.PrintHelp()
	return nil
}

// initTaskConfig 初始化任务配置文件
// 生成默认的任务配置文件，如果文件已存在且未启用强制标志，则返回错误
//
// 返回值:
//   - error: 错误信息
func initTaskConfig() error {
	configPath, err := getTaskConfigPath()
	if err != nil {
		return err
	}
	if err := types.GenerateDefaultTaskConfig(configPath, forceFlag.Get()); err != nil {
		return err
	}
	utils.Logf("已生成任务配置文件: %s\n", configPath)
	return nil
}

// listTasks 列出所有可用任务
// 从配置文件中加载任务列表并显示每个任务的名称和描述
//
// 返回值:
//   - error: 错误信息
func listTasks() error {
	// 获取配置文件路径
	configPath, err := getTaskConfigPath()
	if err != nil {
		return err
	}

	// 加载配置文件
	config, err := types.LoadTaskConfig(configPath)
	if err != nil {
		utils.CL.PrintError(err)
		return err
	}

	// 列出所有任务
	utils.Log("可用任务列表:")
	for name, task := range config.Tasks {
		fmt.Printf("  %-20s - %-20s\n", utils.CL.Scyan(name), task.Desc)
	}

	utils.CL.Yellow("\nUsage: gob task --run <task-name>")
	return nil
}

// runTask 运行指定任务
// 加载配置文件，解析任务依赖关系，并按正确顺序执行任务
//
// 参数:
//   - taskName: 要运行的任务名称
//
// 返回值:
//   - error: 错误信息
func runTask(taskName string) error {
	// 获取配置文件路径
	configPath, err := getTaskConfigPath()
	if err != nil {
		return err
	}

	// 加载配置文件
	config, err := types.LoadTaskConfig(configPath)
	if err != nil {
		return err
	}

	// 检查任务是否存在
	if _, exists := config.Tasks[taskName]; !exists {
		return fmt.Errorf("任务不存在: %s", taskName)
	}

	// 解析任务依赖关系
	taskOrder, err := types.ResolveTaskDependencies(taskName, config.Tasks)
	if err != nil {
		utils.CL.PrintError(err)
		return err
	}

	// 创建任务执行上下文
	context := &types.TaskExecutionContext{
		GlobalConfig: &config.Global,
		TaskName:     taskName,
		AllTasks:     config.Tasks,
	}

	// 设置环境变量
	context.Envs = os.Environ()
	for k, v := range config.Global.Envs {
		context.Envs = append(context.Envs, fmt.Sprintf("%s=%s", k, v))
	}

	// 按依赖顺序执行任务
	utils.Logf("开始执行: %v\n", taskOrder)
	for _, name := range taskOrder {
		if err := executeTask(name, context); err != nil {
			utils.TaskLogf(name, "执行任务失败: %v", err)
			if config.Global.ExitOnError {
				return err
			}
		}
	}

	utils.TaskLog(taskName, "任务执行完成")
	return nil
}

// executeTask 执行单个任务
// 设置任务环境变量、工作目录等参数，然后执行任务中的命令列表
// 根据操作系统自动选择Shell类型：Windows使用PowerShell，Linux/macOS使用bash
//
// 参数:
//   - taskName: 要执行的任务名称
//   - context: 任务执行上下文
//
// 返回值:
//   - error: 错误信息
func executeTask(taskName string, context *types.TaskExecutionContext) error {
	task := context.AllTasks[taskName]

	// 更新上下文中的当前任务配置
	context.TaskConfig = task

	// 准备命令执行参数
	workDir, timeout, taskEnvs := types.PrepareCommandExecution(task, context.GlobalConfig, context.Envs)

	// 确定是否显示输出
	showOutput := task.ShowOutput
	if !task.ShowOutput && context.GlobalConfig.ShowOutput {
		showOutput = context.GlobalConfig.ShowOutput
	}

	utils.TaskLog(taskName, "执行任务")

	// 执行任务命令
	for _, cmdStr := range task.Cmds {
		// 替换变量占位符（包括解析@开头的命令）
		resolvedCmd := types.ReplaceTaskVariables(cmdStr, taskName, context)

		// 执行命令
		var err error
		if showOutput {
			err = shellx.NewCmdStr(resolvedCmd).WithEnvs(taskEnvs).WithWorkDir(workDir).WithTimeout(timeout).WithShell(shellx.ShellDef2).WithStdout(os.Stdout).WithStderr(os.Stderr).Exec()
		} else {
			err = shellx.NewCmdStr(resolvedCmd).WithEnvs(taskEnvs).WithWorkDir(workDir).WithTimeout(timeout).WithShell(shellx.ShellDef2).Exec()
		}

		if err != nil {
			return fmt.Errorf("命令执行失败: %s, 错误: %v", resolvedCmd, err)
		}
	}

	return nil
}
