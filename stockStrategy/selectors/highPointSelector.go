package selectors

import (
	"fmt"
	"stock-go/stockData"
)

// HighPointSelector 高点选股器
// 选择在指定回看期内，最近N天出现过最高点的股票
type HighPointSelector struct {
	LookbackDays int // 回看天数，默认500
	RecentDays   int // 最近N天内出现高点，默认15
}

// NewHighPointSelector 创建高点选股器
func NewHighPointSelector(lookbackDays, recentDays int) *HighPointSelector {
	return &HighPointSelector{
		LookbackDays: lookbackDays,
		RecentDays:   recentDays,
	}
}

// SelectStocks 选择符合高点条件的股票
func (s *HighPointSelector) SelectStocks(allCodes []string) []string {
	var selected []string

	for _, code := range allCodes {
		stockInfo := stockData.GetStockRawBycode(code)
		if stockInfo == nil {
			continue
		}

		// 检查最近是否出现高点
		if s.isRecentHighPoint(stockInfo) {
			selected = append(selected, code)
		}
	}

	return selected
}

// isRecentHighPoint 检查股票是否在最近RecentDays天内出现了LookbackDays的最高点
func (s *HighPointSelector) isRecentHighPoint(stock *stockData.StockInfo) bool {
	dayDatas := stock.Datas.DayDatas
	if len(dayDatas) < s.LookbackDays {
		return false
	}

	// 检查最后RecentDays天内是否有新高点
	startIndex := len(dayDatas) - s.LookbackDays
	endIndex := len(dayDatas)
	recentStartIndex := len(dayDatas) - s.RecentDays

	var maxPrice float32 = 0
	var maxIndex int = -1

	// 找出回看窗口内的最高价格及其索引
	for i := startIndex; i < endIndex; i++ {
		if dayDatas[i].PriceA > maxPrice {
			maxPrice = dayDatas[i].PriceA
			maxIndex = i
		}
	}

	// 最高点出现在最近RecentDays天内
	return maxIndex >= recentStartIndex
}

// GetName 获取选股器名称
func (s *HighPointSelector) GetName() string {
	return fmt.Sprintf("%d天高点选股(最近%d天)", s.LookbackDays, s.RecentDays)
}
