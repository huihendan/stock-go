package tradeTest

import (
	"fmt"
	"stock-go/stockData"
	"stock-go/stockStrategy/strategies"
	"testing"
	"time"
)

// TestTimeBasedBacktestEngine 测试基于时间流逝的回测引擎
func TestTimeBasedBacktestEngine(t *testing.T) {
	// 输出时间戳
	fmt.Printf("\n[测试开始时间: %s]\n", time.Now().Format("15:04:05.000"))

	// 1. 加载票票列表
	stockData.LoadPreStockList()
	fmt.Printf("票票列表加载完成，共 %d 只票票\n", len(stockData.StockList))

	t.Logf("注意：本测试使用随机选择的票票数据，每次运行结果会不同")

	// 2. 创建策略
	strategy := strategies.NewBuyHighSellLowStrategy()
	fmt.Printf("策略名称: %s\n", strategy.GetName())

	// 3. 创建基于时间流逝的回测引擎
	initialCash := 1000000.0 // 100万初始资金
	maxPositions := 2        // 最多同时持有2只票票
	cashPerPosition := 0.5   // 每个持仓使用50%的现金

	engine := NewTimeBasedBacktestEngine(
		initialCash,
		strategy,
		maxPositions,
		cashPerPosition,
	)

	// 4. 执行回测
	result := engine.Run()

	// 5. 验证结果
	if result == nil {
		t.Fatal("回测结果为空")
	}

	// 6. 输出详细结果
	fmt.Printf("\n========== 详细交易记录 ==========\n")
	buyCount := 0
	sellCount := 0
	for _, record := range result.TradeRecords {
		if record.Action == "buy" {
			buyCount++
			if buyCount <= 5 { // 只显示前5笔买入
				fmt.Printf("[买入] %s(%s) 日期:%s 价格:%.2f 数量:%d 金额:%.2f 剩余现金:%.2f\n",
					record.Name, record.Code, record.Date, record.Price,
					record.StockNum, record.Amount, record.Cash)
			}
		} else {
			sellCount++
			if sellCount <= 5 { // 只显示前5笔卖出
				fmt.Printf("[卖出] %s(%s) 日期:%s 价格:%.2f 数量:%d 金额:%.2f 剩余现金:%.2f 原因:%s\n",
					record.Name, record.Code, record.Date, record.Price,
					record.StockNum, record.Amount, record.Cash, record.Reason)
			}
		}
	}

	if buyCount > 5 {
		fmt.Printf("... (还有 %d 笔买入)\n", buyCount-5)
	}
	if sellCount > 5 {
		fmt.Printf("... (还有 %d 笔卖出)\n", sellCount-5)
	}

	fmt.Printf("\n========== 资金曲线（采样） ==========\n")
	sampleInterval := len(result.DailyEquity) / 10 // 采样10个点
	if sampleInterval < 1 {
		sampleInterval = 1
	}
	for i := 0; i < len(result.DailyEquity); i += sampleInterval {
		equity := result.DailyEquity[i]
		fmt.Printf("日期:%s 现金:%.2f 持仓市值:%.2f 总资产:%.2f 持仓数:%d\n",
			equity.Date, equity.Cash, equity.PositionValue,
			equity.TotalAssets, equity.PositionCount)
	}

	// 基本断言
	if result.TotalTrades == 0 {
		t.Error("没有产生任何交易")
	}

	if result.FinalAssets < 0 {
		t.Error("最终资产为负")
	}

	// 验证资金管理
	// 理论上最多使用的资金不应超过初始资金
	// （因为我们严格管理持仓数量和每仓位资金）
	fmt.Printf("\n========== 资金管理验证 ==========\n")
	fmt.Printf("最大持仓数限制: %d\n", maxPositions)
	fmt.Printf("每仓位资金比例: %.1f%%\n", cashPerPosition*100)
	fmt.Printf("理论最大资金使用: %.2f (%.1f%%)\n",
		initialCash*cashPerPosition*float64(maxPositions),
		cashPerPosition*float64(maxPositions)*100)
}

// TestTimeBasedBacktestEngineWithDifferentParams 测试不同参数配置
func TestTimeBasedBacktestEngineWithDifferentParams(t *testing.T) {
	// 输出时间戳
	fmt.Printf("\n[测试开始时间: %s]\n", time.Now().Format("15:04:05.000"))

	// 加载票票列表
	stockData.LoadPreStockList()

	// 测试配置1：高仓位策略（50%%单仓位）
	fmt.Printf("\n========================================\n")
	fmt.Printf("配置1：高仓位策略（50%%单仓位）\n")
	fmt.Printf("========================================\n")

	strategy1 := strategies.NewBuyHighSellLowStrategy()
	engine1 := NewTimeBasedBacktestEngine(
		1000000.0, // 100万
		strategy1,
		2,   // 最多2个持仓
		0.5, // 每仓50%
	)
	result1 := engine1.Run()

	if result1 != nil {
		fmt.Printf("高仓位策略收益率: %.2f%%\n", result1.TotalReturnPct)
	}

	// 测试配置2：分散型策略（多持仓小仓位）
	fmt.Printf("\n========================================\n")
	fmt.Printf("配置2：分散型策略（多持仓小仓位）\n")
	fmt.Printf("========================================\n")

	strategy2 := strategies.NewBuyHighSellLowStrategy()
	engine2 := NewTimeBasedBacktestEngine(
		1000000.0, // 100万
		strategy2,
		10,  // 最多10个持仓
		0.1, // 每仓10%
	)
	result2 := engine2.Run()

	if result2 != nil {
		fmt.Printf("分散型策略收益率: %.2f%%\n", result2.TotalReturnPct)
	}

	// 对比结果
	if result1 != nil && result2 != nil {
		fmt.Printf("\n========== 策略对比 ==========\n")
		fmt.Printf("高仓位策略: 收益 %.2f%% | 最大回撤 %.2f%% | 胜率 %.2f%%\n",
			result1.TotalReturnPct, result1.MaxDrawdown, result1.WinRate)
		fmt.Printf("分散型策略: 收益 %.2f%% | 最大回撤 %.2f%% | 胜率 %.2f%%\n",
			result2.TotalReturnPct, result2.MaxDrawdown, result2.WinRate)
	}
}
