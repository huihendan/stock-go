package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"stock/globalConfig"
	"stock/logger"
	"stock/stockData"
	"stock/stockStrategy"
	"stock/utils"
	"strings"
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
			logger.Warnf("stock数据为空，跳过")
			continue
		}

		if isHighPoint, dataStr := stockStrategy.HighPointStrategyLast(stock.Code); isHighPoint {
			logger.Infof("发现高点: %s - %s", stock.Code, dataStr)
			successCount++
			stockStr := stock.Code + " " +  stock.Name
			stockList = append(stockList, stockStr)
		} else if dataStr != "" {
			logger.Infof("stock %s 处理完成，未发现高点", stock.Code)
			successCount++
		} else {
			logger.Warnf("stock %s 处理失败", stock.Code)
			lastErr = fmt.Errorf("stock %s 处理失败", stock.Code)
		}
	}

	logger.Infof("找到高点stock数量: %d", len(stockList))
	message := ""
	for _, stock := range stockList {
		message += fmt.Sprintf("%s\n", stock)
	}
	if message != "" {
		// 获取公网IP
		publicIP := getPublicIP()
		finalMessage := fmt.Sprintf("公网IP: %s\n发现高点:\n %s", publicIP, message)
		logger.Infof("发送微信消息长度: %d 字节", len(finalMessage))
		logger.Infof("完整消息内容: %s", finalMessage)
		utils.SendWeChatMessage(finalMessage)
	}

	logger.Infof("analyseData end - 成功处理 %d 只stock", successCount)

	return lastErr
}

func getPublicIP() string {
	cmd := exec.Command("curl", "ifconfig.me")
	output, err := cmd.Output()
	if err != nil {
		logger.Warnf("获取公网IP失败: %v", err)
		return "未知"
	}
	return strings.TrimSpace(string(output))
}
