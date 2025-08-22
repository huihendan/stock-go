package main

import (
	"fmt"
	"os"
	"os/signal"
	"stock/globalConfig"
	"stock/logger"
	"stock/stockData"
	"stock/stockStrategy"
	"stock/utils"
	"syscall"
	"time"
)

func main() {

	executeTime := globalConfig.ExecuteAnalyseDataTime

	// 启动定时任务
	utils.DoWorkEveryDayOnce(doworkEveryDay, &executeTime)

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		logger.Infof("收到中断信号，正在优雅关闭...")

		// 等待定时任务停止，最多等待30分钟
		select {
		case <-time.After(30 * time.Minute):
			logger.Warnf("定时任务停止超时")
		}

		logger.Infof("程序退出")
	}
}

func doworkEveryDay() {

	analyseData()
}

func analyseData() error {
	logger.Info("analyseData start")

	stockData.ReLoadAllData()

	var lastErr error
	successCount := 0

	var stockList []string

	for _, stock := range stockData.Stocks {
		if stock == nil {
			logger.Warnf("股票数据为空，跳过")
			continue
		}

		if isHighPoint, dataStr := stockStrategy.HighPointStrategy(stock.Code); isHighPoint {
			logger.Infof("发现高点: %s - %s", stock.Code, dataStr)
			successCount++
			stockList = append(stockList, stock.Code)
		} else if dataStr != "" {
			logger.Infof("股票 %s 处理完成，未发现高点", stock.Code)
			successCount++
		} else {
			logger.Warnf("股票 %s 处理失败", stock.Code)
			lastErr = fmt.Errorf("股票 %s 处理失败", stock.Code)
		}
	}

	message := ""
	for _, stock := range stockList {
		message += fmt.Sprintf("%s\n", stock)
	}
	if message != "" {
		utils.SendWeChatMessage(fmt.Sprintf("发现高点:\n %s", message))
	}

	logger.Infof("analyseData end - 成功处理 %d 只股票", successCount)

	return lastErr
}
