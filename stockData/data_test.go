package stockData

import (
	"fmt"
	"github.com/apache/dubbo-go/common/logger"
	"testing"
)

func TestLoadData(t *testing.T) {
	code := "sz.002236"
	stockData := LoadFromCsv(code)
	var stock1 StockInfo
	stock1.Code = code
	fmt.Println("stockData len", len(stockData.DayDatas))
}

func TestLoadStockList(t *testing.T) {
	//stockList := LoadStockList()
	LoadStockList()

	logger.Infof("stock list size[%d]", len(StockList))
}
