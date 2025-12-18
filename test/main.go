package main

import (
	"fmt"
	"stock-go/stockStrategy"
	"stock-go/tradeTest"
)

func main() {
	fmt.Println("=== Stock Strategy & Backtesting System Test ===")

	testStrategyModule()
	testBacktestingModule()
}

func testStrategyModule() {
	fmt.Println("\n--- Testing Strategy Module ---")

	stockCode := "000001"

	fmt.Printf("Testing Strategy Mode 1 (追涨杀跌1) for stock %s\n", stockCode)
	operates1 := stockStrategy.DealStrategys(stockCode, stockStrategy.Strategy_Mode_1)
	fmt.Printf("Generated %d trade operations\n", len(operates1))

	if len(operates1) > 0 {
		fmt.Println("\nSample operation from Strategy Mode 1:")
		for key, operate := range operates1 {
			fmt.Printf("Key: %s\n", key)
			fmt.Printf("  Stock: %s, Buy Price: %.2f, Sell Price: %.2f, Profit: %.2f\n",
				operate.StockCode, operate.BuyOperate.BuyPrice,
				operate.SellOperate.SellPrice, operate.Profit)
			break
		}
	}
}

func testBacktestingModule() {
	fmt.Println("\n--- Testing Backtesting Module ---")

	initialCash := 100000.0
	testCodes := []string{"000001", "000002"}

	fmt.Printf("Starting backtesting with %.2f initial cash\n", initialCash)
	fmt.Printf("Testing stocks: %v\n", testCodes)

	finalWallet, allRecords := tradeTest.BatchTradeTest(testCodes, stockStrategy.Strategy_Mode_1, initialCash)

	fmt.Printf("\nBacktesting completed!\n")
	fmt.Printf("Final cash: %.2f\n", finalWallet.Cash)
	fmt.Printf("Number of positions: %d\n", len(finalWallet.Positions))

	totalTrades := 0
	for _, records := range allRecords {
		totalTrades += len(records)
	}
	fmt.Printf("Total trades executed: %d\n", totalTrades)

	tradeTest.PrintTradeResults(finalWallet, allRecords)
}
