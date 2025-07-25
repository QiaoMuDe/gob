// Package qflag 根包统一导出入口
// 本文件用于将各子包的核心功能导出到根包，简化外部使用
package qflag

import (
	"gitee.com/MM-Q/qflag/cmd"
	"gitee.com/MM-Q/qflag/flags"
)

/*
项目地址: https://gitee.com/MM-Q/qflag
*/

// 导出子包类型和函数 //

// QCommandLine 导出cmd包的全局默认Command实例
var QCommandLine = cmd.QCommandLine

// cmd 导出cmd包中的Cmd结构体
type Cmd = cmd.Cmd

// NewCmd 导出cmd包中的NewCmd函数
var NewCmd = cmd.NewCmd

// ExampleInfo 导出cmd包中的ExampleInfo结构体
type ExampleInfo = cmd.ExampleInfo

// 导出标志类型 //

// Flag 导出flag包中的Flag结构体
type Flag = flags.Flag

// StringFlag 导出flag包中的StringFlag结构体
type StringFlag = flags.StringFlag

// IntFlag 导出flag包中的IntFlag结构体
type IntFlag = flags.IntFlag

// BoolFlag 导出flag包中的BoolFlag结构体
type BoolFlag = flags.BoolFlag

// DurationFlag 导出flag包中的DurationFlag结构体
type DurationFlag = flags.DurationFlag

// Float64Flag 导出flag包中的Float64Flag结构体
type Float64Flag = flags.Float64Flag

// Int64Flag 导出flag包中的Int64Flag结构体
type Int64Flag = flags.Int64Flag

// SliceFlag 导出flag包中的SliceFlag结构体
type SliceFlag = flags.SliceFlag

// EnumFlag 导出flag包中的EnumFlag结构体
type EnumFlag = flags.EnumFlag

// MapFlag 导出flag包中的MapFlag结构体
type MapFlag = flags.MapFlag

// TimeFlag 导出flag包中的TimeFlag结构体
type TimeFlag = flags.TimeFlag

// PathFlag 导出flag包中的PathFlag结构体
type PathFlag = flags.PathFlag

// Uint16Flag 导出flag包中的UintFlag结构体
type Uint16Flag = flags.Uint16Flag

// Uint32Flag 导出flag包中的Uint32Flag结构体
type Uint32Flag = flags.Uint32Flag

// Uint64Flag 导出flag包中的Uint64Flag结构体
type Uint64Flag = flags.Uint64Flag

// IP4Flag 导出flag包中的Ip4Flag结构体
type IP4Flag = flags.IP4Flag

// IP6Flag 导出flag包中的Ip6Flag结构体
type IP6Flag = flags.IP6Flag

// URLFlag 导出flag包中的URLFlag结构体
type URLFlag = flags.URLFlag
