package stockStrategy

import (
	"fmt"
	"stock/logger"
	"stock/stockData"
	"testing"
)

func TestBuyHighSellLowStrategy(t *testing.T) {
	logger.Info("TestBuyHighSellLowStrategy start")

	// 加载股票列表和数据
	stockData.LoadPreStockList()
	stockData.LoadDataOneByOne()

	// 创建策略实例
	strategy := NewBuyHighSellLowStrategy()

	// 测试几个股票
	testCodes := []string{"000001", "000002", "600000"}

	for _, code := range testCodes {
		logger.Infof("测试股票: %s", code)

		// 执行策略
		operates := strategy.DealStrategy(code)

		// 统计结果
		totalTrades := len(operates)
		completedTrades := 0
		totalProfit := 0.0
		winCount := 0

		for _, record := range operates {
			if record.Status == 2 { // 已完成交易
				completedTrades++
				totalProfit += record.Profit
				if record.Profit > 0 {
					winCount++
				}

				logger.Infof("  交易: 买入日期=%s 价格=%.2f, 卖出日期=%s 价格=%.2f, 收益=%.2f",
					record.BuyOperate.OperateDate,
					record.BuyOperate.BuyPrice,
					record.SellOperate.OperateDate,
					record.SellOperate.SellPrice,
					record.Profit)
			}
		}

		// 输出统计
		if completedTrades > 0 {
			winRate := float64(winCount) / float64(completedTrades) * 100
			avgProfit := totalProfit / float64(completedTrades)

			logger.Infof("股票 %s 统计:", code)
			logger.Infof("  总交易次数: %d", totalTrades)
			logger.Infof("  完成交易次数: %d", completedTrades)
			logger.Infof("  总收益: %.2f", totalProfit)
			logger.Infof("  平均收益: %.2f", avgProfit)
			logger.Infof("  胜率: %.2f%%", winRate)
		} else {
			logger.Infof("股票 %s 没有完成的交易", code)
		}

		fmt.Println()
	}

	logger.Info("TestBuyHighSellLowStrategy end")
}

func TestBuyHighSellLowStrategyCustomConfig(t *testing.T) {
	logger.Info("TestBuyHighSellLowStrategyCustomConfig start")

	// 加载股票列表和数据
	stockData.LoadPreStockList()
	stockData.LoadDataOneByOne()

	// 创建自定义配置的策略实例
	strategy := NewBuyHighSellLowStrategyWithConfig(
		250,  // 回看250天
		0.08, // 止损8%
		20,   // 最大持有20天
	)

	code := "000001"
	logger.Infof("测试自定义配置策略，股票: %s", code)

	// 执行策略
	operates := strategy.DealStrategy(code)

	// 统计结果
	completedTrades := 0
	totalProfit := 0.0

	for _, record := range operates {
		if record.Status == 2 { // 已完成交易
			completedTrades++
			totalProfit += record.Profit
		}
	}

	logger.Infof("自定义配置结果:")
	logger.Infof("  完成交易次数: %d", completedTrades)
	logger.Infof("  总收益: %.2f", totalProfit)

	logger.Info("TestBuyHighSellLowStrategyCustomConfig end")
}

func BenchmarkBuyHighSellLowStrategy(b *testing.B) {
	// 加载股票列表和数据
	stockData.LoadPreStockList()
	stockData.LoadDataOneByOne()

	strategy := NewBuyHighSellLowStrategy()
	code := "000001"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		strategy.DealStrategy(code)
	}
}
