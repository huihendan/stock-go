package signals

import (
	"fmt"
	"stock-go/stockData"
	"stock-go/stockStrategy"
)

// BuyHighSellLowSignal 追涨杀跌信号生成器
// 买入条件：价格达到过去N天的最高价
// 卖出条件：止损或超过最大持有天数
type BuyHighSellLowSignal struct {
	LookbackDays    int     // 回看天数，默认300
	SellDropPercent float64 // 止损百分比，默认0.06（6%）
	MaxHoldDays     int     // 最大持有天数，默认15

	// 内部状态(防止未来函数)
	historyPrices []float32 // 历史价格队列
}

// NewBuyHighSellLowSignal 创建追涨杀跌信号生成器
func NewBuyHighSellLowSignal(lookbackDays int, sellDropPercent float64, maxHoldDays int) *BuyHighSellLowSignal {
	return &BuyHighSellLowSignal{
		LookbackDays:    lookbackDays,
		SellDropPercent: sellDropPercent,
		MaxHoldDays:     maxHoldDays,
		historyPrices:   make([]float32, 0, 500),
	}
}

// Reset 重置策略状态（每只票票回测前调用）
func (sg *BuyHighSellLowSignal) Reset() {
	sg.historyPrices = sg.historyPrices[:0] // 清空历史价格
}

// ProcessDay 处理单日数据，返回交易信号
func (sg *BuyHighSellLowSignal) ProcessDay(
	dayData *stockData.StockDataDay,
	dateIndex int,
	position *stockStrategy.Position,
) int {
	currentPrice := dayData.PriceA

	// 1. 更新历史价格队列
	sg.historyPrices = append(sg.historyPrices, currentPrice)

	// 2. 数据不足，无操作
	if len(sg.historyPrices) < sg.LookbackDays {
		return 0
	}

	// 3. 判断买入信号(空仓时)
	if position == nil {
		if sg.isBuySignal(currentPrice) {
			return 1 // 买入
		}
		return 0
	}

	// 4. 更新持仓的最高价
	if currentPrice > position.HighestPrice {
		position.HighestPrice = currentPrice
	}

	// 5. 判断卖出信号(持仓时)
	if sg.isSellSignal(currentPrice, position) {
		return -1 // 卖出
	}

	return 0
}

// isBuySignal 买入信号：前一天达到历史最高价（避免未来函数）
// 逻辑：判断前一天是否创新高，如果是则次日开盘买入
func (sg *BuyHighSellLowSignal) isBuySignal(currentPrice float32) bool {
	historyLen := len(sg.historyPrices)

	// 数据不足（至少需要回看天数的数据）
	if historyLen < sg.LookbackDays {
		return false
	}

	// 计算回看窗口的起始位置
	startIndex := historyLen - sg.LookbackDays

	// 找出回看窗口内的最高价（不包括昨天，即historyLen-1）
	var highestPrice float32 = 0
	for i := startIndex; i < historyLen-1; i++ {
		if sg.historyPrices[i] > highestPrice {
			highestPrice = sg.historyPrices[i]
		}
	}

	// 判断昨天（historyLen-1）的价格是否达到或超过之前的最高价
	// 注意：此时 sg.historyPrices 还未包含今天的价格
	if historyLen > 0 {
		yesterdayPrice := sg.historyPrices[historyLen-1]
		// 昨天创新高（允许0.5%误差），则今天开盘买入
		return yesterdayPrice >= highestPrice*0.995
	}

	return false
}

// isSellSignal 卖出信号：止损或超过最大持有天数
func (sg *BuyHighSellLowSignal) isSellSignal(currentPrice float32, position *stockStrategy.Position) bool {
	// 条件1: 相对买入价的跌幅止损
	dropPercent := float64(position.BuyPrice-currentPrice) / float64(position.BuyPrice)
	if dropPercent >= sg.SellDropPercent {
		return true
	}

	// 条件2: 相对最高价的回撤止损
	if position.HighestPrice > position.BuyPrice {
		drawdownPercent := float64(position.HighestPrice-currentPrice) / float64(position.HighestPrice)
		if drawdownPercent >= sg.SellDropPercent {
			return true
		}
	}

	// 条件3: 超过最大持有天数
	if position.HoldDays >= sg.MaxHoldDays {
		return true
	}

	return false
}

// GetName 获取信号生成器名称
func (sg *BuyHighSellLowSignal) GetName() string {
	return fmt.Sprintf("追涨杀跌(%d天新高,止损%.1f%%,最多持有%d天)",
		sg.LookbackDays, sg.SellDropPercent*100, sg.MaxHoldDays)
}
