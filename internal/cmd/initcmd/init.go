package initcmd

import (
	"bufio"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gitee.com/MM-Q/gob/internal/globls"
	"gitee.com/MM-Q/qflag"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

// InitData 初始化模板数据
type InitData struct {
	ProjectName string // 项目名称
}

// initTemplate 自定义模板初始化，设置不同分隔符避免与 TOML 占位符冲突
func initTemplate(name, content string) (*template.Template, error) {
	// 使用 <| |> 作为模板分隔符，避免与 {{ }} 冲突
	return template.New(name).Delims("<|", "|>").Parse(content)
}

func init() {
	InitCmd = qflag.NewCmd("init", "i", qflag.ExitOnError)
	initCmdCfg := qflag.CmdConfig{
		Desc:       "初始化gob构建文件",
		UseChinese: true,
	}
	InitCmd.ApplyConfig(initCmdCfg)

	forceFlag = InitCmd.Bool("force", "f", false, "强制生成，覆盖已存在的文件")
	nameFlag = InitCmd.String("name", "n", "", "指定生成的项目名称, 默认从go.mod读取")

	// 设置运行函数
	InitCmd.SetRun(run)
}

// run 执行初始化命令
func run(cmd *qflag.Cmd) error {
	// 获取项目名称
	projectName := getProjectName()
	if nameFlag.Get() != "" {
		projectName = nameFlag.Get()
	}

	if projectName == "" {
		return fmt.Errorf("无法获取项目名称，请通过 --name 指定或确保当前目录存在 go.mod 文件")
	}

	globls.CL.Greenf("%s 项目名称: %s\n", globls.PrintPrefix, projectName)

	// 创建 gobf 目录
	gobfDir := "gobf"
	if err := ensureDirectory(gobfDir); err != nil {
		return fmt.Errorf("创建 gobf 目录失败: %w", err)
	}

	// 准备模板数据
	data := InitData{
		ProjectName: projectName,
	}

	// 生成配置文件
	configs := []string{"dev", "install", "release"}
	for _, name := range configs {
		if err := renderAndWriteConfig(data, gobfDir, name); err != nil {
			return err
		}
	}

	globls.CL.Greenf("%s 初始化完成！已生成 gobf/ 目录及配置文件\n", globls.PrintPrefix)
	return nil
}

// getProjectName 从 go.mod 读取项目名称
func getProjectName() string {
	// 检查 go.mod 是否存在
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		return ""
	}

	// 读取 go.mod 文件
	file, err := os.Open("go.mod")
	if err != nil {
		return ""
	}
	defer func() { _ = file.Close() }()

	// 逐行读取，查找 module 声明
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			// 提取模块名称的最后一部分作为项目名称
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				modulePath := parts[1]
				// 获取路径的最后一部分（如 gitee.com/MM-Q/gob -> gob）
				return filepath.Base(modulePath)
			}
		}
	}

	return ""
}

// ensureDirectory 确保目录存在
func ensureDirectory(dir string) error {
	// 检查目录是否已存在
	if _, err := os.Stat(dir); err == nil {
		if !forceFlag.Get() {
			globls.CL.Yellowf("%s gobf 目录已存在，如需覆盖请使用 --force/-f 参数\n", globls.PrintPrefix)
			return fmt.Errorf("目录已存在: %s", dir)
		}
		globls.CL.Yellowf("%s gobf 目录已存在，使用 --force/-f 参数覆盖\n", globls.PrintPrefix)
	}

	// 创建目录
	return os.MkdirAll(dir, 0755)
}

// renderAndWriteConfig 渲染并写入配置文件
func renderAndWriteConfig(data InitData, dir, name string) error {
	// 模板文件路径
	tmplPath := fmt.Sprintf("templates/%s.tmpl", name)

	// 读取模板内容
	tmplContent, err := templateFS.ReadFile(tmplPath)
	if err != nil {
		return fmt.Errorf("读取模板文件 %s 失败: %w", tmplPath, err)
	}

	// 解析模板（使用自定义分隔符避免与 TOML 占位符冲突）
	tmpl, err := initTemplate(name, string(tmplContent))
	if err != nil {
		return fmt.Errorf("解析模板 %s 失败: %w", name, err)
	}

	// 输出文件路径
	outputPath := filepath.Join(dir, name+".toml")

	// 检查文件是否已存在
	if _, err := os.Stat(outputPath); err == nil && !forceFlag.Get() {
		globls.CL.Yellowf("%s 配置文件已存在: %s\n", globls.PrintPrefix, outputPath)
		return fmt.Errorf("文件已存在: %s", outputPath)
	}

	// 创建文件
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建文件 %s 失败: %w", outputPath, err)
	}
	defer func() { _ = file.Close() }()

	// 执行模板并写入文件
	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("渲染模板 %s 失败: %w", name, err)
	}

	globls.CL.Greenf("%s 已生成: %s\n", globls.PrintPrefix, outputPath)
	return nil
}
