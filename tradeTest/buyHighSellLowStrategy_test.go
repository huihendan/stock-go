package tradeTest

import (
	"fmt"
	"stock/logger"
	"stock/stockStrategy"
	"strings"
	"testing"
)

func TestBuyHighSellLowStrategy(t *testing.T) {
	logger.Info("TestBuyHighSellLowStrategy start")

	// 创建自定义配置的策略实例
	strategy := stockStrategy.NewBuyHighSellLowStrategyWithConfig(
		300,  // 回看250天
		0.08, // 止损8%
		15,   // 最大持有20天
	)

	// 汇总统计变量
	var (
		totalStocks        = 0
		totalAllTrades     = 0
		totalAllCompleted  = 0
		totalAllProfit     = 0.0
		totalAllWinCount   = 0
		totalAllProfitRate = 0.0
		stocksWithTrades   = 0
		totalAllInvestment = 0.0 // 总投入资金
	)

	// 测试几个股票
	testCodes := strategy.DealSelectStockCodes()
	for _, code := range testCodes {
		totalStocks++
		logger.Infof("测试股票: %s", code)

		// 执行策略
		operates := strategy.DealStrategy(code)

		// 统计结果
		totalTrades := len(operates)
		completedTrades := 0
		totalProfit := 0.0
		winCount := 0
		totalInvestment := 0.0 // 该股票的总投入（买入价格 × 股数）

		for _, record := range operates {
			if record.Status == 2 { // 已完成交易
				completedTrades++
				totalProfit += record.Profit
				totalInvestment += record.BuyOperate.BuyPrice * record.StockNum
				if record.Profit > 0 {
					winCount++
				}

				// 单笔交易收益率 = (卖出价 - 买入价) / 买入价 × 100%
				tradeProfitRate := (record.SellOperate.SellPrice - record.BuyOperate.BuyPrice) / record.BuyOperate.BuyPrice * 100

				logger.Infof("  交易: 买入日期=%s 价格=%.2f, 卖出日期=%s 价格=%.2f, 收益=%.2f(%.2f%%)",
					record.BuyOperate.OperateDate,
					record.BuyOperate.BuyPrice,
					record.SellOperate.OperateDate,
					record.SellOperate.SellPrice,
					record.Profit,
					tradeProfitRate)
			}
		}

		// 输出单个股票统计
		if completedTrades > 0 {
			winRate := float64(winCount) / float64(completedTrades) * 100
			avgProfit := totalProfit / float64(completedTrades)
			profitRate := (totalProfit / totalInvestment) * 100 // 收益率 = 总收益 / 总投入

			logger.Infof("股票 %s 统计:", code)
			logger.Infof("  总交易次数: %d", totalTrades)
			logger.Infof("  完成交易次数: %d", completedTrades)
			logger.Infof("  总收益: %.2f", totalProfit)
			logger.Infof("  平均收益: %.2f", avgProfit)
			logger.Infof("  总收益率: %.2f%%", profitRate)
			logger.Infof("  胜率: %.2f%%", winRate)

			// 累加到汇总数据
			stocksWithTrades++
			totalAllTrades += totalTrades
			totalAllCompleted += completedTrades
			totalAllProfit += totalProfit
			totalAllWinCount += winCount
			totalAllProfitRate += profitRate
			totalAllInvestment += totalInvestment
		} else {
			logger.Infof("股票 %s 没有完成的交易", code)
		}

		fmt.Println()
	}

	// 打印汇总统计
	logger.Info(strings.Repeat("=", 60))
	logger.Info("所有股票汇总统计")
	logger.Info(strings.Repeat("=", 60))
	logger.Infof("测试股票总数: %d", totalStocks)
	logger.Infof("有交易的股票数: %d", stocksWithTrades)
	logger.Infof("总交易次数: %d", totalAllTrades)
	logger.Infof("总完成交易次数: %d", totalAllCompleted)

	if totalAllCompleted > 0 {
		overallWinRate := float64(totalAllWinCount) / float64(totalAllCompleted) * 100
		overallAvgProfit := totalAllProfit / float64(totalAllCompleted)
		overallProfitRate := (totalAllProfit / totalAllInvestment) * 100
		avgStockProfitRate := totalAllProfitRate / float64(stocksWithTrades)

		logger.Infof("总收益: %.2f", totalAllProfit)
		logger.Infof("平均每笔收益: %.2f", overallAvgProfit)
		logger.Infof("总收益率: %.2f%%", overallProfitRate)
		logger.Infof("平均股票收益率: %.2f%%", avgStockProfitRate)
		logger.Infof("总胜率: %.2f%%", overallWinRate)
		logger.Infof("平均每股票完成交易次数: %.2f", float64(totalAllCompleted)/float64(stocksWithTrades))
	} else {
		logger.Info("没有完成的交易")
	}
	logger.Info(strings.Repeat("=", 60))

	logger.Info("TestBuyHighSellLowStrategy end")
}
