package stockData

import (
	"github.com/apache/dubbo-go/common/logger"
	"stock/utils"
	"time"
)

type StockInfo struct {
	Code  string
	Name  string
	Datas StockData
}

type StockData struct {
	DayDatas []*StockDataDay
	Points   []*StockDataDay
	Sections []*StockDataDay
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
	logger.Infof("Stocks size[%d]", len(Stocks))
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
	logger.Infof("Stocks size[%d]", len(Stocks))
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
	logger.Infof("Stocks size[%d]", len(Stocks))
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

	logger.Infof("Stocks size[%d]", len(Stocks))
}

func DealStocksPointsByCode(code string) {
	stockInfo, ok := Stocks[code]
	if !ok {
		logger.Errorf("DealStocksPointsByCode： %s failed", code)
	}
	stockInfo.DealStockPoints()
	logger.Infof("DealStocksPointsByCode: %s finish", code)
}

func DealAllStocksPoints() {
	for _, stock := range Stocks {
		stock.DealStockPoints()
	}
	logger.Infof("DealStockData finish")
}

func DealStocksSectionsByCode(code string) {
	stockInfo, ok := Stocks[code]
	if !ok {
		logger.Errorf("DealStocksSectionsByCode： %s failed", code)
	}

	stockInfo.DealStockSession(0)
	logger.Infof("DealStocksSectionsByCode: %s finish", code)
}
func DealAllStocksSections() {
	for _, stock := range Stocks {
		stock.DealStockSession(0)
	}
	logger.Infof("DealStockData finish")
}

func Start() {
	//1. 加载所有股票列表
	LoadStockList()
	//2. 加载股票数据
	go LoadDataOneByOne()
	//LoadAllData()
	//DealAllStocksPoints()
	//DealAllStocksSections()
	logger.Infof("finish")
}
