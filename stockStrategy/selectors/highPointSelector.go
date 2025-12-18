package selectors

import (
	"fmt"
	"stock-go/stockData"
)

// HighPointSelector 高点选股器
// 选择在指定回看期内，最近N天出现过最高点的票票
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

// SelectStocks 不做任何筛选，返回所有票票列表
// 已废弃，请使用 SelectStocksAtDate
func (s *HighPointSelector) SelectStocks(allCodes []string) []string {
	return allCodes
}

// SelectStocksAtDate 不做任何筛选，返回所有票票列表（避免未来数据）
func (s *HighPointSelector) SelectStocksAtDate(allCodes []string, endIndex int) []string {
	// 不做筛选，直接返回所有票票
	return allCodes
}

// isRecentHighPointAtDate 检查票票在指定时间点是否满足高点条件
// endIndex: 数据截止索引（包含），只使用 [0, endIndex] 的数据
func (s *HighPointSelector) isRecentHighPointAtDate(stock *stockData.StockInfo, endIndex int) bool {
	dayDatas := stock.Datas.DayDatas

	// 数据不足
	if len(dayDatas) == 0 || endIndex >= len(dayDatas) {
		return false
	}

	// 可用的数据长度
	availableLength := endIndex + 1
	if availableLength < s.LookbackDays {
		return false
	}

	// 计算回看窗口
	startIndex := availableLength - s.LookbackDays
	checkEndIndex := availableLength // 不包含
	recentStartIndex := availableLength - s.RecentDays

	var maxPrice float32 = 0
	var maxIndex int = -1

	// 找出回看窗口内的最高价格及其索引
	for i := startIndex; i < checkEndIndex; i++ {
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
