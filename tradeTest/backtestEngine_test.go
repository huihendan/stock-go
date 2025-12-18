package tradeTest

import (
	"fmt"
	"stock-go/stockData"
	"stock-go/stockStrategy/strategies"
	"testing"
)

// TestBacktestEngineWithStrategy1 测试新回测引擎和策略1
func TestBacktestEngineWithStrategy1(t *testing.T) {
	// 1. 加载票票列表
	stockData.LoadPreStockList()
	fmt.Printf("票票列表加载完成，共 %d 只票票\n", len(stockData.StockList))

	// 2. 创建策略1
	strategy := strategies.NewBuyHighSellLowStrategy()
	fmt.Printf("策略名称: %s\n", strategy.GetName())

	// 3. 创建回测引擎
	initialCash := 1000000.0 // 100万初始资金
	engine := NewBacktestEngine(initialCash, strategy)

	// 4. 执行回测
	result := engine.Run()

	// 5. 输出结果
	fmt.Printf("\n========== 回测结果 ==========\n")
	fmt.Printf("最终现金: %.2f\n", result.Wallet.Cash)
	fmt.Printf("持仓数量: %d\n", len(result.Wallet.Positions))
	fmt.Printf("交易票票数量: %d\n", len(result.OperateRecords))

	// 统计总交易次数
	totalTrades := 0
	totalProfit := float32(0)
	winningTrades := 0

	for code, records := range result.OperateRecords {
		fmt.Printf("\n票票 %s:\n", code)
		for i, record := range records {
			totalTrades++
			totalProfit += record.Profit
			if record.Profit > 0 {
				winningTrades++
			}

			fmt.Printf("  [%d] 买入: %s %.2f 卖出: %s %.2f 收益: %.2f\n",
				i+1,
				record.BuyOperate.OperateDate,
				record.BuyOperate.BuyPrice,
				record.SellOperate.OperateDate,
				record.SellOperate.SellPrice,
				record.Profit,
			)
		}
	}

	fmt.Printf("\n========== 统计指标 ==========\n")
	fmt.Printf("总交易次数: %d\n", totalTrades)
	fmt.Printf("盈利交易: %d\n", winningTrades)
	fmt.Printf("总收益: %.2f\n", totalProfit)
	if totalTrades > 0 {
		fmt.Printf("胜率: %.2f%%\n", float64(winningTrades)/float64(totalTrades)*100)
		fmt.Printf("平均收益: %.2f\n", totalProfit/float32(totalTrades))
	}
	fmt.Printf("收益率: %.2f%%\n", (totalProfit/float32(initialCash))*100)
	fmt.Printf("========================================\n")

	// 基本断言
	if totalTrades == 0 {
		t.Error("没有产生任何交易")
	}
}

// TestBacktestEngineWithCustomParams 测试自定义参数
func TestBacktestEngineWithCustomParams(t *testing.T) {
	// 加载票票列表
	stockData.LoadPreStockList()

	// 创建自定义参数的策略
	strategy := strategies.NewBuyHighSellLowStrategyWithParams(
		300, // 选股回看300天
		10,  // 最近10天出现高点
		200, // 信号回看200天
		0.08, // 止损8%
		20,  // 最多持有20天
	)

	fmt.Printf("\n测试自定义参数策略\n")
	fmt.Printf("策略名称: %s\n", strategy.GetName())

	// 创建回测引擎
	engine := NewBacktestEngine(500000.0, strategy) // 50万初始资金

	// 执行回测
	result := engine.Run()

	// 简单输出
	fmt.Printf("\n最终现金: %.2f\n", result.Wallet.Cash)
	fmt.Printf("交易票票数量: %d\n", len(result.OperateRecords))
}

// TestBacktestSingleStock 测试单只票票回测
func TestBacktestSingleStock(t *testing.T) {
	// 加载票票列表
	stockData.LoadPreStockList()

	// 选择一只票票进行测试
	testCode := ""
	for code := range stockData.StockList {
		testCode = code
		break // 取第一只
	}

	if testCode == "" {
		t.Skip("没有可用的票票")
	}

	fmt.Printf("\n测试单只票票: %s\n", testCode)

	// 加载票票数据
	stockInfo := stockData.GetStockRawBycode(testCode)
	if stockInfo == nil {
		t.Fatalf("无法加载票票 %s", testCode)
	}

	fmt.Printf("票票名称: %s\n", stockInfo.Name)
	fmt.Printf("数据天数: %d\n", len(stockInfo.Datas.DayDatas))

	// 创建策略并回测
	strategy := strategies.NewBuyHighSellLowStrategy()
	engine := NewBacktestEngine(1000000.0, strategy)
	result := engine.Run()

	fmt.Printf("回测完成，交易记录数: %d\n", len(result.OperateRecords))
}
