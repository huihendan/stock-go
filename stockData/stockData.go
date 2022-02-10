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

func LoadAllData() {
	start1 := time.Now()
	defer utils.CostTime(start1)
	LoadStockList()
	for code, name := range StockList {
		stockInfo := new(StockInfo)
		stockInfo.Datas = LoadFromCsv(code)
		stockInfo.Code = code
		stockInfo.Name = name
		Stocks[code] = stockInfo
	}
	logger.Infof("Stocks size[%d]", len(Stocks))
}
func DealStocksPoints() {
	for _, stock := range Stocks {
		stock.DealStockPoints()
	}
	logger.Infof("DealStockData finish")
}

func DealStocksSections() {
	for _, stock := range Stocks {
		stock.DealStockSession(0)
	}
	logger.Infof("DealStockData finish")
}

func Start() {
	LoadAllData()
	DealStocksPoints()
	DealStocksSections()
	logger.Infof("finish")
}
