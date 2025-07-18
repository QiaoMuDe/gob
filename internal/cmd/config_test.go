package cmd

import (
	"os"
	"reflect"
	"testing"

	"gitee.com/MM-Q/gob/internal/globls"
)

// TestLoadConfig_FileNotFound 测试配置文件不存在时是否返回默认配置
func TestLoadConfig_FileNotFound(t *testing.T) {
	// 使用不存在的文件路径
	config, err := loadConfig("non_existent_config.toml")
	if err != nil {
		t.Fatalf("预期无错误，实际返回错误: %v", err)
	}

	// 验证默认配置值
	defaultInstallPath := getDefaultInstallPath()
	if config.Build.OutputDir != globls.DefaultOutputDir {
		t.Errorf("Build.OutputDir 默认值错误，预期 %s, 实际 %s", globls.DefaultOutputDir, config.Build.OutputDir)
	}
	if config.Build.OutputName != globls.DefaultAppName {
		t.Errorf("Build.OutputName 默认值错误，预期 %s, 实际 %s", globls.DefaultAppName, config.Build.OutputName)
	}
	if config.Install.InstallPath != defaultInstallPath {
		t.Errorf("Install.InstallPath 默认值错误，预期 %s, 实际 %s", defaultInstallPath, config.Install.InstallPath)
	}
	if len(config.Env) != 0 {
		t.Errorf("Env 默认应为空，实际长度 %d", len(config.Env))
	}
}

// TestLoadConfig_ValidFullConfig 测试加载完整配置文件
func TestLoadConfig_ValidFullConfig(t *testing.T) {
	// 创建临时TOML文件
	content := `
[build]
output_dir = "test_dist"
output_name = "test_app"
main_file = "test_main.go"
ldflags = "-X main.version=1.0.0"
use_vendor = true
inject_git_info = true
simple_name = true
proxy = "https://test.proxy"
enable_cgo = true
color_output = true
batch_mode = true
current_platform_only = true
zip_output = true

[install]
install = true
install_path = "/test/install/path"
force = true

[env]
GOOS = "linux"
GOARCH = "amd64"
`

	// 创建临时文件
	f, err := os.CreateTemp("", "config_test_*.toml")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	defer func() {
		if removeErr := os.Remove(f.Name()); removeErr != nil {
			t.Errorf("Failed to remove temp file: %v", err)
		}
	}()

	if _, writeErr := f.WriteString(content); writeErr != nil {
		t.Fatalf("写入临时文件失败: %v", writeErr)
	}
	if closeErr := f.Close(); closeErr != nil {
		t.Errorf("Failed to close temp file: %v", err)
	}

	// 加载配置
	config, err := loadConfig(f.Name())
	if err != nil {
		t.Fatalf("解析配置失败: %v", err)
	}

	// 验证Build配置
	if config.Build.OutputDir != "test_dist" {
		t.Error("Build.OutputDir 解析错误")
	}
	if config.Build.OutputName != "test_app" {
		t.Error("Build.OutputName 解析错误")
	}
	if config.Build.MainFile != "test_main.go" {
		t.Error("Build.MainFile 解析错误")
	}
	if config.Build.Ldflags != "-X main.version=1.0.0" {
		t.Error("Build.Ldflags 解析错误")
	}
	if !config.Build.UseVendor {
		t.Error("Build.UseVendor 解析错误，预期true")
	}
	if !config.Build.InjectGitInfo {
		t.Error("Build.InjectGitInfo 解析错误，预期true")
	}
	if !config.Build.SimpleName {
		t.Error("Build.SimpleName 解析错误，预期true")
	}
	if config.Build.Proxy != "https://test.proxy" {
		t.Error("Build.Proxy 解析错误")
	}
	if !config.Build.EnableCgo {
		t.Error("Build.EnableCgo 解析错误，预期true")
	}
	if !config.Build.ColorOutput {
		t.Error("Build.ColorOutput 解析错误，预期true")
	}
	if !config.Build.BatchMode {
		t.Error("Build.BatchMode 解析错误，预期true")
	}
	if !config.Build.CurrentPlatformOnly {
		t.Error("Build.CurrentPlatformOnly 解析错误，预期true")
	}
	if !config.Build.ZipOutput {
		t.Error("Build.ZipOutput 解析错误，预期true")
	}

	// 验证Install配置
	if !config.Install.Install {
		t.Error("Install.Install 解析错误，预期true")
	}
	if config.Install.InstallPath != "/test/install/path" {
		t.Error("Install.InstallPath 解析错误")
	}
	if !config.Install.Force {
		t.Error("Install.Force 解析错误，预期true")
	}

	// 验证Env配置
	if config.Env["GOOS"] != "linux" {
		t.Error("Env.GOOS 解析错误")
	}
	if config.Env["GOARCH"] != "amd64" {
		t.Error("Env.GOARCH 解析错误")
	}
}

// TestLoadConfig_PartialConfig 测试加载部分配置（未提供的项使用默认值）
func TestLoadConfig_PartialConfig(t *testing.T) {
	content := `
[build]
output_dir = "partial_dist"
use_vendor = true

[install]
force = true

[env]
CGO_ENABLED = "1"
`

	f, err := os.CreateTemp("", "partial_config_test_*.toml")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	defer func() {
		if removeErr := os.Remove(f.Name()); removeErr != nil {
			t.Errorf("Failed to remove temp file: %v", err)
		}
	}()

	if _, writeErr := f.WriteString(content); writeErr != nil {
		t.Fatalf("写入临时文件失败: %v", writeErr)
	}
	if closeErr := f.Close(); closeErr != nil {
		t.Errorf("Failed to close temp file: %v", err)
	}

	config, err := loadConfig(f.Name())
	if err != nil {
		t.Fatalf("解析配置失败: %v", err)
	}

	// 验证已设置的配置
	if config.Build.OutputDir != "partial_dist" {
		t.Error("Build.OutputDir 解析错误")
	}
	if !config.Build.UseVendor {
		t.Error("Build.UseVendor 解析错误，预期true")
	}
	if !config.Install.Force {
		t.Error("Install.Force 解析错误，预期true")
	}
	if config.Env["CGO_ENABLED"] != "1" {
		t.Error("Env.CGO_ENABLED 解析错误")
	}

	// 验证未设置的配置使用默认值
	if config.Build.OutputName != globls.DefaultAppName {
		t.Errorf("Build.OutputName 应使用默认值 %s, 实际 %s", globls.DefaultAppName, config.Build.OutputName)
	}
	if config.Install.InstallPath != getDefaultInstallPath() {
		t.Errorf("Install.InstallPath 应使用默认值 %s, 实际 %s", getDefaultInstallPath(), config.Install.InstallPath)
	}
}

// TestLoadConfig_InvalidToml 测试加载无效格式的TOML文件
func TestLoadConfig_InvalidToml(t *testing.T) {
	// 创建包含无效TOML的文件
	content := `
[build
output_dir = "invalid"
`

	f := createTempFile(t, content)
	defer func() {
		if err := os.Remove(f.Name()); err != nil {
			t.Errorf("Failed to remove temp file: %v", err)
		}
	}()

	// 预期解析失败
	_, err := loadConfig(f.Name())
	if err == nil {
		t.Error("预期解析无效TOML时返回错误，但未返回错误")
	}
}

// TestConfigStruct_TagMatching 验证结构体字段的TOML标签是否正确定义
func TestConfigStruct_TagMatching(t *testing.T) {
	// 验证BuildConfig结构体标签
	buildFields := map[string]string{
		"OutputDir":           "output_dir",
		"OutputName":          "output_name",
		"MainFile":            "main_file",
		"Ldflags":             "ldflags",
		"UseVendor":           "use_vendor",
		"InjectGitInfo":       "inject_git_info",
		"SimpleName":          "simple_name",
		"Proxy":               "proxy",
		"EnableCgo":           "enable_cgo",
		"ColorOutput":         "color_output",
		"BatchMode":           "batch_mode",
		"CurrentPlatformOnly": "current_platform_only",
		"ZipOutput":           "zip_output",
	}

	verifyStructTags(t, BuildConfig{}, buildFields)

	// 验证InstallConfig结构体标签
	installFields := map[string]string{
		"Install":     "install",
		"InstallPath": "install_path",
		"Force":       "force",
	}

	verifyStructTags(t, InstallConfig{}, installFields)
}

// 辅助函数：创建临时文件
func createTempFile(t *testing.T, content string) *os.File {
	f, err := os.CreateTemp("", "config_test_*.toml")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}

	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("写入临时文件失败: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Errorf("Failed to close temp file: %v", err)
	}

	return f
}

// 辅助函数：验证结构体字段的TOML标签是否正确
func verifyStructTags(t *testing.T, s interface{}, expectedTags map[string]string) {
	val := reflect.ValueOf(s)
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldName := field.Name

		// 获取预期的TOML标签
		expectedTag, ok := expectedTags[fieldName]
		if !ok {
			t.Errorf("未预期的结构体字段: %s", fieldName)
			continue
		}

		// 获取实际的TOML标签
		actualTag := field.Tag.Get("toml")
		if actualTag != expectedTag {
			t.Errorf("字段 %s 的TOML标签不匹配，预期 '%s', 实际 '%s'", fieldName, expectedTag, actualTag)
		}
	}

	// 验证是否所有预期标签都已检查
	for fieldName := range expectedTags {
		found := false
		for i := 0; i < typ.NumField(); i++ {
			if typ.Field(i).Name == fieldName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("预期的结构体字段不存在: %s", fieldName)
		}
	}
}

// TestLoadConfig_EmptyFile 测试加载空配置文件（应返回默认配置）
func TestLoadConfig_EmptyFile(t *testing.T) {
	f := createTempFile(t, "")
	defer func() {
		if err := os.Remove(f.Name()); err != nil {
			t.Errorf("Failed to remove temp file: %v", err)
		}
	}()

	config, err := loadConfig(f.Name())
	if err != nil {
		t.Fatalf("解析空文件失败: %v", err)
	}

	// 验证使用默认配置
	if config.Build.OutputDir != globls.DefaultOutputDir {
		t.Errorf("空文件应使用默认OutputDir %s, 实际 %s", globls.DefaultOutputDir, config.Build.OutputDir)
	}
}

// TestConfig_EnvMerge 测试环境变量合并（配置文件中的环境变量应与默认值合并）
func TestConfig_EnvMerge(t *testing.T) {
	// 默认环境变量（如果有）
	// 注意：当前代码中默认Env为空，此测试验证合并逻辑
	content := `
[env]
GOOS = "windows"
`

	f := createTempFile(t, content)
	defer func() {
		if err := os.Remove(f.Name()); err != nil {
			t.Errorf("Failed to remove temp file: %v", err)
		}
	}()

	config, err := loadConfig(f.Name())
	if err != nil {
		t.Fatalf("解析配置失败: %v", err)
	}

	// 验证配置文件中的环境变量被正确加载
	if config.Env["GOOS"] != "windows" {
		t.Error("Env.GOOS 解析错误")
	}
}
