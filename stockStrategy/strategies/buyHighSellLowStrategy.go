package strategies

import (
	"fmt"
	"stock-go/stockStrategy"
	"stock-go/stockStrategy/selectors"
	"stock-go/stockStrategy/signals"
)

// BuyHighSellLowStrategy 追涨杀跌策略
// 选股：选择近期出现高点的股票
// 交易：价格创新高买入，止损或超时卖出
type BuyHighSellLowStrategy struct {
	selector  stockStrategy.StockSelector
	signalGen stockStrategy.SignalGenerator
}

// NewBuyHighSellLowStrategy 创建追涨杀跌策略（使用默认参数）
func NewBuyHighSellLowStrategy() *BuyHighSellLowStrategy {
	return &BuyHighSellLowStrategy{
		selector:  selectors.NewHighPointSelector(500, 15),
		signalGen: signals.NewBuyHighSellLowSignal(300, 0.06, 15),
	}
}

// NewBuyHighSellLowStrategyWithParams 创建追涨杀跌策略（自定义参数）
func NewBuyHighSellLowStrategyWithParams(
	selectorLookback int,
	selectorRecent int,
	signalLookback int,
	signalDropPercent float64,
	signalMaxHoldDays int,
) *BuyHighSellLowStrategy {
	return &BuyHighSellLowStrategy{
		selector:  selectors.NewHighPointSelector(selectorLookback, selectorRecent),
		signalGen: signals.NewBuyHighSellLowSignal(signalLookback, signalDropPercent, signalMaxHoldDays),
	}
}

// GetSelector 获取选股器
func (s *BuyHighSellLowStrategy) GetSelector() stockStrategy.StockSelector {
	return s.selector
}

// GetSignalGenerator 获取信号生成器
func (s *BuyHighSellLowStrategy) GetSignalGenerator() stockStrategy.SignalGenerator {
	return s.signalGen
}

// GetName 获取策略名称
func (s *BuyHighSellLowStrategy) GetName() string {
	return fmt.Sprintf("策略1[%s + %s]",
		s.selector.GetName(), s.signalGen.GetName())
}
