package utils

import (
	"fmt"
	"log/slog"
	"stock-go/logger"
	"time"
)

func CostTime(start time.Time) {
	elapsed := time.Since(start)
	//fmt.Println("cost:%v",elapsed)
	slog.Info("cost time", "elapsed", elapsed)
}

// DoWorkEveryDayOnce 每天在指定时间执行一次任务
// executeTime 参数格式为 "HH:MM"，如 "19:00"，当为 nil 时默认为 "19:00"
func DoWorkEveryDayOnce(f func(), executeTime *string) bool {
	// 设置默认执行时间
	defaultTime := "19:00"
	targetTime := defaultTime
	if executeTime != nil {
		targetTime = *executeTime
	}

	logger.Infof("启动定时任务，每隔5分钟检查一次是否为工作日且在%s以后", targetTime)

	// 记录已执行过的日期
	var lastExecutedDate string

	// 启动定时器
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		// 第一次检查，不需要等待
		checkAndExecute := func() {
			now := time.Now()
			currentDate := now.Format("2006-01-02")

			// 检查是否为工作日并且在指定时间以后
			if isWorkdayAndAfterTime(now, targetTime) {
				// 检查今天是否已经执行过
				if lastExecutedDate != currentDate {
					logger.Infof("检测到工作日且时间在%s后，开始执行数据更新任务", targetTime)

					f()

					// 标记今天已执行
					lastExecutedDate = currentDate

				} else {
					logger.Infof("今天已经执行过数据更新任务，跳过执行")
				}
			} else {
				logger.Infof("当前为周末或时间不在%s后，跳过任务执行", targetTime)
			}
		}

		// 立即执行第一次检查
		checkAndExecute()

		for {
			select {
			case <-ticker.C:
				checkAndExecute()
			}
		}
	}()

	return true
}

// isWorkdayAndAfterTime 检查是否为工作日且时间在指定时间之后
// timeStr 格式为 "HH:MM"，如 "19:00"
func isWorkdayAndAfterTime(t time.Time, timeStr string) bool {
	// 检查是否为工作日（周一至周五）
	//weekday := t.Weekday()
	// isWorkday := weekday >= time.Monday && weekday <= time.Friday
	isWorkday := true

	// 解析目标时间
	targetHour, targetMinute := 19, 0 // 默认值
	_, err := fmt.Sscanf(timeStr, "%d:%d", &targetHour, &targetMinute)
	if err != nil {
		logger.Errorf("解析时间格式错误: %v，使用默认值19:00", err)
		targetHour, targetMinute = 19, 0
	}

	// 检查当前时间是否在目标时间之后
	currentHour := t.Hour()
	currentMinute := t.Minute()

	isAfterTargetTime := false
	if currentHour > targetHour {
		isAfterTargetTime = true
	} else if currentHour == targetHour && currentMinute >= targetMinute {
		isAfterTargetTime = true
	}

	return isWorkday && isAfterTargetTime
}

// isWorkdayAndAfter19 检查是否为工作日且时间在19点之后
func isWorkdayAndAfter19(t time.Time) bool {
	return isWorkdayAndAfterTime(t, "19:00")
}
