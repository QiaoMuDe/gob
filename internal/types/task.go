package types

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gitee.com/MM-Q/shellx"
	"github.com/pelletier/go-toml/v2"
)

// TaskFileConfig 表示任务式配置文件的完整结构
// 对应gobfile.toml或task.toml配置文件的结构
type TaskFileConfig struct {
	Global GlobalConfig           `toml:"global" comment:"全局配置"`
	Tasks  map[string]*TaskConfig `toml:"task" comment:"任务定义"` // 注意: toml中是task而不是tasks
}

// GlobalConfig 表示全局配置项
// 对应task.toml中的[global]部分
type GlobalConfig struct {
	Envs        map[string]string `toml:"envs" comment:"全局环境变量"`                 // 全局环境变量
	Vars        map[string]string `toml:"vars" comment:"全局变量"`                   // 全局变量
	WorkDir     string            `toml:"work_dir" comment:"工作目录"`               // 工作目录
	Timeout     string            `toml:"timeout" comment:"超时时间"`                // 超时时间
	ShowOutput  bool              `toml:"show_output" comment:"是否显示输出"`          // 是否显示输出
	ExitOnError bool              `toml:"exit_on_error" comment:"任务执行失败时是否退出程序"` // 任务执行失败时是否退出程序
	ShowCmd     bool              `toml:"show_cmd" comment:"是否显示执行的命令"`          // 是否显示执行的命令
}

// TaskConfig 表示单个任务的配置项
// 对应task.toml中的[task.xxx]部分
type TaskConfig struct {
	Desc       string            `toml:"desc" comment:"任务描述"`          // 任务描述
	Cmds       []string          `toml:"cmds" comment:"命令列表"`          // 命令列表
	Envs       map[string]string `toml:"envs" comment:"任务环境变量"`        // 任务环境变量
	Vars       map[string]string `toml:"vars" comment:"任务变量"`          // 任务变量
	WorkDir    string            `toml:"work_dir" comment:"任务工作目录"`    // 任务工作目录
	Timeout    string            `toml:"timeout" comment:"任务超时时间"`     // 任务超时时间
	DependsOn  []string          `toml:"depends_on" comment:"依赖任务列表"`  // 依赖任务列表
	ShowOutput bool              `toml:"show_output" comment:"是否显示输出"` // 是否显示输出
}

// TaskExecutionContext 表示任务执行上下文
// 包含执行任务时所需的所有信息
type TaskExecutionContext struct {
	GlobalConfig *GlobalConfig          // 全局配置
	TaskConfig   *TaskConfig            // 当前任务配置
	TaskName     string                 // 任务名称
	AllTasks     map[string]*TaskConfig // 所有任务配置
	Envs         []string               // 最终环境变量列表
}

// LoadTaskConfig 从指定路径加载任务式配置文件并解析为TaskFileConfig结构体
//
// 参数:
//   - filePath: TOML配置文件的路径
//
// 返回:
//   - 解析后的TaskFileConfig结构体指针和可能的错误
func LoadTaskConfig(filePath string) (*TaskFileConfig, error) {
	// 创建核心配置结构体（只包含核心全局配置和空任务映射）
	config := GetCoreConfig()

	// 如果文件不存在, 则返回默认配置
	if info, err := os.Stat(filePath); os.IsNotExist(err) {
		return config, nil
	} else if info.IsDir() {
		return nil, fmt.Errorf("file '%s' is a directory", filePath)
	}

	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// 解析TOML内容到配置结构体
	if err := toml.Unmarshal(content, config); err != nil {
		// 提取TOML解析错误的详细位置信息
		if decodeErr, ok := err.(*toml.DecodeError); ok {
			row, col := decodeErr.Position() // 获取行和列信息
			return nil, fmt.Errorf("TOML解析错误 (行 %d, 列 %d): %v", row, col, decodeErr.Error())
		}
		return nil, fmt.Errorf("加载任务配置文件 %s 失败: %w", filePath, err)
	}

	return config, nil
}

// GetDefaultTaskConfig 获取任务配置的默认值
// 包含完整的示例配置，用于生成配置文件
//
// 返回值:
//   - *TaskFileConfig: 包含所有默认配置值的结构体指针
func GetDefaultTaskConfig() *TaskFileConfig {
	return &TaskFileConfig{
		Global: GlobalConfig{
			Envs: map[string]string{
				"GOOS":   "windows",
				"GOARCH": "amd64",
			}, // 示例全局环境变量
			Vars: map[string]string{
				"app_name": "myapp",
				"version":  "1.0.0",
			}, // 示例全局变量
			WorkDir:     ".",   // 默认当前目录
			Timeout:     "30s", // 默认30秒超时
			ShowOutput:  true,  // 默认显示输出
			ExitOnError: true,  // 默认遇到错误时退出
			ShowCmd:     false, // 默认不显示执行的命令
		},
		Tasks: map[string]*TaskConfig{
			"run": {
				Desc: "运行应用程序",
				Cmds: []string{
					"echo 启动应用...",
					"echo 应用已启动",
				},
				Envs: map[string]string{
					"PORT": "8080",
				}, // 示例任务环境变量
				Vars: map[string]string{
					"run_mode": "production",
				}, // 示例任务变量
				WorkDir:    ".",        // 任务工作目录
				Timeout:    "60s",      // 任务超时时间
				DependsOn:  []string{}, // 无依赖
				ShowOutput: true,       // 显示输出
			},
		},
	}
}

// GetCoreConfig 获取核心全局配置
// 只包含核心全局配置和空任务映射，用于加载配置文件
//
// 返回值:
//   - *TaskFileConfig: 包含核心配置的结构体指针
func GetCoreConfig() *TaskFileConfig {
	return &TaskFileConfig{
		Global: GlobalConfig{
			Envs:        make(map[string]string), // 默认空环境变量
			Vars:        make(map[string]string), // 默认空变量
			WorkDir:     ".",                     // 默认当前目录
			Timeout:     "30s",                   // 默认30秒超时
			ShowOutput:  true,                    // 默认显示输出
			ExitOnError: true,                    // 默认遇到错误时退出
			ShowCmd:     false,                   // 默认不显示执行的命令
		},
		Tasks: make(map[string]*TaskConfig), // 默认空任务映射
	}
}

// GenerateDefaultTaskConfig 生成默认的任务式配置文件
//
// 参数值:
//   - filePath: 配置文件路径
//   - f: 是否强制覆盖已存在的配置文件
//
// 返回值:
//   - error: 错误信息，如果生成成功则返回nil
func GenerateDefaultTaskConfig(filePath string, f bool) error {
	// 获取完整的示例配置（包含所有字段和示例任务）
	config := GetDefaultTaskConfig()

	// 检查配置文件是否已存在
	if _, err := os.Stat(filePath); err == nil {
		// 如果没有启用f, 则返回错误
		if !f {
			return fmt.Errorf("配置文件 %s 已存在，使用 --force/-f 强制覆盖", filePath)
		}
	}

	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建任务配置文件失败: %v", err)
	}
	defer func() { _ = file.Close() }()

	// 使用toml.Marshal序列化配置
	data, err := toml.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化任务配置文件失败: %v", err)
	}

	// 写入文件
	// 先写入配置文件注释
	comment := []byte(TaskConfigFileHeaderComment)
	if _, err := file.Write(comment); err != nil {
		return fmt.Errorf("写入注释失败: %v", err)
	}

	// 再写入配置数据
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("写入任务配置文件失败: %v", err)
	}

	return nil
}

// ResolveTaskDependencies 解析任务依赖关系并返回按执行顺序排列的任务列表
//
// 参数:
//   - taskName: 要执行的任务名称
//   - tasks: 所有任务配置
//
// 返回值:
//   - []string: 按依赖顺序排列的任务列表
//   - error: 错误信息
func ResolveTaskDependencies(taskName string, tasks map[string]*TaskConfig) ([]string, error) {
	// 实现拓扑排序算法
	visited := make(map[string]bool)
	tempMark := make(map[string]bool)
	result := make([]string, 0)

	var visit func(string) error
	visit = func(name string) error {
		if _, exists := tasks[name]; !exists {
			return fmt.Errorf("任务 '%s' 不存在", name)
		}

		if tempMark[name] {
			return fmt.Errorf("检测到任务依赖循环: %s", name)
		}

		if visited[name] {
			return nil
		}

		tempMark[name] = true

		// 先访问所有依赖
		task := tasks[name]
		for _, dep := range task.DependsOn {
			if err := visit(dep); err != nil {
				return err
			}
		}

		tempMark[name] = false
		visited[name] = true
		result = append(result, name)

		return nil
	}

	if err := visit(taskName); err != nil {
		return nil, err
	}

	// 反转结果，因为我们是从依赖到被依赖的顺序添加的
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result, nil
}

// ResolveTaskValue 解析任务变量值
// 如果值以@开头，则作为命令执行，返回命令输出
// 否则返回原字符串值
//
// 参数:
//   - value: 变量值
//   - context: 任务执行上下文
//
// 返回值:
//   - string: 解析后的值
//   - error: 错误信息
func ResolveTaskValue(value string, context *TaskExecutionContext) (string, error) {
	// 检查是否以@开头（命令执行）
	if strings.HasPrefix(value, "@") {
		// 提取命令部分
		cmdStr := strings.TrimSpace(value[1:])
		if cmdStr == "" {
			return "", fmt.Errorf("命令不能为空")
		}

		// 执行命令并获取输出
		output, err := executeCommandForVariable(cmdStr, context)
		if err != nil {
			return "", fmt.Errorf("执行命令 '%s' 失败: %v", cmdStr, err)
		}

		// 去除输出末尾的换行符
		return strings.TrimSpace(output), nil
	}

	// 普通字符串，直接返回
	return value, nil
}

// executeCommandForVariable 执行命令获取变量值
// 根据上下文自动确定工作目录、超时时间和环境变量
// 根据操作系统自动选择Shell类型：Windows使用PowerShell，Linux/macOS使用bash
//
// 参数:
//   - cmdStr: 要执行的命令字符串
//   - context: 任务执行上下文
//
// 返回值:
//   - string: 命令输出
//   - error: 错误信息
func executeCommandForVariable(cmdStr string, context *TaskExecutionContext) (string, error) {
	// 准备命令执行参数
	workDir, timeout, envs := PrepareCommandExecution(context.TaskConfig, context.GlobalConfig, os.Environ())

	// 创建命令并设置参数
	cmd := shellx.NewCmdStr(cmdStr).
		WithEnvs(envs).
		WithWorkDir(workDir).
		WithShell(shellx.ShellDef2).
		WithTimeout(timeout)

	// 执行命令并返回输出
	output, err := cmd.ExecOutput()
	return string(output), err
}

// ReplaceTaskVariables 替换任务中的变量占位符（简化版）
// 只处理变量和全局配置的替换，环境变量直接从系统获取
//
// 参数:
//   - template: 包含占位符的模板字符串
//   - taskName: 任务名称
//   - context: 任务执行上下文
//
// 返回值:
//   - string: 替换后的字符串
func ReplaceTaskVariables(template string, taskName string, context *TaskExecutionContext) string {
	result := template

	// 解析全局变量（支持@符号执行命令）
	for key, value := range context.GlobalConfig.Vars {
		if resolvedValue, err := ResolveTaskValue(value, context); err == nil {
			placeholder := fmt.Sprintf("{{global.vars.%s}}", key)
			result = strings.ReplaceAll(result, placeholder, resolvedValue)
		}
	}

	// 解析任务变量（支持@符号执行命令）
	if task, exists := context.AllTasks[taskName]; exists {
		for key, value := range task.Vars {
			if resolvedValue, err := ResolveTaskValue(value, context); err == nil {
				placeholder := fmt.Sprintf("{{task.%s.vars.%s}}", taskName, key)
				result = strings.ReplaceAll(result, placeholder, resolvedValue)
			}
		}
	}

	return result
}

// ParseTimeout 解析超时时间字符串为time.Duration
//
// 参数:
//   - timeoutStr: 超时时间字符串
//
// 返回值:
//   - time.Duration: 解析后的时间间隔
//   - error: 错误信息
func ParseTimeout(timeoutStr string) (time.Duration, error) {
	if timeoutStr == "" {
		return 30 * time.Second, nil // 默认30秒
	}
	return time.ParseDuration(timeoutStr)
}

// PrepareCommandExecution 准备命令执行参数
// 统一处理工作目录、超时时间和环境变量的逻辑
//
// 参数:
//   - taskConfig: 任务配置
//   - globalConfig: 全局配置
//   - baseEnvs: 基础环境变量列表（通常为系统环境变量）
//
// 返回值:
//   - workDir: 最终的工作目录
//   - timeout: 最终的超时时间
//   - envs: 最终的环境变量列表
func PrepareCommandExecution(taskConfig *TaskConfig, globalConfig *GlobalConfig, baseEnvs []string) (workDir string, timeout time.Duration, envs []string) {
	// 确定工作目录
	workDir = taskConfig.WorkDir
	if workDir == "" {
		workDir = globalConfig.WorkDir
	}
	if workDir == "" {
		workDir = "." // 最后的后备：当前目录
	}

	// 解析超时时间
	var err error
	timeout, err = ParseTimeout(taskConfig.Timeout)
	if err != nil {
		timeout, err = ParseTimeout(globalConfig.Timeout)
		if err != nil {
			timeout = 30 * time.Second // 最后的后备：30秒
		}
	}

	// 合并环境变量
	envs = make([]string, len(baseEnvs))
	copy(envs, baseEnvs)

	// 添加全局环境变量
	for k, v := range globalConfig.Envs {
		envs = append(envs, fmt.Sprintf("%s=%s", k, v))
	}

	// 添加任务环境变量（优先级更高）
	for k, v := range taskConfig.Envs {
		envs = append(envs, fmt.Sprintf("%s=%s", k, v))
	}

	return workDir, timeout, envs
}

// 任务式配置文件头部注释
const TaskConfigFileHeaderComment = `# gob 任务编排文件
# 项目地址: https://gitee.com/MM-Q/gob.git
# 使用方法: gob task --run <任务名>

`

// 任务式配置文件名
const (
	TaskConfigFileName    = "task.toml"
	AltTaskConfigFileName = "Task.toml"
)
