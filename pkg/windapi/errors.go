package windapi

import (
	"errors"
	"fmt"
)

var errMap = map[int32]error{
	-40520001: errors.New("wind: 未知错误(-40520001)"),
	-40520002: errors.New("wind: 内部错误(-40520002)"),
	-40520003: errors.New("wind: 系统错误(-40520003)"),
	-40520004: errors.New("wind: 登录失败(-40520004)"),
	-40520005: errors.New("wind: 无权限(-40520005)"),
	-40520006: errors.New("wind: 用户取消(-40520006)"),
	-40520007: errors.New("wind: 无数据(-40520007)"),
	-40520008: errors.New("wind: 超时错误(-40520008)"),
	-40520009: errors.New("wind: 本地WBOX错误(-40520009)"),
	-40520010: errors.New("wind: 需要内容不存在(-40520010)"),
	-40520011: errors.New("wind: 需要服务器不存在(-40520011)"),
	-40520012: errors.New("wind: 引用不存在(-40520012)"),
	-40520013: errors.New("wind: 其他地方登录错误(-40520013)"),
	-40520014: errors.New("wind: 未登录使用WIM工具，故无法登录(-40520014)"),
	-40520015: errors.New("wind: 连续登录失败次数过多(-40520015)"),
	-40521001: errors.New("wind: IO操作错误(-40521001)"),
	-40521002: errors.New("wind: 后台服务器不可用(-40521002)"),
	-40521003: errors.New("wind: 网络连接失败(-40521003)"),
	-40521004: errors.New("wind: 请求发送失败(-40521004)"),
	-40521005: errors.New("wind: 数据接收失败(-40521005)"),
	-40521006: errors.New("wind: 网络错误(-40521006)"),
	-40521007: errors.New("wind: 服务器拒绝请求(-40521007)"),
	-40521008: errors.New("wind: 错误的应答(-40521008)"),
	-40521009: errors.New("wind: 数据解码失败(-40521009)"),
	-40521010: errors.New("wind: 网络超时(-40521010)"),
	-40521011: errors.New("wind: 频繁访问(-40521011)"),
	-40522001: errors.New("wind: 无合法会话(-40522001)"),
	-40522002: errors.New("wind: 非法数据服务(-40522002)"),
	-40522003: errors.New("wind: 非法请求(-40522003)"),
	-40522004: errors.New("wind: 万得代码语法错误(-40522004)"),
	-40522005: errors.New("wind: 不支持的万得代码(-40522005)"),
	-40522006: errors.New("wind: 指标语法错误(-40522006)"),
	-40522007: errors.New("wind: 不支持的指标(-40522007)"),
	-40522008: errors.New("wind: 指标参数语法错误(-40522008)"),
	-40522009: errors.New("wind: 不支持的指标参数(-40522009)"),
	-40522010: errors.New("wind: 日期与时间语法错误(-40522010)"),
	-40522011: errors.New("wind: 不支持的日期与时间(-40522011)"),
	-40522012: errors.New("wind: 不支持的请求参数(-40522012)"),
	-40522013: errors.New("wind: 数组下标越界(-40522013)"),
	-40522014: errors.New("wind: 重复的WQID(-40522014)"),
	-40522015: errors.New("wind: 请求无相应权限(-40522015)"),
	-40522016: errors.New("wind: 不支持的数据类型(-40522016)"),
	-40522017: errors.New("wind: 数据提取量超限(-40522017)"),
}

func parseErr(errCode int32) error {
	if 0 == errCode {
		return nil
	}
	if err, ok := errMap[errCode]; ok {
		return err
	}
	return fmt.Errorf("wind: unknown error(%d)", errCode)
}
