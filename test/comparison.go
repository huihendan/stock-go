package main

import (
	"fmt"
	"stock/stockStrategy"
	"stock/tradeTest"
)

func main() {
	fmt.Println("=== 策略优化前后对比分析 ===")
	
	initialCash := 100000.0
	testCodes := []string{"000001", "000002"}
	
	fmt.Println("\n=== 优化后策略测试 ===")
	
	fmt.Println("策略模式1 (追涨杀跌1 - 优化版):")
	finalWallet1, allRecords1 := tradeTest.BatchTradeTest(testCodes, stockStrategy.Strategy_Mode_1, initialCash)
	stats1 := tradeTest.CalculatePortfolioPerformance(finalWallet1, allRecords1)
	printStats("策略1", stats1)
	
	fmt.Println("\n策略模式2 (追涨杀跌2 - 优化版):")
	finalWallet2, allRecords2 := tradeTest.BatchTradeTest(testCodes, stockStrategy.Strategy_Mode_2, initialCash)
	stats2 := tradeTest.CalculatePortfolioPerformance(finalWallet2, allRecords2)
	printStats("策略2", stats2)
	
	fmt.Println("\n策略模式3 (波动策略 - 优化版):")
	finalWallet3, allRecords3 := tradeTest.BatchTradeTest(testCodes, stockStrategy.Strategy_Mode_3, initialCash)
	stats3 := tradeTest.CalculatePortfolioPerformance(finalWallet3, allRecords3)
	printStats("策略3", stats3)
	
	fmt.Println("\n=== 优化效果总结 ===")
	fmt.Printf("原始策略: 胜率 15.91%%, 总交易 220笔, 总亏损 -2914.60\n")
	fmt.Printf("优化策略1: 胜率 %.2f%%, 总交易 %d笔, 总盈亏 %.2f\n", 
		stats1.WinRate*100, stats1.TotalTrades, stats1.TotalProfit)
	fmt.Printf("优化策略2: 胜率 %.2f%%, 总交易 %d笔, 总盈亏 %.2f\n", 
		stats2.WinRate*100, stats2.TotalTrades, stats2.TotalProfit)
	fmt.Printf("优化策略3: 胜率 %.2f%%, 总交易 %d笔, 总盈亏 %.2f\n", 
		stats3.WinRate*100, stats3.TotalTrades, stats3.TotalProfit)
}

func printStats(strategyName string, stats tradeTest.PortfolioStats) {
	fmt.Printf("  总交易数: %d\n", stats.TotalTrades)
	fmt.Printf("  盈利交易: %d\n", stats.WinningTrades)
	fmt.Printf("  胜率: %.2f%%\n", stats.WinRate*100)
	fmt.Printf("  总盈亏: %.2f\n", stats.TotalProfit)
	fmt.Printf("  投资组合价值: %.2f\n", stats.CurrentValue)
	fmt.Printf("  剩余现金: %.2f\n", stats.CashRemaining)
}