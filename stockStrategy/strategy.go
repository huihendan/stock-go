package stockStrategy

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"time"
)

// 策略模式常量
const (
	Strategy_Mode_1 = 1 // 追涨杀跌1
	Strategy_Mode_2 = 2 // 突破策略
	Strategy_Mode_3 = 3 // 波段策略
	Strategy_Mode_4 = 4 // 大盘策略
	Strategy_Mode_5 = 5 // 板块突破
	Strategy_Mode_6 = 6 // 板块轮动
)

// StockData 旧版本的票票数据结构（用于CSV加载）
type StockData struct {
	Date        time.Time
	Open        float64
	Close       float64
	High        float64
	Low         float64
	PeTTM       float64
	PbMRQ       float64
	TradeStatus int
}

// OperateRecord 操作记录
type OperateRecord struct {
	StockCode   string
	StockName   string
	StockNum    float64
	BuyOperate  Operate
	SellOperate Operate
	Status      int     // 1-已买入 2-已卖出
	Profit      float64 // 收益
	OperateDate string
	OperateTime string
}

// Operate 单次操作
type Operate struct {
	OperateType int // 1-买入 2-卖出
	BuyPrice    float64
	SellPrice   float64
	OperateDate string
	OperateTime string
	StockCode   string
	StockName   string
	StockNum    int
}

func NewStrategy(strategyMode int) StockStrategy {
	switch strategyMode {
	case Strategy_Mode_1:
		return NewBuyHighSellLowStrategy()
	case Strategy_Mode_2:
		//return NewBreakoutStrategy()
	case Strategy_Mode_3:
		//return NewSwingStrategy()
	case Strategy_Mode_4:
		//return New大盘Strategy()
	case Strategy_Mode_5:
		//return New板块突破Strategy()
	case Strategy_Mode_6:
		//return New板块轮动Strategy()
	}

	return nil
}

// DealStrategys 策略分发函数 - 根据策略模式调用相应策略
func DealStrategys(code string, strategyMode int) map[string]OperateRecord {
	strategy := NewStrategy(strategyMode)
	return strategy.DealStrategy(code)
}

// dealStrategysMode1 已废弃 - 请使用 BuyHighSellLowStrategy
// 保留此函数用于向后兼容，但实际调用新的策略实现
// 注意：此旧实现存在"未来函数"问题（findSellPriceOptimized会查看未来数据）
func dealStrategysMode1(code string, strategyMode int) map[string]OperateRecord {
	// 直接调用新的策略实现
	strategy := NewBuyHighSellLowStrategy()
	return strategy.DealStrategy(code)
}

// dealStrategysMode1Old 旧的实现保留作为参考
// 警告：此实现存在未来函数问题，不建议使用
func dealStrategysMode1Old(code string) map[string]OperateRecord {
	operates := make(map[string]OperateRecord)

	stockData, err := loadStockData(code)
	if err != nil {
		fmt.Printf("Error loading stock data for %s: %v\n", code, err)
		return operates
	}

	if len(stockData) < 300 {
		return operates
	}

	sort.Slice(stockData, func(i, j int) bool {
		return stockData[i].Date.Before(stockData[j].Date)
	})

	windowSize := 300
	var lastBuyIndex int = -50

	for i := windowSize; i < len(stockData); i++ {
		if i-lastBuyIndex < 10 {
			continue
		}

		window := stockData[i-windowSize : i]

		highestPrice := 0.0
		for _, data := range window {
			if data.High > highestPrice {
				highestPrice = data.High
			}
		}

		currentPrice := stockData[i].Close
		ma20 := calculateMA(stockData, i, 20)
		ma60 := calculateMA(stockData, i, 60)

		if currentPrice >= highestPrice*0.995 &&
			ma20 > ma60 &&
			currentPrice > ma20 &&
			isUptrend(stockData, i, 10) &&
			stockData[i].TradeStatus == 1 {

			buyOperate := Operate{
				OperateType: 1,
				BuyPrice:    currentPrice,
				OperateDate: stockData[i].Date.Format("2006-01-02"),
				OperateTime: stockData[i].Date.Format("15:04:05"),
				StockCode:   code,
				StockNum:    100,
			}

			// 警告：findSellPriceOptimized 会查看未来数据
			sellPrice := findSellPriceOptimized(stockData, i+1, currentPrice)
			sellOperate := Operate{
				OperateType: 2,
				SellPrice:   sellPrice,
				OperateDate: stockData[i].Date.Format("2006-01-02"),
				OperateTime: stockData[i].Date.Format("15:04:05"),
				StockCode:   code,
				StockNum:    100,
			}

			profit := (sellPrice - currentPrice) * 100

			recordKey := fmt.Sprintf("%s_%d", code, i)
			operates[recordKey] = OperateRecord{
				StockCode:   code,
				StockNum:    100,
				BuyOperate:  buyOperate,
				SellOperate: sellOperate,
				Status:      2,
				Profit:      profit,
				OperateDate: stockData[i].Date.Format("2006-01-02"),
				OperateTime: stockData[i].Date.Format("15:04:05"),
			}

			lastBuyIndex = i
		}
	}

	return operates
}

// dealStrategysMode2 突破策略（待重构为新策略模式）
// TODO: 重构为独立的策略结构体，避免未来函数问题
func dealStrategysMode2(code string, strategyMode int) map[string]OperateRecord {
	operates := make(map[string]OperateRecord)

	stockData, err := loadStockData(code)
	if err != nil {
		return operates
	}

	sort.Slice(stockData, func(i, j int) bool {
		return stockData[i].Date.Before(stockData[j].Date)
	})

	lookbackPeriod := 20
	var lastBuyIndex int = -15

	for i := lookbackPeriod + 60; i < len(stockData); i++ {
		if i-lastBuyIndex < 5 {
			continue
		}

		recentWindow := stockData[i-lookbackPeriod : i]

		recentHigh := 0.0
		for _, data := range recentWindow {
			if data.High > recentHigh {
				recentHigh = data.High
			}
		}

		currentPrice := stockData[i].Close
		ma5 := calculateMA(stockData, i, 5)
		ma20 := calculateMA(stockData, i, 20)
		ma60 := calculateMA(stockData, i, 60)
		rsi := calculateRSI(stockData, i, 14)

		if currentPrice > recentHigh &&
			currentPrice > ma5 && ma5 > ma20 && ma20 > ma60 &&
			rsi < 70 && rsi > 30 &&
			isBreakoutValid(stockData, i, recentHigh) &&
			stockData[i].TradeStatus == 1 {

			buyOperate := Operate{
				OperateType: 1,
				BuyPrice:    currentPrice,
				OperateDate: stockData[i].Date.Format("2006-01-02"),
				OperateTime: stockData[i].Date.Format("15:04:05"),
				StockCode:   code,
				StockNum:    100,
			}

			sellPrice := findSellPriceOptimized(stockData, i+1, currentPrice)
			sellOperate := Operate{
				OperateType: 2,
				SellPrice:   sellPrice,
				OperateDate: stockData[i].Date.Format("2006-01-02"),
				OperateTime: stockData[i].Date.Format("15:04:05"),
				StockCode:   code,
				StockNum:    100,
			}

			profit := (sellPrice - currentPrice) * 100

			recordKey := fmt.Sprintf("%s_%d", code, i)
			operates[recordKey] = OperateRecord{
				StockCode:   code,
				StockNum:    100,
				BuyOperate:  buyOperate,
				SellOperate: sellOperate,
				Status:      2,
				Profit:      profit,
				OperateDate: stockData[i].Date.Format("2006-01-02"),
				OperateTime: stockData[i].Date.Format("15:04:05"),
			}

			lastBuyIndex = i
		}
	}

	return operates
}

// dealStrategysMode3 波段策略（待重构为新策略模式）
// TODO: 重构为独立的策略结构体，避免未来函数问题
func dealStrategysMode3(code string, strategyMode int) map[string]OperateRecord {
	operates := make(map[string]OperateRecord)

	stockData, err := loadStockData(code)
	if err != nil {
		return operates
	}

	sort.Slice(stockData, func(i, j int) bool {
		return stockData[i].Date.Before(stockData[j].Date)
	})

	swingPeriod := 60
	var lastBuyIndex int = -20

	for i := swingPeriod + 60; i < len(stockData); i++ {
		if i-lastBuyIndex < 15 {
			continue
		}

		window := stockData[i-swingPeriod : i]

		var prices []float64
		for _, data := range window {
			prices = append(prices, data.Close)
		}

		mean := calculateMean(prices)
		stdDev := calculateStdDev(prices, mean)

		currentPrice := stockData[i].Close
		ma20 := calculateMA(stockData, i, 20)
		ma60 := calculateMA(stockData, i, 60)
		rsi := calculateRSI(stockData, i, 14)

		lowerBollinger := ma20 - 2*stdDev

		if currentPrice <= lowerBollinger &&
			currentPrice < mean*0.92 &&
			rsi < 35 &&
			ma20 < ma60*1.05 &&
			isOversold(stockData, i, 10) &&
			stockData[i].TradeStatus == 1 {

			buyOperate := Operate{
				OperateType: 1,
				BuyPrice:    currentPrice,
				OperateDate: stockData[i].Date.Format("2006-01-02"),
				OperateTime: stockData[i].Date.Format("15:04:05"),
				StockCode:   code,
				StockNum:    100,
			}

			sellPrice := findSellPriceOptimized(stockData, i+1, currentPrice)
			sellOperate := Operate{
				OperateType: 2,
				SellPrice:   sellPrice,
				OperateDate: stockData[i].Date.Format("2006-01-02"),
				OperateTime: stockData[i].Date.Format("15:04:05"),
				StockCode:   code,
				StockNum:    100,
			}

			profit := (sellPrice - currentPrice) * 100

			recordKey := fmt.Sprintf("%s_%d", code, i)
			operates[recordKey] = OperateRecord{
				StockCode:   code,
				StockNum:    100,
				BuyOperate:  buyOperate,
				SellOperate: sellOperate,
				Status:      2,
				Profit:      profit,
				OperateDate: stockData[i].Date.Format("2006-01-02"),
				OperateTime: stockData[i].Date.Format("15:04:05"),
			}

			lastBuyIndex = i
		}
	}

	return operates
}

// dealStrategysMode4 大盘策略（待实现）
// TODO: 实现为独立的策略结构体
func dealStrategysMode4(code string, strategyMode int) map[string]OperateRecord {
	operates := make(map[string]OperateRecord)
	return operates
}

// dealStrategysMode5 板块突破策略（待实现）
// TODO: 实现为独立的策略结构体
func dealStrategysMode5(code string, strategyMode int) map[string]OperateRecord {
	operates := make(map[string]OperateRecord)
	return operates
}

// dealStrategysMode6 板块轮动策略（待实现）
// TODO: 实现为独立的策略结构体
func dealStrategysMode6(code string, strategyMode int) map[string]OperateRecord {
	operates := make(map[string]OperateRecord)
	return operates
}

// ====== 辅助函数 ======
// 以下是策略计算的辅助函数，可被各个策略共用

// loadStockData 从CSV文件加载票票数据（旧方法）
// 注意：新的策略实现应该使用 stockData.GetstockBycode
func loadStockData(code string) ([]StockData, error) {
	filename := fmt.Sprintf("Data/%s_ALL.csv", code)
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var stockData []StockData
	for i, record := range records {
		if i == 0 {
			continue
		}

		date, _ := time.Parse("2006-01-02", record[0])
		open, _ := strconv.ParseFloat(record[1], 64)
		peTTM, _ := strconv.ParseFloat(record[2], 64)
		pbMRQ, _ := strconv.ParseFloat(record[3], 64)
		tradeStatus, _ := strconv.Atoi(record[4])
		close, _ := strconv.ParseFloat(record[5], 64)
		high, _ := strconv.ParseFloat(record[6], 64)
		low, _ := strconv.ParseFloat(record[7], 64)

		stockData = append(stockData, StockData{
			Date:        date,
			Open:        open,
			Close:       close,
			High:        high,
			Low:         low,
			PeTTM:       peTTM,
			PbMRQ:       pbMRQ,
			TradeStatus: tradeStatus,
		})
	}

	return stockData, nil
}

// findSellPriceOptimized 寻找卖出价格（存在未来函数问题）
// 警告：此函数会查看未来数据，仅用于旧策略，不建议在新策略中使用
// 新策略应该在遍历时实时判断卖出条件
func findSellPriceOptimized(stockData []StockData, startIndex int, buyPrice float64) float64 {
	targetProfit := buyPrice * 1.15
	stopLoss := buyPrice * 0.96
	trailingStop := buyPrice * 0.98
	maxPrice := buyPrice

	for i := startIndex; i < len(stockData) && i < startIndex+40; i++ {
		if stockData[i].High > maxPrice {
			maxPrice = stockData[i].High
			trailingStop = maxPrice * 0.95
		}

		if stockData[i].High >= targetProfit {
			return targetProfit
		}

		if stockData[i].Low <= stopLoss {
			return stopLoss
		}

		if maxPrice > buyPrice*1.05 && stockData[i].Low <= trailingStop {
			return trailingStop
		}
	}

	if startIndex < len(stockData) {
		return stockData[startIndex].Close
	}
	return buyPrice
}

// calculateMA 计算移动平均线（MA）
func calculateMA(stockData []StockData, index int, period int) float64 {
	if index < period-1 {
		return 0
	}

	sum := 0.0
	for i := index - period + 1; i <= index; i++ {
		sum += stockData[i].Close
	}
	return sum / float64(period)
}

// calculateRSI 计算相对强弱指标（RSI）
func calculateRSI(stockData []StockData, index int, period int) float64 {
	if index < period {
		return 50
	}

	gains := 0.0
	losses := 0.0

	for i := index - period + 1; i <= index; i++ {
		if i == 0 {
			continue
		}
		change := stockData[i].Close - stockData[i-1].Close
		if change > 0 {
			gains += change
		} else {
			losses += -change
		}
	}

	if losses == 0 {
		return 100
	}

	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)
	rs := avgGain / avgLoss
	rsi := 100 - (100 / (1 + rs))

	return rsi
}

// isUptrend 判断是否处于上升趋势
func isUptrend(stockData []StockData, index int, period int) bool {
	if index < period {
		return false
	}

	upDays := 0
	for i := index - period + 1; i <= index; i++ {
		if i > 0 && stockData[i].Close > stockData[i-1].Close {
			upDays++
		}
	}

	return float64(upDays)/float64(period) > 0.6
}

// isBreakoutValid 判断突破是否有效
func isBreakoutValid(stockData []StockData, index int, recentHigh float64) bool {
	if index < 3 {
		return false
	}

	currentPrice := stockData[index].Close
	prevPrice := stockData[index-1].Close

	breakoutStrength := (currentPrice - recentHigh) / recentHigh
	return breakoutStrength > 0.02 && currentPrice > prevPrice
}

// isOversold 判断是否超卖
func isOversold(stockData []StockData, index int, period int) bool {
	if index < period {
		return false
	}

	downDays := 0
	for i := index - period + 1; i <= index; i++ {
		if i > 0 && stockData[i].Close < stockData[i-1].Close {
			downDays++
		}
	}

	return float64(downDays)/float64(period) > 0.7
}

// calculateMean 计算平均值
func calculateMean(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// calculateStdDev 计算标准差
func calculateStdDev(values []float64, mean float64) float64 {
	sumSquares := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquares += diff * diff
	}
	variance := sumSquares / float64(len(values))
	return math.Sqrt(variance)
}

// ====== 文件说明 ======
//
// 本文件保留了旧的策略实现，主要用于向后兼容。
//
// 新的策略实现请参考：
// - buyHighSellLowStrategy.go (策略1的新实现)
// - types.go (策略接口和类型定义)
//
// 主要变化：
// 1. 策略1已重构，使用新的BuyHighSellLowStrategy，避免了未来函数问题
// 2. 策略2-3保留旧实现，但标记为待重构（存在未来函数问题）
// 3. 策略4-6尚未实现
// 4. 辅助函数保留供各策略共用
//
// 未来规划：
// - 将所有策略重构为实现StockStrategy接口的独立结构体
// - 消除所有未来函数问题
// - 提高代码可测试性和可维护性
