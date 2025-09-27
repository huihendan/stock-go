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

const (
	Strategy_Mode_1 = 1
	Strategy_Mode_2 = 2
	Strategy_Mode_3 = 3
	Strategy_Mode_4 = 4
	Strategy_Mode_5 = 5
	Strategy_Mode_6 = 6
)

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

type OperateRecord struct {
	StockCode    string
	StockName    string
	StockNum     float64
	BuyOperate   Operate
	SellOperate  Operate
	Status       int
	Profit       float64
	OperateDate  string
	OperateTime  string
}

type Operate struct {
	OperateType int
	BuyPrice    float64
	SellPrice   float64
	OperateDate string
	OperateTime string
	StockCode   string
	StockName   string
	StockNum    int
}

func DealStrategys(code string, strategyMode int) map[string]OperateRecord {
	switch strategyMode {
	case Strategy_Mode_1:
		return dealStrategysMode1(code, strategyMode)
	case Strategy_Mode_2:
		return dealStrategysMode2(code, strategyMode)
	case Strategy_Mode_3:
		return dealStrategysMode3(code, strategyMode)
	case Strategy_Mode_4:
		return dealStrategysMode4(code, strategyMode)
	case Strategy_Mode_5:
		return dealStrategysMode5(code, strategyMode)
	case Strategy_Mode_6:
		return dealStrategysMode6(code, strategyMode)
	default:
		return make(map[string]OperateRecord)
	}
}

func dealStrategysMode1(code string, strategyMode int) map[string]OperateRecord {
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

func dealStrategysMode4(code string, strategyMode int) map[string]OperateRecord {
	operates := make(map[string]OperateRecord)
	return operates
}

func dealStrategysMode5(code string, strategyMode int) map[string]OperateRecord {
	operates := make(map[string]OperateRecord)
	return operates
}

func dealStrategysMode6(code string, strategyMode int) map[string]OperateRecord {
	operates := make(map[string]OperateRecord)
	return operates
}

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

func isBreakoutValid(stockData []StockData, index int, recentHigh float64) bool {
	if index < 3 {
		return false
	}
	
	currentPrice := stockData[index].Close
	prevPrice := stockData[index-1].Close
	
	breakoutStrength := (currentPrice - recentHigh) / recentHigh
	return breakoutStrength > 0.02 && currentPrice > prevPrice
}

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

func calculateMean(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func calculateStdDev(values []float64, mean float64) float64 {
	sumSquares := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquares += diff * diff
	}
	variance := sumSquares / float64(len(values))
	return math.Sqrt(variance)
}