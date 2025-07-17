package cmd

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gitee.com/MM-Q/gob/internal/globls"
	"gitee.com/MM-Q/verman"
)

// runCmd 执行指定系统命令，仅使用指定的环境变量
//
// 参数：
//   - args: 命令行参数切片，args[0] 为命令本身
//   - env: 完整的环境变量切片，形如 "KEY=VALUE"；传 nil 或空切片表示不额外设置
//
// 返回：
//   - result: 标准输出与标准错误合并后的内容
//   - err: 命令执行期间的任何错误
func runCmd(args []string, env []string) ([]byte, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("empty command")
	}
	cmd := exec.Command(args[0], args[1:]...)
	if len(env) > 0 {
		cmd.Env = env // 直接覆盖，不再继承系统环境
	}
	return cmd.CombinedOutput()
}

// genOutputName 生成输出文件名
//
// 参数：
//   - appName: 应用名
//   - useSimpleName: 是否使用简单的文件名
//   - version: 版本号
//   - sysPlatform: 系统平台
//   - sysArch: 系统架构
//
// 返回：
//   - 生成的输出文件名
//
// 注意：
//   - 简单模式：示例, `myapp`
//   - 完整模式：示例, `myapp_linux_amd64_1.0.0`
func genOutputName(appName string, useSimpleName bool, version string, sysPlatform string, sysArch string) string {
	if useSimpleName && batchFlag.Get() {
		globls.CL.PrintWarn("使用批量构建时, 简单模式将失效")
	}

	// 简单模式，不添加平台和版本信息
	if useSimpleName && !batchFlag.Get() {
		switch sysPlatform {
		case "windows":
			return fmt.Sprint(strings.TrimSuffix(appName, ".exe"), ".exe")
		case "darwin":
			return fmt.Sprint(strings.TrimSuffix(appName, ".app"), ".app")
		default:
			return appName
		}
	}

	// 完整模式，添加平台和版本信息
	switch sysPlatform {
	case "windows":
		return fmt.Sprintf("%s_%s_%s_%s.exe", appName, sysPlatform, sysArch, version)
	case "darwin":
		return fmt.Sprintf("%s_%s_%s_%s.app", appName, sysPlatform, sysArch, version)
	default:
		return fmt.Sprintf("%s_%s_%s_%s", appName, sysPlatform, sysArch, version)
	}
}

// checkBaseEnv 检查基础环境以及格式化和静态检查
func checkBaseEnv() error {
	// 检查go环境
	if _, err := runCmd([]string{"go", "version"}, os.Environ()); err != nil {
		return fmt.Errorf("未找到go环境, 请先安装go环境: %w", err)
	}

	// 检查当前目录下是否存在go.mod
	if _, statErr := os.Stat("go.mod"); os.IsNotExist(statErr) {
		return fmt.Errorf("当前目录下不存在go.mod文件, 请先初始化go.mod文件, 或前往项目根目录执行: %w", statErr)
	}

	// 检查指定的入口文件是否存在
	if _, statErr := os.Stat(mainFlag.Get()); os.IsNotExist(statErr) {
		return fmt.Errorf("入口文件不存在: %w", statErr)
	}

	// 如果启用vendor模式，检查vendor目录是否存在
	if vendorFlag.Get() {
		if _, statErr := os.Stat("vendor"); os.IsNotExist(statErr) {
			return fmt.Errorf("当前路径下不存在vendor目录, 请先执行 go mod vendor 命令生成vendor目录: %w", statErr)
		}
	}

	// 定义用于判断选择检查模式的变量
	var checkMode bool

	// 检查系统中是否存在golangci-lint否则执行默认的处理命令
	if _, err := runCmd([]string{"golangci-lint", "version"}, os.Environ()); err != nil {
		checkMode = true
	}

	// 根据checkMode的值执行不同的处理命令
	var cmds []globls.CommandGroup
	if checkMode {
		cmds = append(cmds, globls.DefaultCheckCmds...)
	} else {
		cmds = append(cmds, globls.GolangciLintCheckCmds...)
	}

	// 获取环境变量
	env := os.Environ()

	// 设置Go代理
	env = append(env, fmt.Sprintf("GOPROXY=%s", proxyFlag.Get()))

	// 遍历处理命令组
	for _, cmdGroup := range cmds {
		if result, runErr := runCmd(cmdGroup.Cmds, env); runErr != nil {
			return fmt.Errorf("执行 %s 失败: \n%s \n%w", cmdGroup.Cmds, string(result), runErr)
		}
	}

	// 检查输出目录是否存在，不存在则创建
	if _, err := os.Stat(outputFlag.Get()); os.IsNotExist(err) {
		if err := os.MkdirAll(outputFlag.Get(), os.ModePerm); err != nil {
			return fmt.Errorf("创建输出目录失败: %w", err)
		}
	}

	return nil
}

// getGitMetaData 获取git元数据
//
// 参数：
//   - v: verman.VerMan 结构体指针，用于存储获取到的git元数据
//
// 返回值：
//   - error: 错误信息，如果获取成功则返回nil
func getGitMetaData(v *verman.VerMan) error {
	// 定义命令和对应字段的映射
	commands := []struct {
		cmd   globls.CommandGroup
		field *string
	}{
		{globls.GitVersionCmd, &v.GitVersion},
		{globls.GitCommitHashCmd, &v.GitCommit},
		{globls.GitCommitTimeCmd, &v.GitCommitTime},
	}

	// 处理常规git信息
	for _, item := range commands {
		result, err := runCmd(item.cmd.Cmds, os.Environ())
		if err != nil {
			return fmt.Errorf("%s: \n\t%s \n%w", item.cmd.Name, string(result), err)
		}
		// 设置字段值，并去除首尾空格
		*item.field = strings.TrimSpace(string(result))
	}

	// 特殊处理git树状态
	result, err := runCmd(globls.GitTreeStatusCmd.Cmds, os.Environ())
	if err != nil {
		return fmt.Errorf("%s: \n\t%s \n%w", globls.GitTreeStatusCmd.Name, string(result), err)
	}

	// 根据git树状态设置GitTreeState字段
	if strings.TrimSpace(string(result)) == "" {
		v.GitTreeState = "clean"
	} else {
		v.GitTreeState = "dirty"
	}

	// 设置appName字段
	v.AppName = nameFlag.Get()

	return nil
}

// createZip 函数用于创建ZIP压缩文件
//
// 参数:
//   - zipFilePath: 生成的ZIP文件路径
//   - sourceDir: 需要压缩的源目录路径
//
// 返回值:
//   - error: 操作过程中遇到的错误
func createZip(zipFilePath string, sourceDir string) error {
	// 检查zipFilePath是否为绝对路径，如果不是，将其转换为绝对路径
	if !filepath.IsAbs(zipFilePath) {
		absPath, err := filepath.Abs(zipFilePath)
		if err != nil {
			return fmt.Errorf("转换zipFilePath为绝对路径失败: %w", err)
		}
		zipFilePath = absPath
	}

	// 检查sourceDir是否为绝对路径，如果不是，将其转换为绝对路径
	if !filepath.IsAbs(sourceDir) {
		absPath, err := filepath.Abs(sourceDir)
		if err != nil {
			return fmt.Errorf("转换sourceDir为绝对路径失败: %w", err)
		}
		sourceDir = absPath
	}

	// 创建 ZIP 文件
	zipFile, err := os.Create(zipFilePath)
	if err != nil {
		return fmt.Errorf("创建 ZIP 文件失败: %w", err)
	}
	defer func() { _ = zipFile.Close() }()

	// 创建 ZIP 写入器
	zipWriter := zip.NewWriter(zipFile)
	defer func() { _ = zipWriter.Close() }()

	// 遍历目录并添加文件到 ZIP 包
	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("遍历目录时出错: %w", err)
		}

		// 获取相对路径，保留顶层目录
		headerName, err := filepath.Rel(filepath.Dir(sourceDir), path)
		if err != nil {
			return fmt.Errorf("获取相对路径失败: %w", err)
		}

		// 替换路径分隔符为正斜杠（ZIP 文件格式要求）
		headerName = filepath.ToSlash(headerName)

		// 获取文件的详细状态
		fileStat, err := os.Lstat(path)
		if err != nil {
			return fmt.Errorf("获取文件状态失败: %w", err)
		}

		// 根据文件类型处理
		switch mode := fileStat.Mode(); {
		case mode.IsRegular():
			// 普通文件
			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return fmt.Errorf("创建 ZIP 文件头失败: %w", err)
			}
			// 设置文件头的名称
			header.Name = headerName

			// 设置压缩方法为 Deflate
			header.Method = zip.Deflate

			// 创建 ZIP 写入器
			fileWriter, err := zipWriter.CreateHeader(header)
			if err != nil {
				return fmt.Errorf("创建 ZIP 写入器失败: %w", err)
			}

			// 打开文件
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("打开文件失败: %w", err)
			}
			defer func() { _ = file.Close() }()

			// 获取文件大小
			fileInfo, err := file.Stat()
			if err != nil {
				return fmt.Errorf("获取文件信息失败: %w", err)
			}
			fileSize := fileInfo.Size()

			// 根据文件大小设置缓冲区大小
			bufferSize := getBufferSize(fileSize)

			// 创建带缓冲的读取器
			bufferedReader := bufio.NewReaderSize(file, bufferSize)

			// 使用缓冲区进行文件复制，提高性能
			buffer := make([]byte, bufferSize) // 动态分配缓冲区大小
			if _, err := io.CopyBuffer(fileWriter, bufferedReader, buffer); err != nil {
				return fmt.Errorf("写入 ZIP 文件失败: %w", err)
			}

		case mode.IsDir():
			// 目录
			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return fmt.Errorf("创建 ZIP 文件头失败: %w", err)
			}
			// 设置目录的名称，末尾添加斜杠
			header.Name = headerName + "/"

			// 设置压缩方法为 Store（不压缩）
			header.Method = zip.Store

			// 创建目录
			if _, err := zipWriter.CreateHeader(header); err != nil {
				return fmt.Errorf("创建 ZIP 目录失败: %w", err)
			}

		case mode&os.ModeSymlink != 0:
			// 软链接
			target, err := os.Readlink(path)
			if err != nil {
				return fmt.Errorf("读取软链接目标失败: %w", err)
			}

			// 创建软链接文件头
			header := &zip.FileHeader{
				Name:   headerName,
				Method: zip.Store,
			}
			// 设置软链接的元数据
			header.SetMode(mode)

			// 创建软链接
			writer, err := zipWriter.CreateHeader(header)
			if err != nil {
				return fmt.Errorf("创建 ZIP 软链接失败: %w", err)
			}
			if _, err := writer.Write([]byte(target)); err != nil {
				return fmt.Errorf("写入软链接目标失败: %w", err)
			}

		case mode&os.ModeDevice != 0:
			// 设备文件
			header := &zip.FileHeader{
				Name:   headerName,
				Method: zip.Store,
			}
			// 设置设备文件的元数据
			header.SetMode(mode)

			// 创建设备文件
			writer, err := zipWriter.CreateHeader(header)
			if err != nil {
				return fmt.Errorf("创建 ZIP 设备文件失败: %w", err)
			}
			// 设备文件通常不包含数据，只记录其元数据
			if _, err := writer.Write([]byte{}); err != nil {
				return fmt.Errorf("写入设备文件失败: %w", err)
			}

		default:
			// 其他特殊文件类型
			header := &zip.FileHeader{
				Name:   headerName,
				Method: zip.Store,
			}
			// 设置特殊文件的元数据
			header.SetMode(mode)

			// 创建特殊文件
			writer, err := zipWriter.CreateHeader(header)
			if err != nil {
				return fmt.Errorf("创建 ZIP 特殊文件失败: %w", err)
			}
			// 特殊文件通常不包含数据，只记录其元数据
			if _, err := writer.Write([]byte{}); err != nil {
				return fmt.Errorf("写入特殊文件失败: %w", err)
			}
		}

		return nil
	})

	// 检查是否有错误发生
	if err != nil {
		return fmt.Errorf("打包目录到 ZIP 失败: %w", err)
	}

	return nil
}

// getBufferSize 根据文件大小动态设置缓冲区大小。该函数会根据传入的文件大小，
// 选择合适的缓冲区大小，以优化文件读写操作的性能。不同的文件大小范围对应不同的缓冲区大小。
//
// 参数:
//   - fileSize: 文件的大小，单位为字节，类型为 int64。
//
// 返回值:
//   - 缓冲区的大小，单位为字节，类型为 int。
func getBufferSize(fileSize int64) int {
	switch {
	// 当文件大小小于 512KB 时，设置缓冲区大小为 32KB
	case fileSize < 512*1024:
		return 32 * 1024
	// 当文件大小小于 1MB 时，设置缓冲区大小为 64KB
	case fileSize < 1*1024*1024:
		return 64 * 1024
	// 当文件大小小于 5MB 时，设置缓冲区大小为 128KB
	case fileSize < 5*1024*1024:
		return 128 * 1024
	// 当文件大小小于 10MB 时，设置缓冲区大小为 256KB
	case fileSize < 10*1024*1024:
		return 256 * 1024
	// 当文件大小小于 100MB 时，设置缓冲区大小为 512KB
	case fileSize < 100*1024*1024:
		return 512 * 1024
	// 当文件大小大于等于 100MB 时，设置缓冲区大小为 1MB
	default:
		return 1024 * 1024
	}
}
