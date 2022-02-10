package stockData

import (
	"encoding/csv"
	"github.com/apache/dubbo-go/common/logger"
	"os"
	"strconv"
	"strings"
)

var path = "./Data/"

func LoadStockList() [][]string {
	fileName := path + "stockList.csv"
	fs1, _ := os.Open(fileName)
	r1 := csv.NewReader(fs1)
	content, err := r1.ReadAll()
	if err != nil {
		logger.Errorf("can not readall, err is %+v", err)
	}

	for index, row := range content {
		if index != 0 {
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

		//调试阶段，只取前10个数据
		if index == 10 {
			break
		}
	}

	logger.Infof("stock2 list size %d", len(content))
	return content
}

func LoadFromCsv(code string) (stockData StockData) {
	fileName := path + code + "_ALL.csv"
	fs1, _ := os.Open(fileName)
	r1 := csv.NewReader(fs1)
	content, err := r1.ReadAll()
	if err != nil {
		logger.Errorf("can not readall, err is %+v", err)
	}

	priceEndY := 0.0
	Interest := 1.0
	for index, row := range content {
		stock := new(StockDataDay)
		stock.Index = index
		if index != 0 {
			//丢弃停牌数据
			if row[4] == "0" {
				continue
			}
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
			stockData.DayDatas = append(stockData.DayDatas, stock)
			priceEndY = float64(priceEnd)
		}
	}
	//return content
	return
}
