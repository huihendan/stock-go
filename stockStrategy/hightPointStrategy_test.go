package stockStrategy

import (
	"stock/logger"
	"stock/stockData"
	"testing"
)

func TestHighPointStrategy(t *testing.T) {
	logger.Info("TestHighPointStrategy start")
	//1. 加载所有股票列表
	stockData.LoadStockList()
	//2. 加载股票数据
	stockData.LoadDataOneByOne()

	for _, stock := range stockData.Stocks {
		HighPointStrategy(stock.Code)
	}

	logger.Info("TestHighPointStrategy end")
}
