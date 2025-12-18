package tradeTest

import (
	"fmt"
	"sort"
	globaldefine "stock-go/globalDefine"
	"stock-go/stockStrategy"
)

const (
	OperateType_Buy  = 1
	OperateType_Sell = 2

	Status_Holding = 1
	Status_Sold    = 2
)

func TradeTest(code string, strategyMode int, wallet globaldefine.Wallet) (globaldefine.Wallet, []globaldefine.OperateRecord) {
	var operateRecords []globaldefine.OperateRecord

	strategies := stockStrategy.DealStrategys(code, strategyMode)

	var sortedKeys []string
	for key := range strategies {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Strings(sortedKeys)

	for _, key := range sortedKeys {
		strategy := strategies[key]

		requiredCash := float32(strategy.BuyOperate.BuyPrice) * float32(strategy.BuyOperate.StockNum)

		if wallet.Cash >= requiredCash {
			wallet.Cash -= requiredCash

			position := globaldefine.Position{
				StockCode: strategy.StockCode,
				StockName: strategy.StockName,
				StockNum:  int(strategy.BuyOperate.StockNum),
				BuyPrice:  float32(strategy.BuyOperate.BuyPrice),
				SellPrice: float32(strategy.SellOperate.SellPrice),
				BuyDate:   strategy.BuyOperate.OperateDate,
				Profit:    float32(strategy.Profit),
			}

			wallet.Positions = append(wallet.Positions, position)

			if strategy.Status == Status_Sold {
				sellAmount := float32(strategy.SellOperate.SellPrice) * float32(strategy.SellOperate.StockNum)
				wallet.Cash += sellAmount
			}

			operateRecord := globaldefine.OperateRecord{
				StockCode:   strategy.StockCode,
				StockName:   strategy.StockName,
				StockNum:    int(strategy.StockNum),
				BuyOperate:  convertToTradeTestOperate(strategy.BuyOperate),
				SellOperate: convertToTradeTestOperate(strategy.SellOperate),
				Status:      strategy.Status,
				Profit:      float32(strategy.Profit),
				OperateDate: strategy.OperateDate,
				OperateTime: strategy.OperateTime,
			}

			operateRecords = append(operateRecords, operateRecord)
		}
	}

	return wallet, operateRecords
}

func BatchTradeTest(codes []string, strategyMode int, initialCash float64) (globaldefine.Wallet, map[string][]globaldefine.OperateRecord) {
	wallet := globaldefine.Wallet{
		Cash:      float32(initialCash),
		Positions: []globaldefine.Position{},
	}

	allOperateRecords := make(map[string][]globaldefine.OperateRecord)

	for _, code := range codes {
		fmt.Printf("Testing strategy %d for stock %s\n", strategyMode, code)

		updatedWallet, records := TradeTest(code, strategyMode, wallet)
		wallet = updatedWallet

		if len(records) > 0 {
			allOperateRecords[code] = records
		}
	}

	return wallet, allOperateRecords
}

func CalculatePortfolioPerformance(wallet globaldefine.Wallet, operateRecords map[string][]globaldefine.OperateRecord) PortfolioStats {
	totalTrades := 0
	totalProfit := 0.0
	winningTrades := 0

	for _, records := range operateRecords {
		for _, record := range records {
			totalTrades++
			totalProfit += float64(record.Profit)

			if record.Profit > 0 {
				winningTrades++
			}
		}
	}

	currentValue := float64(wallet.Cash)
	for _, position := range wallet.Positions {
		currentValue += float64(position.SellPrice) * float64(position.StockNum)
	}

	winRate := 0.0
	if totalTrades > 0 {
		winRate = float64(winningTrades) / float64(totalTrades)
	}

	return PortfolioStats{
		TotalTrades:    totalTrades,
		WinningTrades:  winningTrades,
		TotalProfit:    totalProfit,
		WinRate:        winRate,
		CurrentValue:   currentValue,
		CashRemaining:  float64(wallet.Cash),
		PositionsCount: len(wallet.Positions),
	}
}

type PortfolioStats struct {
	TotalTrades    int
	WinningTrades  int
	TotalProfit    float64
	WinRate        float64
	CurrentValue   float64
	CashRemaining  float64
	PositionsCount int
}

func convertToTradeTestOperate(strategyOperate stockStrategy.Operate) globaldefine.Operate {
	return globaldefine.Operate{
		OperateType: strategyOperate.OperateType,
		BuyPrice:    float32(strategyOperate.BuyPrice),
		SellPrice:   float32(strategyOperate.SellPrice),
		OperateDate: strategyOperate.OperateDate,
		OperateTime: strategyOperate.OperateTime,
		StockCode:   strategyOperate.StockCode,
		StockName:   strategyOperate.StockName,
		StockNum:    strategyOperate.StockNum,
	}
}

func PrintTradeResults(wallet globaldefine.Wallet, operateRecords map[string][]globaldefine.OperateRecord) {
	fmt.Println("=== Trade Test Results ===")

	stats := CalculatePortfolioPerformance(wallet, operateRecords)

	fmt.Printf("Total Trades: %d\n", stats.TotalTrades)
	fmt.Printf("Winning Trades: %d\n", stats.WinningTrades)
	fmt.Printf("Win Rate: %.2f%%\n", stats.WinRate*100)
	fmt.Printf("Total Profit: %.2f\n", stats.TotalProfit)
	fmt.Printf("Current Portfolio Value: %.2f\n", stats.CurrentValue)
	fmt.Printf("Cash Remaining: %.2f\n", stats.CashRemaining)
	fmt.Printf("Active Positions: %d\n", stats.PositionsCount)

	fmt.Println("\n=== Trade Details ===")
	for stockCode, records := range operateRecords {
		fmt.Printf("Stock: %s\n", stockCode)
		for i, record := range records {
			fmt.Printf("  Trade %d: Buy %.2f, Sell %.2f, Profit %.2f, Date %s\n",
				i+1, record.BuyOperate.BuyPrice, record.SellOperate.SellPrice,
				record.Profit, record.OperateDate)
		}
	}
}
