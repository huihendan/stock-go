package utils

import (
	"fmt"
	"stock-go/logger"
)

// ExampleDoWorkEveryDayOnce 展示如何使用 DoWorkEveryDayOnce 函数
func ExampleDoWorkEveryDayOnce() {
	// 示例1：使用默认时间（19:00）
	DoWorkEveryDayOnce(func() {
		logger.Infof("这是一个使用默认时间的任务")
	}, nil)

	// 示例2：指定时间为 15:30
	customTime := "15:30"
	DoWorkEveryDayOnce(func() {
		logger.Infof("这是一个在 15:30 之后执行的任务")
	}, &customTime)

	// 防止程序退出
	fmt.Println("定时任务已启动，按 Ctrl+C 退出")
	select {}
}
