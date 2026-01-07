package types

import (
	"gitee.com/MM-Q/verman"
)

// BuildContext 构建上下文, 封装构建所需的所有参数
type BuildContext struct {
	VerMan      *verman.Info // verman对象
	Env         []string     // 环境变量
	SysPlatform string       // 系统平台
	SysArch     string       // 系统架构
	Config      *GobConfig   // 配置对象
}
