package stockData

import (
	"stock/logger"
	"stock/utils"
	"time"
)

// 原始数据
var StocksRaw = make(map[string]*StockInfo)

func LoadRawDataOneByCode(code string) (stock *StockInfo) {
	start1 := time.Now()
	defer utils.CostTime(start1)
	stock = &StockInfo{
		Code:  code,
		Datas: LoadFromCsv(code),
	}
	StocksRaw[code] = stock
	return
}

func LoadRawDataAll() {
	start1 := time.Now()
	defer utils.CostTime(start1)
	for code := range StockList {
		StocksRaw[code] = LoadRawDataOneByCode(code)
	}
	logger.Infof("StocksRaw loaded size=%d", len(StocksRaw))
}

func GetStockRawBycode(code string) *StockInfo {
	if code == "" {
		logger.Warnf("stock代码为空")
		return nil
	}

	stock := StocksRaw[code]
	if stock == nil {
		stock = LoadRawDataOneByCode(code)
		if stock == nil {
			logger.Warnf("加载stock %s 数据失败", code)
		}
	}
	return stock
}
