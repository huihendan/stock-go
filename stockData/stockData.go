package stockData

import (
	"stock/logger"
	"stock/utils"
	"time"
)

type StockInfo struct {
	Code  string
	Name  string
	Datas StockData
}

type StockDataDayList []*StockDataDay

type StockData struct {
	DayDatas   StockDataDayList
	Points     StockDataDayList
	Sections   StockDataDayList
	HighPoints StockDataDayList
	LowPoints  StockDataDayList
}

type StockDataDay struct {
	Index      int
	DataStr    string
	PointType  int
	Trend      int
	PriceA     float32
	PriceBegin float32
	PriceEnd   float32
	PriceHigh  float32
	PriceLow   float32
	PriceShow  float32
}

var StockList = make(map[string]string)
var Stocks = make(map[string]*StockInfo)

func GetstockBycode(code string) *StockInfo {
	stock := Stocks[code]
	if stock == nil {
		stock = LoadDataOneByCode(code)
	}
	return stock
}

func LoadDataOneByOne() {
	start1 := time.Now()
	defer utils.CostTime(start1)
	for code, name := range StockList {
		Stocks[code] = &StockInfo{
			Code:  code,
			Name:  name,
			Datas: LoadFromCsv(code),
		}
		DealStocksPointsByCode(code)
		DealStocksSectionsByCode(code)
	}
	logger.Infof("Stocks loaded size=%d", len(Stocks))
}

func LoadDataOneByCode(code string) (stock *StockInfo) {
	start1 := time.Now()
	defer utils.CostTime(start1)
	stock = &StockInfo{
		Code:  code,
		Datas: LoadFromCsv(code),
	}
	Stocks[code] = stock
	DealStocksPointsByCode(code)
	DealStocksSectionsByCode(code)
	logger.Infof("Stocks loaded size=%d", len(Stocks))
	return stock
}

func LoadAllData() {
	start1 := time.Now()
	defer utils.CostTime(start1)
	//LoadStockList()
	for code, name := range StockList {
		stockInfo := new(StockInfo)
		stockInfo.Datas = LoadFromCsv(code)
		stockInfo.Code = code
		stockInfo.Name = name
		Stocks[code] = stockInfo
	}
	logger.Infof("Stocks loaded size=%d", len(Stocks))
}

func LoadDataByCode(code string) {
	start1 := time.Now()
	defer utils.CostTime(start1)
	stockInfo, ok := Stocks[code]
	if ok {
		stockInfo.Datas = LoadFromCsv(code)
	} else {
		stockInfo := new(StockInfo)
		stockInfo.Datas = LoadFromCsv(code)
		stockInfo.Code = code
		Stocks[code] = stockInfo
	}

	logger.Infof("Stocks loaded size=%d", len(Stocks))
}

func DealStocksPointsByCode(code string) {
	stockInfo, ok := Stocks[code]
	if !ok {
		logger.Errorf("DealStocksPointsByCode failed code=%s", code)
	}
	stockInfo.DealStockPoints()
	logger.Infof("DealStocksPointsByCode finished code=%s", code)
}

func DealAllStocksPoints() {
	for _, stock := range Stocks {
		stock.DealStockPoints()
	}
	logger.Info("DealStockData finished")
}

func DealStocksSectionsByCode(code string) {
	stockInfo, ok := Stocks[code]
	if !ok {
		logger.Errorf("DealStocksSectionsByCode failed code=%s", code)
	}

	stockInfo.DealStockSession(0)
	logger.Infof("DealStocksSectionsByCode finished code=%s", code)
}

func DealAllStocksSections() {
	for _, stock := range Stocks {
		stock.DealStockSession(0)
	}
	logger.Info("DealStockData finished")
}

func Start() {
	//1. 加载所有股票列表
	LoadStockList()
	//2. 加载股票数据
	go LoadDataOneByOne()
	//LoadAllData()
	//DealAllStocksPoints()
	//DealAllStocksSections()
	logger.Info("Stock data loading started")
}
