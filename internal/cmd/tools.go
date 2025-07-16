package cmd

import (
	"fmt"
	"os/exec"
)

// RunCmd 执行指定系统命令，仅使用指定的环境变量
//
// 参数：
//   - args: 命令行参数切片，args[0] 为命令本身
//   - env: 完整的环境变量切片，形如 "KEY=VALUE"；传 nil 或空切片表示不额外设置
//
// 返回：
//   - result: 标准输出与标准错误合并后的内容
//   - err: 命令执行期间的任何错误
func RunCmd(args []string, env []string) ([]byte, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("empty command")
	}
	cmd := exec.Command(args[0], args[1:]...)
	if len(env) > 0 {
		cmd.Env = env // 直接覆盖，不再继承系统环境
	}
	return cmd.CombinedOutput()
}
