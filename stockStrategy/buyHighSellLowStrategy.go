package stockStrategy

import (
	"fmt"
	globaldefine "stock/globalDefine"
	"stock/stockData"
)

// BuyHighSellLowStrategy 追涨杀跌策略1
// 达到设定的时间（默认300天）内最高值时买入
type BuyHighSellLowStrategy struct {
	LookbackDays    int       // 回看天数，默认300
	SellDropPercent float64   // 止损跌幅百分比，默认0.06（6%）
	MaxHoldDays     int       // 最大持有天数，默认15
	historyPrices   []float32 // 历史价格队列（内部维护，防止访问未来数据）
}

// NewBuyHighSellLowStrategy 创建策略1实例
func NewBuyHighSellLowStrategy() *BuyHighSellLowStrategy {
	return &BuyHighSellLowStrategy{
		LookbackDays:    300,
		SellDropPercent: 0.06,
		MaxHoldDays:     15,
		historyPrices:   make([]float32, 0),
	}
}

// NewBuyHighSellLowStrategyWithConfig 使用自定义配置创建策略实例
func NewBuyHighSellLowStrategyWithConfig(lookbackDays int, sellDropPercent float64, maxHoldDays int) *BuyHighSellLowStrategy {
	return &BuyHighSellLowStrategy{
		LookbackDays:    lookbackDays,
		SellDropPercent: sellDropPercent,
		MaxHoldDays:     maxHoldDays,
		historyPrices:   make([]float32, 0),
	}
}

// 暂时采用随机策略
func (strategy *BuyHighSellLowStrategy) DealSelectStockCodes() (stockCodes []string) {
	// 加载股票列表和数据
	stockList := stockData.LoadPreStockList()
	for stockCode, _ := range stockList {
		stock := stockData.GetstockBycode(stockCode)
		if stock == nil {
			continue
		}
		if stock.Datas.DayDatas == nil {
			continue
		}
		if len(stock.Datas.DayDatas) < strategy.LookbackDays {
			continue
		}
		stockCodes = append(stockCodes, stockCode)
	}

	return stockCodes
}

// DealStrategy 执行策略
func (strategy *BuyHighSellLowStrategy) DealStrategy(code string) (operates map[string]OperateRecord) {
	operates = make(map[string]OperateRecord)

	// 通过GetstockBycode获取股票数据
	stock := stockData.GetstockBycode(code)
	if stock == nil {
		return operates
	}

	dayDatas := stock.Datas.DayDatas

	// 数据不足，无法执行策略
	if len(dayDatas) < strategy.LookbackDays {
		return operates
	}

	// 重置历史价格队列
	strategy.historyPrices = make([]float32, 0)

	// 初始化持仓状态为空仓
	position := globaldefine.Position{
		StockCode: code,
		StockName: stock.Name,
		StockNum:  0,
		Profit:    0,
		BuyPrice:  0,
		SellPrice: 0,
		BuyDate:   "",
	}

	// 遍历所有数据，模拟实时接收价格
	for i := 0; i < len(dayDatas); i++ {
		if dayDatas[i] == nil {
			continue
		}

		currentPrice := dayDatas[i].PriceA

		// 将当前价格加入历史队列（模拟实时获取数据）
		strategy.historyPrices = append(strategy.historyPrices, currentPrice)

		// 只有历史数据足够时才开始判断买卖
		if len(strategy.historyPrices) < strategy.LookbackDays {
			continue
		}

		if position.Status == 2 { // 空仓状态，判断是否买入
			shouldBuy := strategy.DealStrategyBuy(currentPrice, i)

			if shouldBuy {
				// 记录买入信息
				position.Status = 1 // 1-持仓
				position.BuyPrice = currentPrice
				position.BuyDate = dayDatas[i].DataStr
				position.BuyIndex = i

				// 创建买入操作记录
				buyOperate := Operate{
					OperateType: 1,
					BuyPrice:    float64(currentPrice),
					OperateDate: dayDatas[i].DataStr,
					StockCode:   code,
					StockNum:    100,
				}

				// 创建操作记录（买入时还没有卖出信息）
				recordKey := fmt.Sprintf("%s_%d", code, i)
				operates[recordKey] = OperateRecord{
					StockCode:   code,
					StockName:   stock.Name,
					StockNum:    100,
					BuyOperate:  buyOperate,
					Status:      1, // 1-已买入
					OperateDate: dayDatas[i].DataStr,
				}
			}

		} else if position.Status == 1 { // 持仓状态，判断是否卖出
			shouldSell := strategy.DealStrategySell(currentPrice, i, &position)

			if shouldSell {
				// 创建卖出操作记录
				sellOperate := Operate{
					OperateType: 2,
					SellPrice:   float64(currentPrice),
					OperateDate: dayDatas[i].DataStr,
					StockCode:   code,
					StockNum:    100,
				}

				// 更新操作记录
				recordKey := fmt.Sprintf("%s_%d", code, position.BuyIndex)
				if record, exists := operates[recordKey]; exists {
					record.SellOperate = sellOperate
					record.Status = 2 // 2-已卖出
					record.Profit = float64(currentPrice-position.BuyPrice) * record.StockNum
					operates[recordKey] = record
				}

				// 重置钱包状态为空仓
				position.Status = 2
				position.BuyPrice = 0
				position.BuyDate = ""
				position.BuyIndex = 0
			}
		}
	}

	return operates
}

// DealStrategyBuy 判断是否符合买入条件
func (strategy *BuyHighSellLowStrategy) DealStrategyBuy(price float32, dateIndex int) (buy bool) {
	// 判断 price 是否是当前300天（LookbackDays）内的最高价格
	// 如果是就买入，如果不是则不买入

	// 使用内部维护的历史价格队列，只能访问到当前时刻之前的数据
	historyLen := len(strategy.historyPrices)
	if historyLen < strategy.LookbackDays {
		return false
	}

	// 计算回看窗口的起始索引
	startIndex := historyLen - strategy.LookbackDays

	// 找出回看窗口内的最高价格（不包括当前价格）
	var highestPrice float32 = 0.0
	for i := startIndex; i < historyLen-1; i++ {
		if strategy.historyPrices[i] > highestPrice {
			highestPrice = strategy.historyPrices[i]
		}
	}

	// 当前价格是否达到或超过历史最高价（允许0.5%的误差）
	if price >= highestPrice*0.995 {
		return true
	}

	return false
}

// DealStrategySell 判断是否符合卖出条件
func (strategy *BuyHighSellLowStrategy) DealStrategySell(price float32, dataIndex int, position *globaldefine.Position) (sell bool) {
	/*卖出条件：
	1、距离买入价下跌超过6%则卖出（止损）
	2、买入超过15天，则卖出
	*/

	// 条件1：止损 - 距离买入价下跌超过设定百分比
	dropPercent := (position.BuyPrice - price) / position.BuyPrice
	if dropPercent >= float32(strategy.SellDropPercent) {
		return true
	}

	// 条件2：持有时间达到最大持有天数
	holdDays := dataIndex - position.BuyIndex
	if holdDays >= strategy.MaxHoldDays {
		return true
	}

	return false
}
