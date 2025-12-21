package stockData

import (
	"encoding/csv"
	"log/slog"
	"math/rand/v2"
	"os"
	globalDefine "stock-go/globalDefine"
	"stock-go/logger"
	"strconv"
	"strings"
)

// var path = "../Data/"

// func init() {
// 	sysType := runtime.GOOS

// 	if sysType == "linux" {
// 		path = "../Data/"
// 	}

// 	if sysType == "windows" {
// 		path = "D:\\Data\\"
// 	}
// 	slog.Info("Data Path", "path", path)
// }

// 加载stock列表
func LoadAllStockList() [][]string {
	fileName := globalDefine.DATA_PATH + "stockList.csv"
	fs1, _ := os.Open(fileName)
	r1 := csv.NewReader(fs1)
	content, err := r1.ReadAll()
	if err != nil {
		logger.Error("can not readall", "err", err)
	}

	for _, row := range content {

		row0 := string(row[0])
		row1 := string(row[1])
		row2 := string(row[2])
		var code string
		if strings.Contains(row0, "SZ") {
			code = "sz." + row1
		} else {
			code = "sh." + row1
		}
		StockList[code] = row2

	}

	slog.Info("stock list size", "size", len(content))
	return content
}

// 加载票票列表，随机选择 1/STOCK_DATA_LOAD_PCT 的数据
func LoadPreStockList() map[string]string {
	// 清空现有数据，确保每次都是重新随机选择
	StockList = make(map[string]string)
	StocksRaw = make(map[string]*StockInfo)

	fileName := globalDefine.DATA_PATH + "stockList.csv"
	fs1, _ := os.Open(fileName)
	r1 := csv.NewReader(fs1)
	content, err := r1.ReadAll()
	if err != nil {
		logger.Error("can not readall", "err", err)
	}

	// 使用 math/rand/v2 的全局随机数生成器，自动使用随机种子
	// 生成一个随机标识来验证每次调用确实是新的
	randomMarker := rand.IntN(1000000)

	for _, row := range content {
		// 随机选择 1/STOCK_DATA_LOAD_PCT 的数据
		if rand.IntN(globalDefine.STOCK_DATA_LOAD_PCT) == 0 {
			row0 := string(row[0])
			row1 := string(row[1])
			row2 := string(row[2])
			var code string
			if strings.Contains(row0, "SZ") {
				code = "sz." + row1
			} else {
				code = "sh." + row1
			}
			StockList[code] = row2
		}
	}

	slog.Info("stock list loaded", "size", len(StockList), "random_marker", randomMarker)
	return StockList
}

func LoadFromCsv(code string) (stockData StockData) {
	fileName := globalDefine.DATA_PATH + code + "_ALL.csv"
	fs1, _ := os.Open(fileName)
	r1 := csv.NewReader(fs1)
	content, err := r1.ReadAll()
	if err != nil {
		logger.Error("can not readall", "err", err)
	}

	priceEndY := 0.0
	Interest := 1.0
	i := 0
	for index, row := range content {
		if index != 0 {
			//丢弃停牌数据
			if row[4] == "0" {
				continue
			}
			i++
			stock := new(StockDataDay)
			stock.Index = i
			stock.DataStr = row[0]
			priceBegin, _ := strconv.ParseFloat(row[1], 32)
			priceEnd, _ := strconv.ParseFloat(row[5], 32)
			priceHigh, _ := strconv.ParseFloat(row[6], 32)
			priceLow, _ := strconv.ParseFloat(row[7], 32)

			if priceEndY != 0 && (priceBegin/priceEndY < 0.85) {
				Interest = Interest / priceBegin * priceEndY
			}
			stock.PriceShow = float32(priceBegin+priceEnd) / 2
			stock.PriceBegin = float32(priceBegin * Interest)
			stock.PriceEnd = float32(priceEnd * Interest)
			stock.PriceHigh = float32(priceHigh * Interest)
			stock.PriceLow = float32(priceLow * Interest)
			stock.PriceA = (stock.PriceBegin + stock.PriceEnd) / 2
			stock.PriceBegin = stock.PriceEnd
			stock.PriceA = stock.PriceEnd
			stockData.DayDatas = append(stockData.DayDatas, stock)
			priceEndY = float64(priceEnd)
		}
	}
	//return content
	return
}
