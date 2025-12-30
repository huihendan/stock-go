// TimeBasedBacktestEngine 基于时间流逝的回测引擎
//
// 设计理念：
// 1. 按日期顺序模拟时间流逝
// 2. 每天遍历所有票票，检查买入/卖出信号
// 3. 严格的资金管理，避免超额使用资金
// 4. 支持多票票同时持仓
//
// 核心流程：
// for each_day:
//     1. 更新所有持仓（检查卖出信号、止损、持有时间等）
//     2. 卖出需要卖出的持仓
//     3. 检查所有候选票票的买入信号
//     4. 根据资金情况和持仓限制，买入符合条件的票票
//     5. 记录当日持仓价值和总资产

package tradeTest

import (
	"fmt"
	"sort"
	"stock-go/logger"
	"stock-go/stockData"
	"stock-go/stockStrategy"
	"stock-go/stockStrategy/signals"
)

// TimeBasedBacktestEngine 基于时间流逝的回测引擎
type TimeBasedBacktestEngine struct {
	initialCash     float64                // 初始资金
	strategy        stockStrategy.Strategy // 交易策略
	maxPositions    int                    // 最大持仓数量
	cashPerPosition float64                // 每个持仓的资金比例（0-1）

	// 手续费配置
	commissionRate float64 // 佣金费率（买入和卖出都收取）
	stampTaxRate   float64 // 印花税率（仅卖出时收取）
	transferFeeRate float64 // 过户费率（买入和卖出都收取）
	minCommission  float64 // 最低佣金（单笔交易）

	// 回测状态
	currentDate      string                                   // 当前日期
	wallet           *Wallet                                  // 钱包
	positions        map[string]*PositionState                // 持仓状态（key: 票票代码）
	signalGenerators map[string]stockStrategy.SignalGenerator // 每只票票的信号生成器
	buyCooldowns     map[string]int                           // 买入冷却期（key: 票票代码, value: 冷却结束的dayIndex）

	// 回测数据
	allStockData map[string]*stockData.StockInfo // 所有票票的数据
	tradingDays  []string                        // 所有交易日（排序后）

	// 回测结果
	dailyEquity   []DailyEquity // 每日权益
	tradeRecords  []TradeRecord // 交易记录
	totalFees     float64       // 总手续费（佣金+印花税+过户费）
}

// Wallet 钱包
type Wallet struct {
	Cash          float64 // 现金
	TotalAssets   float64 // 总资产（现金+持仓市值）
	PositionValue float64 // 持仓市值
}

// PositionState 持仓状态
type PositionState struct {
	Code         string                        // 票票代码
	Name         string                        // 票票名称
	StockNum     int                           // 持仓数量
	BuyPrice     float64                       // 买入价格
	BuyDate      string                        // 买入日期
	BuyIndex     int                           // 买入时的数据索引
	HoldDays     int                           // 持有天数
	HighestPrice float64                       // 持有期间最高价
	CurrentPrice float64                       // 当前价格
	SignalGen    stockStrategy.SignalGenerator // 该持仓的信号生成器
}

// DailyEquity 每日权益
type DailyEquity struct {
	Date          string  // 日期
	Cash          float64 // 现金
	PositionValue float64 // 持仓市值
	TotalAssets   float64 // 总资产
	PositionCount int     // 持仓数量
}

// TradeRecord 交易记录
type TradeRecord struct {
	Code       string  // 票票代码
	Name       string  // 票票名称
	Action     string  // 动作：buy/sell
	Date       string  // 日期
	Price      float64 // 价格
	StockNum   int     // 数量
	Amount     float64 // 金额（不含手续费）
	Commission float64 // 佣金
	StampTax   float64 // 印花税（仅卖出）
	TransferFee float64 // 过户费
	TotalFee   float64 // 总手续费（佣金+印花税+过户费）
	Cash       float64 // 交易后现金
	Reason     string  // 原因（买入信号、止损、止盈等）
}

// NewTimeBasedBacktestEngine 创建基于时间流逝的回测引擎
func NewTimeBasedBacktestEngine(
	initialCash float64,
	strategy stockStrategy.Strategy,
	maxPositions int,
	cashPerPosition float64,
) *TimeBasedBacktestEngine {
	return &TimeBasedBacktestEngine{
		initialCash:     initialCash,
		strategy:        strategy,
		maxPositions:    maxPositions,
		cashPerPosition: cashPerPosition,
		// 手续费配置（A股标准费率）
		commissionRate: 0.0001, // 万1佣金
		stampTaxRate:   0.0005, // 万5印花税（仅卖出）
		transferFeeRate: 0.00001, // 10万分之1过户费（买入和卖出都收取）
		minCommission:  5.0,    // 最低佣金5元
		wallet: &Wallet{
			Cash:        initialCash,
			TotalAssets: initialCash,
		},
		positions:        make(map[string]*PositionState),
		signalGenerators: make(map[string]stockStrategy.SignalGenerator),
		buyCooldowns:     make(map[string]int),
		allStockData:     make(map[string]*stockData.StockInfo),
		dailyEquity:      make([]DailyEquity, 0),
		tradeRecords:     make([]TradeRecord, 0),
		totalFees:        0,
	}
}

// Run 执行回测
func (e *TimeBasedBacktestEngine) Run() *TimeBasedBacktestResult {
	logger.Infof("========================================")
	logger.Infof("基于时间流逝的回测引擎")
	logger.Infof("策略: %s", e.strategy.GetName())
	logger.Infof("初始资金: %.2f", e.initialCash)
	logger.Infof("最大持仓数: %d", e.maxPositions)
	logger.Infof("每仓位资金: %.1f%%", e.cashPerPosition*100)
	logger.Infof("========================================")

	// 1. 加载所有票票数据
	if err := e.loadAllStockData(); err != nil {
		logger.Infof("加载票票数据失败: %v", err)
		return nil
	}

	// 2. 构建交易日列表
	e.buildTradingDays()
	logger.Infof("回测时间范围: %s 至 %s (共 %d 个交易日)",
		e.tradingDays[0], e.tradingDays[len(e.tradingDays)-1], len(e.tradingDays))

	// 3. 初始选股
	selectedCodes := e.performStockSelection(500) // 使用第500天的数据进行初始选股
	logger.Infof("初始选股结果: %d 只票票", len(selectedCodes))

	// 4. 逐日模拟
	e.runDailySimulation(selectedCodes)

	// 5. 强制平仓所有持仓
	e.closeAllPositions()

	// 6. 生成回测结果
	result := e.generateResult()
	e.printSummary(result)

	return result
}

// loadAllStockData 加载所有票票数据
func (e *TimeBasedBacktestEngine) loadAllStockData() error {
	allCodes := getAllStockCodes()

	loadedCount := 0
	for _, code := range allCodes {
		stockInfo := stockData.GetStockRawBycode(code)
		if stockInfo != nil && len(stockInfo.Datas.DayDatas) >= 500 {
			e.allStockData[code] = stockInfo
			loadedCount++
		}
	}

	logger.Infof("成功加载 %d 只票票数据", loadedCount)
	return nil
}

// buildTradingDays 构建交易日列表
// 从所有票票数据中提取共同的交易日
func (e *TimeBasedBacktestEngine) buildTradingDays() {
	dateMap := make(map[string]bool)

	// 收集所有日期
	for _, stockInfo := range e.allStockData {
		for _, dayData := range stockInfo.Datas.DayDatas {
			dateMap[dayData.DataStr] = true
		}
	}

	// 转换为切片并排序
	dates := make([]string, 0, len(dateMap))
	for date := range dateMap {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	// 只保留有足够数据的日期（至少500天历史）
	if len(dates) > 500 {
		e.tradingDays = dates[500:]
	} else {
		e.tradingDays = dates
	}
}

// performStockSelection 执行选股
func (e *TimeBasedBacktestEngine) performStockSelection(dateIndex int) []string {
	allCodes := make([]string, 0, len(e.allStockData))
	for code := range e.allStockData {
		allCodes = append(allCodes, code)
	}
	sort.Strings(allCodes)

	return e.strategy.GetSelector().SelectStocksAtDate(allCodes, dateIndex)
}

// runDailySimulation 逐日模拟
func (e *TimeBasedBacktestEngine) runDailySimulation(candidateCodes []string) {
	for dayIdx, date := range e.tradingDays {
		e.currentDate = date

		// 1. 处理卖出（必须先卖后买）
		e.processSells(dayIdx)

		// 2. 处理买入
		e.processBuys(candidateCodes, dayIdx)

		// 3. 更新持仓价格和统计
		e.updatePositions(dayIdx)

		// 4. 记录每日权益
		e.recordDailyEquity()

		// 每100个交易日输出一次进度
		if (dayIdx+1)%100 == 0 {
			logger.Infof("回测进度: %d/%d 日期: %s 持仓: %d 总资产: %.2f",
				dayIdx+1, len(e.tradingDays), date, len(e.positions), e.wallet.TotalAssets)
		}
	}
}

// processSells 处理卖出
// 直接实现卖出逻辑，不调用 ProcessDay，以便正确更新持仓状态
func (e *TimeBasedBacktestEngine) processSells(dayIdx int) {
	// 用于存储需要卖出的持仓及其原因
	type sellInfo struct {
		pos    *PositionState
		reason string
	}
	sellList := make([]sellInfo, 0)

	for _, pos := range e.positions {
		// 获取当天数据
		dayData := e.getDayData(pos.Code, e.currentDate)
		if dayData == nil {
			continue
		}

		currentPrice := float64(dayData.PriceBegin)

		// 更新最高价（在卖出判断之前）
		if currentPrice > pos.HighestPrice {
			pos.HighestPrice = currentPrice
		}

		// 判断卖出信号
		// 这里直接实现卖出逻辑，而不是调用 ProcessDay
		// 参考 BuyHighSellLowSignal 的卖出条件
		shouldSell := false
		sellReason := ""

		// 从信号生成器获取参数
		// 注意：这里需要类型断言，假设使用的是 BuyHighSellLowSignal
		// 如果使用其他信号生成器，需要相应调整
		if bhslSignal, ok := pos.SignalGen.(*signals.BuyHighSellLowSignal); ok {
			// 条件1: 相对买入价的跌幅止损
			dropPercent := (pos.BuyPrice - currentPrice) / pos.BuyPrice
			if dropPercent >= bhslSignal.SellDropPercent {
				shouldSell = true
				sellReason = fmt.Sprintf("止损(跌幅%.2f%%)", dropPercent*100)
			}

			// 条件2: 相对最高价的回撤止损
			if !shouldSell && pos.HighestPrice > pos.BuyPrice {
				drawdownPercent := (pos.HighestPrice - currentPrice) / pos.HighestPrice
				if drawdownPercent >= bhslSignal.SellDropPercent {
					shouldSell = true
					sellReason = fmt.Sprintf("回撤止损(回撤%.2f%%)", drawdownPercent*100)
				}
			}

			// 条件3: 超过最大持有天数
			if !shouldSell && pos.HoldDays >= bhslSignal.MaxHoldDays {
				shouldSell = true
				sellReason = fmt.Sprintf("持有超过%d天", bhslSignal.MaxHoldDays)
			}
		}

		if shouldSell {
			sellList = append(sellList, sellInfo{
				pos:    pos,
				reason: sellReason,
			})
		} else {
			// 不卖出，更新持有天数
			pos.HoldDays++
		}

		// 更新当前价格
		pos.CurrentPrice = currentPrice
	}

	// 执行卖出
	for _, info := range sellList {
		reason := info.reason
		if reason == "" {
			reason = "卖出信号"
		}
		e.executeSell(info.pos, reason)
	}
}

// processBuys 处理买入
func (e *TimeBasedBacktestEngine) processBuys(candidateCodes []string, dayIdx int) {
	// 如果已达到最大持仓数，不再买入
	if len(e.positions) >= e.maxPositions {
		return
	}

	// 收集所有买入信号
	buySignals := make([]string, 0)

	for _, code := range candidateCodes {
		// 已持仓的不再买入
		if _, exists := e.positions[code]; exists {
			continue
		}

		// 检查是否在冷却期内
		if cooldownEnd, inCooldown := e.buyCooldowns[code]; inCooldown {
			if dayIdx < cooldownEnd {
				// 仍在冷却期内，跳过
				continue
			} else {
				// 冷却期已结束，清除记录
				delete(e.buyCooldowns, code)
			}
		}

		// 检查最近5天是否有高风险（单日涨幅超过7%）
		if e.checkRecentHighRisk(code, e.currentDate) {
			// 存在高风险，加入冷却期50天
			e.buyCooldowns[code] = dayIdx + 50
			continue
		}

		// 获取当天数据
		dayData := e.getDayData(code, e.currentDate)
		if dayData == nil {
			continue
		}

		// 获取或创建信号生成器
		signalGen := e.getOrCreateSignalGenerator(code)

		// 检查买入信号
		signal := signalGen.ProcessDay(dayData, dayIdx, nil)

		if signal == 1 {
			buySignals = append(buySignals, code)
		}
	}

	// 执行买入
	for _, code := range buySignals {
		// 检查持仓数限制
		if len(e.positions) >= e.maxPositions {
			break
		}

		// 检查是否有足够现金买入（至少能买一手）
		dayData := e.getDayData(code, e.currentDate)
		if dayData == nil {
			continue
		}
		price := float64(dayData.PriceBegin)
		minCost := price * 100 // 一手的成本

		// 如果现金不足，跳过本次买入
		if e.wallet.Cash < minCost {
			continue
		}

		// 尝试买入
		e.executeBuy(code, dayIdx, 0)
	}
}

// executeBuy 执行买入
// 全仓买入策略：使用所有可用现金买入
func (e *TimeBasedBacktestEngine) executeBuy(code string, dayIdx int, _ float64) {
	dayData := e.getDayData(code, e.currentDate)
	if dayData == nil {
		return
	}

	stockInfo := e.allStockData[code]
	price := float64(dayData.PriceBegin)

	// 检查是否触及涨跌停板
	if e.checkPriceLimit(code, e.currentDate, price) {
		// 触及涨停板，买入可能无法成交
		// 将该票票加入冷却期，50天内禁止买入
		e.buyCooldowns[code] = dayIdx + 50
		return
	}

	// 全仓买入：使用所有可用现金
	cashToUse := e.wallet.Cash

	// 计算买入数量（整手），需要预留手续费
	// 总成本 = 价格 * 数量 + 佣金 + 过户费
	// 佣金 = max(价格 * 数量 * 佣金率, 最低佣金)
	// 过户费 = 价格 * 数量 * 过户费率
	// 为简化计算，先估算可买入的数量，然后验证是否有足够现金
	stockNum := int(cashToUse / price / 100) * 100
	if stockNum < 100 {
		return // 资金不足买一手
	}

	// 计算实际成本和手续费
	amount := price * float64(stockNum)
	commission := e.calculateCommission(amount)
	transferFee := e.calculateTransferFee(amount)
	totalCost := amount + commission + transferFee

	// 如果总成本超过现金，减少买入数量
	for totalCost > e.wallet.Cash && stockNum >= 100 {
		stockNum -= 100
		amount = price * float64(stockNum)
		commission = e.calculateCommission(amount)
		transferFee = e.calculateTransferFee(amount)
		totalCost = amount + commission + transferFee
	}

	if stockNum < 100 {
		return // 资金不足买一手（含手续费）
	}

	// 扣除资金（包括手续费）
	e.wallet.Cash -= totalCost
	e.totalFees += commission + transferFee

	// 创建持仓
	signalGen := e.getOrCreateSignalGenerator(code)
	pos := &PositionState{
		Code:         code,
		Name:         stockInfo.Name,
		StockNum:     stockNum,
		BuyPrice:     price,
		BuyDate:      e.currentDate,
		BuyIndex:     dayIdx,
		HoldDays:     0,
		HighestPrice: price,
		CurrentPrice: price,
		SignalGen:    signalGen,
	}

	e.positions[code] = pos

	// 记录交易
	e.tradeRecords = append(e.tradeRecords, TradeRecord{
		Code:       code,
		Name:       stockInfo.Name,
		Action:     "buy",
		Date:       e.currentDate,
		Price:      price,
		StockNum:   stockNum,
		Amount:     amount,
		Commission: commission,
		StampTax:   0,
		TransferFee: transferFee,
		TotalFee:   commission + transferFee,
		Cash:       e.wallet.Cash,
		Reason:     "买入信号",
	})
}

// executeSell 执行卖出
func (e *TimeBasedBacktestEngine) executeSell(pos *PositionState, reason string) {
	dayData := e.getDayData(pos.Code, e.currentDate)
	if dayData == nil {
		return
	}

	price := float64(dayData.PriceBegin)

	// 检查是否触及涨跌停板
	if e.checkPriceLimit(pos.Code, e.currentDate, price) {
		// 触及跌停板，卖出可能无法成交，放弃卖出（持仓继续保留）
		return
	}

	// 计算卖出金额和手续费
	amount := price * float64(pos.StockNum)
	commission := e.calculateCommission(amount)
	stampTax := e.calculateStampTax(amount)
	transferFee := e.calculateTransferFee(amount)
	totalFee := commission + stampTax + transferFee
	netAmount := amount - totalFee // 实际到手金额

	// 增加资金（扣除手续费后）
	e.wallet.Cash += netAmount
	e.totalFees += totalFee

	// 记录交易
	e.tradeRecords = append(e.tradeRecords, TradeRecord{
		Code:       pos.Code,
		Name:       pos.Name,
		Action:     "sell",
		Date:       e.currentDate,
		Price:      price,
		StockNum:   pos.StockNum,
		Amount:     amount,
		Commission: commission,
		StampTax:   stampTax,
		TransferFee: transferFee,
		TotalFee:   totalFee,
		Cash:       e.wallet.Cash,
		Reason:     reason,
	})

	// 删除持仓
	delete(e.positions, pos.Code)
}

// sellHalfPosition 卖出一半仓位
// 返回卖出获得的资金，如果卖出失败则返回0
func (e *TimeBasedBacktestEngine) sellHalfPosition(pos *PositionState, reason string) float64 {
	dayData := e.getDayData(pos.Code, e.currentDate)
	if dayData == nil {
		return 0
	}

	price := float64(dayData.PriceBegin)

	// 检查是否触及涨跌停板
	if e.checkPriceLimit(pos.Code, e.currentDate, price) {
		// 触及跌停板，卖出可能无法成交
		return 0
	}

	// 计算卖出一半的数量（向下取整到整手）
	halfNum := (pos.StockNum / 2 / 100) * 100
	if halfNum < 100 {
		// 如果一半不足一手，则不卖出
		return 0
	}

	// 计算卖出金额和手续费
	amount := price * float64(halfNum)
	commission := e.calculateCommission(amount)
	stampTax := e.calculateStampTax(amount)
	transferFee := e.calculateTransferFee(amount)
	totalFee := commission + stampTax + transferFee
	netAmount := amount - totalFee // 实际到手金额

	// 增加资金（扣除手续费后）
	e.wallet.Cash += netAmount
	e.totalFees += totalFee

	// 更新持仓数量
	pos.StockNum -= halfNum

	// 记录交易
	e.tradeRecords = append(e.tradeRecords, TradeRecord{
		Code:       pos.Code,
		Name:       pos.Name,
		Action:     "sell",
		Date:       e.currentDate,
		Price:      price,
		StockNum:   halfNum,
		Amount:     amount,
		Commission: commission,
		StampTax:   stampTax,
		TransferFee: transferFee,
		TotalFee:   totalFee,
		Cash:       e.wallet.Cash,
		Reason:     reason,
	})

	return netAmount
}

// updatePositions 更新持仓价格
func (e *TimeBasedBacktestEngine) updatePositions(dayIdx int) {
	totalValue := 0.0

	for _, pos := range e.positions {
		dayData := e.getDayData(pos.Code, e.currentDate)
		if dayData != nil {
			pos.CurrentPrice = float64(dayData.PriceEnd)
			totalValue += pos.CurrentPrice * float64(pos.StockNum)
		}
	}

	e.wallet.PositionValue = totalValue
	e.wallet.TotalAssets = e.wallet.Cash + e.wallet.PositionValue
}

// recordDailyEquity 记录每日权益
func (e *TimeBasedBacktestEngine) recordDailyEquity() {
	e.dailyEquity = append(e.dailyEquity, DailyEquity{
		Date:          e.currentDate,
		Cash:          e.wallet.Cash,
		PositionValue: e.wallet.PositionValue,
		TotalAssets:   e.wallet.TotalAssets,
		PositionCount: len(e.positions),
	})
}

// closeAllPositions 强制平仓所有持仓
func (e *TimeBasedBacktestEngine) closeAllPositions() {
	for _, pos := range e.positions {
		e.executeSell(pos, "回测结束强制平仓")
	}
}

// getDayData 获取指定日期的数据
func (e *TimeBasedBacktestEngine) getDayData(code, date string) *stockData.StockDataDay {
	stockInfo, exists := e.allStockData[code]
	if !exists {
		return nil
	}

	for _, dayData := range stockInfo.Datas.DayDatas {
		if dayData.DataStr == date {
			return dayData
		}
	}

	return nil
}

// getPreviousDayData 获取前一个交易日的数据
func (e *TimeBasedBacktestEngine) getPreviousDayData(code, currentDate string) *stockData.StockDataDay {
	stockInfo, exists := e.allStockData[code]
	if !exists {
		return nil
	}

	// 找到当前日期的索引
	var currentIndex = -1
	for i, dayData := range stockInfo.Datas.DayDatas {
		if dayData.DataStr == currentDate {
			currentIndex = i
			break
		}
	}

	// 如果找不到当前日期或者是第一天，返回nil
	if currentIndex <= 0 {
		return nil
	}

	return stockInfo.Datas.DayDatas[currentIndex-1]
}

// checkPriceLimit 检查价格是否触及涨跌停板
// 返回 true 表示触及涨跌停，交易可能无法成交
func (e *TimeBasedBacktestEngine) checkPriceLimit(code, currentDate string, currentPrice float64) bool {
	prevDayData := e.getPreviousDayData(code, currentDate)
	if prevDayData == nil {
		// 无法获取前一天数据，保守起见不交易
		return true
	}

	prevClosePrice := float64(prevDayData.PriceEnd)
	if prevClosePrice <= 0 {
		return true
	}

	// 计算涨跌幅
	changePercent := (currentPrice - prevClosePrice) / prevClosePrice

	// 涨跌幅超过9.5%视为触及涨跌停
	return changePercent > 0.095 || changePercent < -0.095
}

// findWeakestPosition 找出最弱的持仓（盈利最少或亏损最大）
// 策略：优先卖出表现差的票票，保留表现好的
func (e *TimeBasedBacktestEngine) findWeakestPosition() *PositionState {
	if len(e.positions) == 0 {
		return nil
	}

	var weakest *PositionState
	minProfitRate := 999999.0 // 设置一个很大的初始值

	for _, pos := range e.positions {
		// 获取当前价格
		dayData := e.getDayData(pos.Code, e.currentDate)
		if dayData == nil {
			continue
		}

		currentPrice := float64(dayData.PriceBegin)

		// 计算盈利率
		profitRate := (currentPrice - pos.BuyPrice) / pos.BuyPrice

		// 找出盈利率最低的（可能是亏损最大的）
		if profitRate < minProfitRate {
			minProfitRate = profitRate
			weakest = pos
		}
	}

	return weakest
}

// checkRecentHighRisk 检查最近5天是否有高风险（单日涨幅超过7%）
// 返回 true 表示存在高风险
func (e *TimeBasedBacktestEngine) checkRecentHighRisk(code, currentDate string) bool {
	stockInfo, exists := e.allStockData[code]
	if !exists {
		return true // 无数据，保守起见判定为高风险
	}

	// 找到当前日期的索引
	var currentIndex = -1
	for i, dayData := range stockInfo.Datas.DayDatas {
		if dayData.DataStr == currentDate {
			currentIndex = i
			break
		}
	}

	if currentIndex < 5 {
		// 数据不足5天，无法判断
		return false
	}

	// 检查最近5天（不包括当天）
	for i := currentIndex - 5; i < currentIndex; i++ {
		if i < 0 {
			continue
		}

		currentDay := stockInfo.Datas.DayDatas[i]

		// 获取前一天数据计算涨幅
		if i > 0 {
			prevDay := stockInfo.Datas.DayDatas[i-1]
			prevClose := float64(prevDay.PriceEnd)
			currentClose := float64(currentDay.PriceEnd)

			if prevClose > 0 {
				changePercent := (currentClose - prevClose) / prevClose

				// 单日涨幅超过7%
				if changePercent > 0.07 {
					return true
				}
			}
		}
	}

	return false
}

// calculateCommission 计算佣金（买入和卖出都需要）
func (e *TimeBasedBacktestEngine) calculateCommission(amount float64) float64 {
	commission := amount * e.commissionRate
	if commission < e.minCommission {
		commission = e.minCommission
	}
	return commission
}

// calculateStampTax 计算印花税（仅卖出时收取）
func (e *TimeBasedBacktestEngine) calculateStampTax(amount float64) float64 {
	return amount * e.stampTaxRate
}

// calculateTransferFee 计算过户费（买入和卖出都需要）
func (e *TimeBasedBacktestEngine) calculateTransferFee(amount float64) float64 {
	return amount * e.transferFeeRate
}

// getOrCreateSignalGenerator 获取或创建信号生成器
func (e *TimeBasedBacktestEngine) getOrCreateSignalGenerator(code string) stockStrategy.SignalGenerator {
	if gen, exists := e.signalGenerators[code]; exists {
		return gen
	}

	// 创建新的信号生成器
	gen := e.strategy.GetSignalGenerator()
	gen.Reset()
	e.signalGenerators[code] = gen

	return gen
}

// TimeBasedBacktestResult 回测结果
type TimeBasedBacktestResult struct {
	InitialCash    float64
	FinalCash      float64
	FinalAssets    float64
	TotalReturn    float64
	TotalReturnPct float64

	TotalTrades int
	BuyCount    int
	SellCount   int
	WinCount    int
	LoseCount   int
	WinRate     float64

	DailyEquity  []DailyEquity
	TradeRecords []TradeRecord

	// 新增统计
	MaxDrawdown  float64
	SharpeRatio  float64
	MaxPositions int
	AvgHoldDays  float64
	TotalFees    float64 // 总手续费（佣金+印花税+过户费）
}

// generateResult 生成回测结果
func (e *TimeBasedBacktestEngine) generateResult() *TimeBasedBacktestResult {
	result := &TimeBasedBacktestResult{
		InitialCash:  e.initialCash,
		FinalCash:    e.wallet.Cash,
		FinalAssets:  e.wallet.TotalAssets,
		DailyEquity:  e.dailyEquity,
		TradeRecords: e.tradeRecords,
		TotalFees:    e.totalFees,
	}

	// 计算总收益
	result.TotalReturn = result.FinalAssets - result.InitialCash
	result.TotalReturnPct = (result.TotalReturn / result.InitialCash) * 100

	// 统计交易
	buyTrades := make(map[string]*TradeRecord) // key: code_date

	for i := range e.tradeRecords {
		record := &e.tradeRecords[i]

		if record.Action == "buy" {
			result.BuyCount++
			key := record.Code + "_" + record.Date
			buyTrades[key] = record
		} else if record.Action == "sell" {
			result.SellCount++

			// 找到对应的买入记录
			// 简化处理：假设先买后卖
			for _, buyRecord := range buyTrades {
				if buyRecord.Code == record.Code {
					profit := record.Amount - buyRecord.Amount
					if profit > 0 {
						result.WinCount++
					} else {
						result.LoseCount++
					}
					break
				}
			}
		}
	}

	result.TotalTrades = result.SellCount // 完成的交易数
	if result.TotalTrades > 0 {
		result.WinRate = float64(result.WinCount) / float64(result.TotalTrades) * 100
	}

	// 计算最大回撤
	result.MaxDrawdown = e.calculateMaxDrawdown()

	return result
}

// calculateMaxDrawdown 计算最大回撤
func (e *TimeBasedBacktestEngine) calculateMaxDrawdown() float64 {
	if len(e.dailyEquity) == 0 {
		return 0
	}

	maxAssets := e.dailyEquity[0].TotalAssets
	maxDrawdown := 0.0

	for _, equity := range e.dailyEquity {
		if equity.TotalAssets > maxAssets {
			maxAssets = equity.TotalAssets
		}

		drawdown := (maxAssets - equity.TotalAssets) / maxAssets
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}

	return maxDrawdown * 100
}

// printSummary 打印总结
func (e *TimeBasedBacktestEngine) printSummary(result *TimeBasedBacktestResult) {
	logger.Infof("")
	logger.Infof("========================================")
	logger.Infof("回测总结")
	logger.Infof("========================================")
	logger.Infof("初始资金: %.2f", result.InitialCash)
	logger.Infof("最终资金: %.2f", result.FinalCash)
	logger.Infof("最终总资产: %.2f", result.FinalAssets)
	logger.Infof("总收益: %.2f (%.2f%%)", result.TotalReturn, result.TotalReturnPct)
	logger.Infof("总手续费: %.2f (佣金+印花税+过户费)", result.TotalFees)
	logger.Infof("手续费占初始资金比例: %.2f%%", (result.TotalFees/result.InitialCash)*100)
	logger.Infof("最大回撤: %.2f%%", result.MaxDrawdown)
	logger.Infof("")
	logger.Infof("交易统计:")
	logger.Infof("  总交易次数: %d", result.TotalTrades)
	logger.Infof("  买入次数: %d", result.BuyCount)
	logger.Infof("  卖出次数: %d", result.SellCount)
	logger.Infof("  盈利次数: %d", result.WinCount)
	logger.Infof("  亏损次数: %d", result.LoseCount)
	logger.Infof("  胜率: %.2f%%", result.WinRate)
	logger.Infof("========================================")
}
