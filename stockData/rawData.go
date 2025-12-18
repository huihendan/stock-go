package stockData

import (
	"stock-go/logger"
	"stock-go/utils"
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
		logger.Warnf("票票代码为空")
		return nil
	}

	stock := StocksRaw[code]
	if stock == nil {
		stock = LoadRawDataOneByCode(code)
		if stock == nil {
			logger.Warnf("加载票票 %s 数据失败", code)
		}
	}
	return stock
}

// 批量加载所有原始数据（逐个加载）
func LoadRawDataOneByOne() {
	start1 := time.Now()
	defer utils.CostTime(start1)
	for code := range StockList {
		LoadRawDataOneByCode(code)
	}
	logger.Infof("StocksRaw loaded one by one, size=%d", len(StocksRaw))
}

// 重新加载所有原始数据
func ReLoadRawDataAll() {
	start1 := time.Now()
	defer utils.CostTime(start1)
	// 清空现有数据
	StocksRaw = make(map[string]*StockInfo)
	// 重新加载
	LoadRawDataAll()
	logger.Infof("StocksRaw reloaded, size=%d", len(StocksRaw))
}

// 清空原始数据缓存
func ClearRawData() {
	start1 := time.Now()
	defer utils.CostTime(start1)
	oldSize := len(StocksRaw)
	StocksRaw = make(map[string]*StockInfo)
	logger.Infof("StocksRaw cleared, old size=%d", oldSize)
}

// 启动函数：加载原始数据
func StartRaw() {
	// 1. 加载所有票票列表
	LoadPreStockList()
	// 2. 异步加载原始数据
	go LoadRawDataOneByOne()
	logger.Info("Raw stock data loading started")
}

// 从原始数据获取并处理单个票票数据到 Stocks
func GetProcessedStockFromRaw(code string) *StockInfo {
	if code == "" {
		logger.Warnf("票票代码为空")
		return nil
	}

	// 先检查 Stocks 中是否已有处理过的数据
	if stock, ok := Stocks[code]; ok {
		return stock
	}

	// 从 StocksRaw 获取原始数据
	rawStock := GetStockRawBycode(code)
	if rawStock == nil {
		logger.Warnf("获取原始数据失败 code=%s", code)
		return nil
	}

	// 深拷贝原始数据到新的 StockInfo
	stock := &StockInfo{
		Code: rawStock.Code,
		Name: rawStock.Name,
		Datas: StockData{
			DayDatas: make(StockDataDayList, len(rawStock.Datas.DayDatas)),
		},
	}

	// 复制日数据
	for i, day := range rawStock.Datas.DayDatas {
		newDay := *day
		stock.Datas.DayDatas[i] = &newDay
	}

	// 处理数据
	stock.DealStockPoints()
	stock.DealStockSession(0)

	// 存入 Stocks
	Stocks[code] = stock
	return stock
}

// 从原始数据批量处理所有票票数据到 Stocks
func ProcessAllStocksFromRaw() {
	start1 := time.Now()
	defer utils.CostTime(start1)

	for code := range StocksRaw {
		GetProcessedStockFromRaw(code)
	}
	logger.Infof("Processed all stocks from raw data, size=%d", len(Stocks))
}

// 获取原始数据数量
func GetRawDataCount() int {
	return len(StocksRaw)
}

// 检查原始数据是否已加载
func IsRawDataLoaded(code string) bool {
	_, ok := StocksRaw[code]
	return ok
}

// 获取所有已加载的原始数据代码列表
func GetRawDataCodes() []string {
	codes := make([]string, 0, len(StocksRaw))
	for code := range StocksRaw {
		codes = append(codes, code)
	}
	return codes
}
