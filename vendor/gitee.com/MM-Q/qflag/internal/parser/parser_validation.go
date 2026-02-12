package parser

import (
	"fmt"

	"gitee.com/MM-Q/qflag/internal/types"
)

// buildSetFlagsMap 构建已设置标志的映射
//
// 参数:
//   - cmd: 要验证的命令
//
// 返回值:
//   - map[string]bool: 已设置标志名称的映射，key为标志名，value为true
//
// 功能说明:
//   - 遍历所有标志，收集已设置的标志
//   - 构建映射以支持快速查询
//   - 避免在验证过程中重复调用 GetFlag() 和 IsSet()
//   - 将结果缓存到解析器的 setFlagsMap 字段中
//   - 根据标志数量预分配map空间，减少扩容开销
func (p *DefaultParser) buildSetFlagsMap(cmd types.Command) map[string]bool {
	// 根据标志数量预分配map空间，减少扩容开销
	flags := cmd.Flags()
	p.setFlagsMap = make(map[string]bool, len(flags))

	for _, flag := range flags {
		if flag.IsSet() {
			p.setFlagsMap[flag.Name()] = true
		}
	}

	return p.setFlagsMap
}

// validateMutexGroups 验证命令的互斥组规则
//
// 参数:
//   - config: 命令配置
//
// 返回值:
//   - error: 如果互斥组验证失败返回错误
//
// 功能说明:
//   - 检查每个互斥组中是否有多个标志被设置
//   - 检查不允许为空的互斥组中是否有至少一个标志被设置
//   - 提供清晰的错误信息, 指出冲突的标志和组名
//
// 验证规则:
//   - 互斥组中最多只能有一个标志被设置
//   - 如果 AllowNone 为 false, 则必须至少有一个标志被设置
//
// 错误处理:
//   - 使用 types.NewError 创建结构化错误
//   - 错误信息包含互斥组名称和冲突的标志列表
//
// 性能优化:
//   - 使用缓存的已设置标志映射，避免重复的 GetFlag() 和 IsSet() 调用
func (p *DefaultParser) validateMutexGroups(config *types.CmdConfig) error {
	// 检查互斥组是否为空
	if len(config.MutexGroups) == 0 {
		return nil
	}

	// 使用缓存的已设置标志映射
	setFlags := p.setFlagsMap

	// 遍历所有互斥组
	for _, group := range config.MutexGroups {
		setCount := 0
		var setFlagsList []string

		// 检查互斥组中的每个标志是否被设置
		for _, flagName := range group.Flags {
			if setFlags[flagName] {
				setCount++
				setFlagsList = append(setFlagsList, flagName)
			}
		}

		// 如果互斥组中设置了多个标志, 返回错误
		if setCount > 1 {
			return types.NewError("MUTEX_GROUP_VIOLATION",
				fmt.Sprintf("mutually exclusive flags %v in group '%s' cannot be used together", setFlagsList, group.Name),
				nil)
		}

		// 如果不允许为空, 且互斥组中没有设置任何标志, 返回错误
		if !group.AllowNone && setCount == 0 {
			return types.NewError("MUTEX_GROUP_REQUIRED",
				fmt.Sprintf("one of the mutually exclusive flags %v in group '%s' must be set", group.Flags, group.Name),
				nil)
		}
	}

	return nil
}

// validateRequiredGroups 验证命令的必需组规则
//
// 参数:
//   - config: 命令配置
//
// 返回值:
//   - error: 如果必需组验证失败返回错误
//
// 功能说明:
//   - 检查每个必需组中是否有标志未被设置
//   - 提供清晰的错误信息，指出未设置的标志和组名
//
// 验证规则:
//   - 必需组中的所有标志都必须被设置
//   - 如果有任何一个标志未被设置，返回错误
//
// 错误处理:
//   - 使用 types.NewError 创建结构化错误
//   - 错误信息包含必需组名称和未设置的标志列表
//
// 性能优化:
//   - 使用缓存的已设置标志映射，避免重复的 GetFlag() 和 IsSet() 调用
func (p *DefaultParser) validateRequiredGroups(config *types.CmdConfig) error {
	if len(config.RequiredGroups) == 0 {
		return nil
	}

	// 使用缓存的已设置标志映射
	setFlags := p.setFlagsMap

	// 遍历所有必需组
	for _, group := range config.RequiredGroups {
		var unsetFlags []string

		// 遍历组中的每个标志
		for _, flagName := range group.Flags {
			if !setFlags[flagName] {
				unsetFlags = append(unsetFlags, flagName)
			}
		}

		// 如果组中有未设置的标志, 返回错误
		if len(unsetFlags) > 0 {
			return types.NewError("REQUIRED_GROUP_VIOLATION",
				fmt.Sprintf("required flags %v in group '%s' must be set", unsetFlags, group.Name),
				nil)
		}
	}

	return nil
}
