package stockStrategy

import (
	"stock/logger"
	"stock/painter/plot"
	"stock/stockData"
	"testing"
	"time"
)

func TestHighPointStrategy(t *testing.T) {
	logger.Info("TestHighPointStrategy start")
	//1. 加载所有stock列表
	stockData.LoadPreStockList()
	//2. 加载stock数据
	stockData.LoadDataOneByOne()

	for _, stock := range stockData.Stocks {
		HighPointStrategy(stock.Code)
	}

	logger.Info("TestHighPointStrategy end")
}

func TestHighPointStrategyDayByDay(t *testing.T) {
	startTime := time.Now()
	logger.Info("TestHighPointStrategyDayByDay start")
	//1. 加载所有stock列表
	stockData.LoadPreStockList()
	//2. 加载stock数据
	stockData.LoadDataOneByOne()

	for _, stock := range stockData.Stocks {
		if stock.DealSessionHighPointDayByDay() {
			//painter.PaintStockKline(stock.Code)
			plot.PlotPoints(stock.Code)

		}
	}

	logger.Infof("TestHighPointStrategyDayByDay end cost time:%dms", time.Since(startTime).Milliseconds())
}

func TestHighPointStrategyDayByDay2(t *testing.T) {
	startTime := time.Now()
	logger.Info("TestHighPointStrategyDayByDay2 start")
	//1. 加载所有stock列表
	stockData.LoadPreStockList()
	//2. 加载stock数据
	stockData.LoadDataOneByOne()

	for _, stock := range stockData.Stocks {
		if stock.DealSessionHighPointDayByDay2() {
			//painter.PaintStockKline(stock.Code)
		}
	}

	logger.Infof("TestHighPointStrategyDayByDay2 end cost time:%dms", time.Since(startTime).Milliseconds())
}
