package stockData

import (
	"fmt"
	"log/slog"
	"testing"
)

func TestLoadData(t *testing.T) {
	code := "sz.002236"
	stockData := LoadFromCsv(code)
	fmt.Println("stockData len", len(stockData.DayDatas))
}

func TestLoadStockList(t *testing.T) {
	//stockList := LoadStockList()
	LoadPreStockList()

	slog.Info("stock list size", "size", len(StockList))
}
