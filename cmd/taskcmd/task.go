package taskcmd

import (
	"fmt"
	"os"
	"strings"

	"gitee.com/MM-Q/gob/internal/types"
	"gitee.com/MM-Q/gob/internal/utils"
	"gitee.com/MM-Q/qflag"
	"gitee.com/MM-Q/shellx"
)

// getTaskConfigPath 获取任务配置文件路径
// 优先使用 --file 参数指定的路径, 否则使用默认路径
// 如果文件不存在且未指定 --file, 则尝试备用文件名
//
// 参数:
//   - allowCreate: 是否允许文件不存在 (用于初始化场景)
//
// 返回值:
//   - string: 配置文件路径
//   - error: 错误信息
func getTaskConfigPath(allowCreate bool) (string, error) {
	// 确定配置文件路径
	configPath := types.TaskConfigFileName
	if fileFlag.Get() != "" {
		configPath = fileFlag.Get()
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 如果允许创建文件, 直接返回路径 (不检查备用文件名)
		if allowCreate {
			return configPath, nil
		}

		// 尝试备用文件名
		if fileFlag.Get() == "" { // 只有在未指定文件时才尝试备用文件名
			configPath = types.AltTaskConfigFileName
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				return "", fmt.Errorf("任务 ['%s'|'%s'] 配置文件不存在", types.TaskConfigFileName, types.AltTaskConfigFileName)
			}
		} else {
			return "", fmt.Errorf("指定的任务配置文件不存在: %s", configPath)
		}
	}

	return configPath, nil
}

// run 任务命令主入口
// 根据传入的标志执行相应的操作: 初始化配置、列出任务或运行指定任务
//
// 参数:
//   - cmd: 命令对象
//
// 返回值:
//   - error: 错误信息
func run(cmd *qflag.Cmd) error {
	// 处理 --example 参数
	if exampleFlag.Get() {
		return printTaskExample()
	}

	// 处理 --check 参数
	if checkFlag.Get() {
		return checkTaskConfig()
	}

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

	// 如果没有指定参数, 显示帮助信息
	cmd.PrintHelp()
	return nil
}

// checkTaskConfig 校验任务配置文件
// 检查配置文件的格式、字段完整性和业务逻辑正确性
//
// 返回值:
//   - error: 错误信息
func checkTaskConfig() error {
	// 获取配置文件路径
	configPath, err := getTaskConfigPath(false)
	if err != nil {
		return err
	}

	utils.Logf("正在校验任务配置文件: %s\n", configPath)

	// 加载配置文件
	config, err := types.LoadTaskConfig(configPath)
	if err != nil {
		return err
	}

	// 执行校验
	validationErrors := validateTaskConfig(config)

	// 显示校验结果
	if len(validationErrors) == 0 {
		utils.Logf("配置文件校验通过，未发现问题")
		return nil
	}

	utils.CL.Red("校验失败, 发现以下问题:")
	for i, errMsg := range validationErrors {
		fmt.Printf("  %d. %s\n", i+1, errMsg)
	}

	return fmt.Errorf("配置文件校验失败，发现 %d 个问题", len(validationErrors))
}

// validateTaskConfig 校验任务配置内容
// 检查全局配置、任务配置和依赖关系的正确性
//
// 参数:
//   - config: 任务配置
//
// 返回值:
//   - []string: 错误信息列表
func validateTaskConfig(config *types.TaskFileConfig) []string {
	var errors []string

	// 校验全局配置
	errors = append(errors, validateGlobalConfig(&config.Global)...)

	// 校验任务配置
	for taskName, task := range config.Tasks {
		taskErrors := validateTaskConfigItem(taskName, task, config.Tasks)
		errors = append(errors, taskErrors...)
	}

	// 校验任务依赖关系
	depErrors := validateTaskDependencies(config.Tasks)
	errors = append(errors, depErrors...)

	return errors
}

// printTaskExample 打印完整的任务配置示例
//
// 返回值:
//   - error: 错误信息
func printTaskExample() error {
	fmt.Println(exampleConfig)
	return nil
}

// validateGlobalConfig 校验全局配置
//
// 参数:
//   - global: 全局配置
//
// 返回值:
//   - []string: 错误信息列表
func validateGlobalConfig(global *types.GlobalConfig) []string {
	var errors []string

	// 校验超时时间格式
	if global.Timeout != "" {
		if _, err := types.ParseTimeout(global.Timeout); err != nil {
			errors = append(errors, fmt.Sprintf("全局配置: 超时时间格式错误 '%s' (应为有效的时间间隔，如 '30s', '5m')", global.Timeout))
		}
	}

	// 校验工作目录
	if global.WorkDir != "" {
		if info, err := os.Stat(global.WorkDir); os.IsNotExist(err) {
			errors = append(errors, fmt.Sprintf("全局配置: 工作目录不存在 '%s'", global.WorkDir))
		} else if err != nil {
			errors = append(errors, fmt.Sprintf("全局配置: 无法访问工作目录 '%s': %v", global.WorkDir, err))
		} else if !info.IsDir() {
			errors = append(errors, fmt.Sprintf("全局配置: 工作路径不是目录 '%s'", global.WorkDir))
		}
	}

	return errors
}

// validateTaskConfigItem 校验单个任务配置
//
// 参数:
//   - taskName: 任务名称
//   - task: 任务配置
//   - allTasks: 所有任务配置（用于依赖检查）
//
// 返回值:
//   - []string: 错误信息列表
func validateTaskConfigItem(taskName string, task *types.TaskConfig, allTasks map[string]*types.TaskConfig) []string {
	var errors []string

	// 检查必需字段
	if len(task.Cmds) == 0 {
		errors = append(errors, fmt.Sprintf("任务 '%s': 缺少命令列表 (cmds)", taskName))
	}

	// 校验任务工作目录
	if task.WorkDir != "" {
		if info, err := os.Stat(task.WorkDir); os.IsNotExist(err) {
			errors = append(errors, fmt.Sprintf("任务 '%s': 工作目录不存在 '%s'", taskName, task.WorkDir))
		} else if err != nil {
			errors = append(errors, fmt.Sprintf("任务 '%s': 无法访问工作目录 '%s': %v", taskName, task.WorkDir, err))
		} else if !info.IsDir() {
			errors = append(errors, fmt.Sprintf("任务 '%s': 工作路径不是目录 '%s'", taskName, task.WorkDir))
		}
	}

	// 校验超时时间格式
	if task.Timeout != "" {
		if _, err := types.ParseTimeout(task.Timeout); err != nil {
			errors = append(errors, fmt.Sprintf("任务 '%s': 超时时间格式错误 '%s' (应为有效的时间间隔，如 '30s', '5m')", taskName, task.Timeout))
		}
	}

	// 检查依赖任务是否存在
	for _, depName := range task.DependsOn {
		if _, exists := allTasks[depName]; !exists {
			errors = append(errors, fmt.Sprintf("任务 '%s': 依赖的任务 '%s' 不存在", taskName, depName))
		}
	}

	return errors
}

// validateTaskDependencies 校验所有任务的依赖关系
// 检查是否存在循环依赖
//
// 参数:
//   - tasks: 所有任务配置
//
// 返回值:
//   - []string: 错误信息列表
func validateTaskDependencies(tasks map[string]*types.TaskConfig) []string {
	var errors []string

	// 对每个任务进行依赖关系校验
	for taskName := range tasks {
		_, err := types.ResolveTaskDependencies(taskName, tasks)
		if err != nil {
			if strings.Contains(err.Error(), "检测到任务依赖循环") {
				errors = append(errors, fmt.Sprintf("依赖关系: %s", err.Error()))
			}
		}
	}

	return errors
}

// initTaskConfig 初始化任务配置文件
// 生成默认的任务配置文件, 如果文件已存在且未启用强制标志, 则返回错误
//
// 返回值:
//   - error: 错误信息
func initTaskConfig() error {
	configPath, err := getTaskConfigPath(true)
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
	configPath, err := getTaskConfigPath(false)
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
// 加载配置文件, 解析任务依赖关系, 并按正确顺序执行任务
//
// 参数:
//   - taskName: 要运行的任务名称
//
// 返回值:
//   - error: 错误信息
func runTask(taskName string) error {
	// 获取配置文件路径
	configPath, err := getTaskConfigPath(false)
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
			utils.TaskErrf(name, "执行任务失败: %v\n", err)
			if config.Global.ExitOnError {
				return err
			}
		}
	}

	utils.TaskLog(taskName, "任务执行完成")
	return nil
}

// executeTask 执行单个任务
// 设置任务环境变量、工作目录等参数, 然后执行任务中的命令列表
// 根据操作系统自动选择Shell类型: Windows使用PowerShell, Linux/macOS使用bash
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

	// 确定是否显示命令输出
	showOutput := task.ShowOutput
	if !task.ShowOutput && context.GlobalConfig.ShowOutput {
		showOutput = context.GlobalConfig.ShowOutput
	}

	// 确定是否显示执行的命令
	showCommand := context.GlobalConfig.ShowCmd

	// 显示任务描述
	if task.Desc != "" {
		utils.TaskLogf(taskName, "执行任务: %s\n", task.Desc)
	} else {
		utils.TaskLog(taskName, "执行任务")
	}

	// 执行任务命令
	for _, cmdStr := range task.Cmds {
		// 替换变量占位符 (包括解析@开头的命令)
		resolvedCmd := types.ReplaceTaskVariables(cmdStr, taskName, context)

		// 显示执行的命令 (如果启用)
		if showCommand {
			utils.TaskLogf(taskName, "执行命令: %s\n", resolvedCmd)
		}

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

var exampleConfig = `# gob 任务编排文件
# 项目地址: https://gitee.com/MM-Q/gob.git
# 使用方法: gob task --run <任务名>

# ==================== 全局配置 ====================
[global]
# 工作目录 (默认: .) 
work_dir = '.'

# 超时时间 (默认: 30s) 
timeout = '30s'

# 是否显示输出 (默认: true) 
show_output = true

# 任务执行失败时是否退出程序 (默认: true) 
exit_on_error = true

# 是否显示执行的命令 (默认: false) 
show_cmd = false

# 全局环境变量 (可选) 
[global.envs]
# GOOS = "windows"
# GOARCH = "amd64"

# 全局变量 (可选) 
[global.vars]
# app_name = "myapp"
# version = "1.0.0"

# ==================== 任务定义 ====================
# 任务名称 (示例: task.task_name) 
[task.example]
# 任务描述 (可选) 
desc = '任务描述'

# 命令列表 (必需) 
cmds = [
    'echo "执行命令1"',
    'echo "执行命令2"'
]

# 任务环境变量 (可选) 
[task.example.envs]
# ENV_VAR = "value"

# 任务变量 (可选) 
[task.example.vars]
# task_var = "value"

# 任务工作目录 (可选, 默认使用全局配置) 
work_dir = '.'

# 任务超时时间 (可选, 默认使用全局配置) 
timeout = '60s'

# 依赖任务列表 (可选) 
depends_on = ['other_task']

# 是否显示输出 (可选, 默认使用全局配置) 
show_output = true
`
