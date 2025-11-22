package selectors

// AllMarketSelector 全市场选股器，返回所有股票代码
type AllMarketSelector struct{}

// NewAllMarketSelector 创建全市场选股器
func NewAllMarketSelector() *AllMarketSelector {
	return &AllMarketSelector{}
}

// SelectStocks 返回所有股票代码
func (s *AllMarketSelector) SelectStocks(allCodes []string) []string {
	return allCodes
}

// GetName 获取选股器名称
func (s *AllMarketSelector) GetName() string {
	return "全市场"
}
