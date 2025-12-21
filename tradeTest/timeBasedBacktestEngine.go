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
	"stock-go/stockData"
	"stock-go/stockStrategy"
)

// TimeBasedBacktestEngine 基于时间流逝的回测引擎
type TimeBasedBacktestEngine struct {
	initialCash      float64                    // 初始资金
	strategy         stockStrategy.Strategy      // 交易策略
	maxPositions     int                         // 最大持仓数量
	cashPerPosition  float64                     // 每个持仓的资金比例（0-1）

	// 回测状态
	currentDate      string                              // 当前日期
	wallet           *Wallet                             // 钱包
	positions        map[string]*PositionState           // 持仓状态（key: 票票代码）
	signalGenerators map[string]stockStrategy.SignalGenerator // 每只票票的信号生成器

	// 回测数据
	allStockData     map[string]*stockData.StockInfo    // 所有票票的数据
	tradingDays      []string                           // 所有交易日（排序后）

	// 回测结果
	dailyEquity      []DailyEquity                      // 每日权益
	tradeRecords     []TradeRecord                      // 交易记录
}

// Wallet 钱包
type Wallet struct {
	Cash            float64  // 现金
	TotalAssets     float64  // 总资产（现金+持仓市值）
	PositionValue   float64  // 持仓市值
}

// PositionState 持仓状态
type PositionState struct {
	Code         string                       // 票票代码
	Name         string                       // 票票名称
	StockNum     int                          // 持仓数量
	BuyPrice     float64                      // 买入价格
	BuyDate      string                       // 买入日期
	BuyIndex     int                          // 买入时的数据索引
	HoldDays     int                          // 持有天数
	HighestPrice float64                      // 持有期间最高价
	CurrentPrice float64                      // 当前价格
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
	Code      string  // 票票代码
	Name      string  // 票票名称
	Action    string  // 动作：buy/sell
	Date      string  // 日期
	Price     float64 // 价格
	StockNum  int     // 数量
	Amount    float64 // 金额
	Cash      float64 // 交易后现金
	Reason    string  // 原因（买入信号、止损、止盈等）
}

// NewTimeBasedBacktestEngine 创建基于时间流逝的回测引擎
func NewTimeBasedBacktestEngine(
	initialCash float64,
	strategy stockStrategy.Strategy,
	maxPositions int,
	cashPerPosition float64,
) *TimeBasedBacktestEngine {
	return &TimeBasedBacktestEngine{
		initialCash:      initialCash,
		strategy:         strategy,
		maxPositions:     maxPositions,
		cashPerPosition:  cashPerPosition,
		wallet: &Wallet{
			Cash:        initialCash,
			TotalAssets: initialCash,
		},
		positions:        make(map[string]*PositionState),
		signalGenerators: make(map[string]stockStrategy.SignalGenerator),
		allStockData:     make(map[string]*stockData.StockInfo),
		dailyEquity:      make([]DailyEquity, 0),
		tradeRecords:     make([]TradeRecord, 0),
	}
}

// Run 执行回测
func (e *TimeBasedBacktestEngine) Run() *TimeBasedBacktestResult {
	fmt.Printf("========================================\n")
	fmt.Printf("基于时间流逝的回测引擎\n")
	fmt.Printf("策略: %s\n", e.strategy.GetName())
	fmt.Printf("初始资金: %.2f\n", e.initialCash)
	fmt.Printf("最大持仓数: %d\n", e.maxPositions)
	fmt.Printf("每仓位资金: %.1f%%\n", e.cashPerPosition*100)
	fmt.Printf("========================================\n")

	// 1. 加载所有票票数据
	if err := e.loadAllStockData(); err != nil {
		fmt.Printf("加载票票数据失败: %v\n", err)
		return nil
	}

	// 2. 构建交易日列表
	e.buildTradingDays()
	fmt.Printf("回测时间范围: %s 至 %s (共 %d 个交易日)\n",
		e.tradingDays[0], e.tradingDays[len(e.tradingDays)-1], len(e.tradingDays))

	// 3. 初始选股
	selectedCodes := e.performStockSelection(500) // 使用第500天的数据进行初始选股
	fmt.Printf("初始选股结果: %d 只票票\n", len(selectedCodes))

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

	fmt.Printf("成功加载 %d 只票票数据\n", loadedCount)
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
			fmt.Printf("回测进度: %d/%d 日期: %s 持仓: %d 总资产: %.2f\n",
				dayIdx+1, len(e.tradingDays), date, len(e.positions), e.wallet.TotalAssets)
		}
	}
}

// processSells 处理卖出
func (e *TimeBasedBacktestEngine) processSells(dayIdx int) {
	sellList := make([]*PositionState, 0)

	for _, pos := range e.positions {
		// 获取当天数据
		dayData := e.getDayData(pos.Code, e.currentDate)
		if dayData == nil {
			continue
		}

		// 转换为策略Position
		strategyPos := &stockStrategy.Position{
			StockCode:    pos.Code,
			StockName:    pos.Name,
			StockNum:     pos.StockNum,
			BuyPrice:     float32(pos.BuyPrice),
			BuyDate:      pos.BuyDate,
			BuyIndex:     pos.BuyIndex,
			HoldDays:     pos.HoldDays,
			HighestPrice: float32(pos.HighestPrice),
		}

		// 获取卖出信号
		signal := pos.SignalGen.ProcessDay(dayData, dayIdx, strategyPos)

		if signal == -1 {
			sellList = append(sellList, pos)
		} else {
			// 更新持有天数和最高价
			pos.HoldDays++
			if float64(dayData.PriceHigh) > pos.HighestPrice {
				pos.HighestPrice = float64(dayData.PriceHigh)
			}
		}
	}

	// 执行卖出
	for _, pos := range sellList {
		e.executeSell(pos, "卖出信号")
	}
}

// processBuys 处理买入
func (e *TimeBasedBacktestEngine) processBuys(candidateCodes []string, dayIdx int) {
	// 如果已达到最大持仓数，不再买入
	if len(e.positions) >= e.maxPositions {
		return
	}

	// 计算可用于单个持仓的资金
	cashPerPos := e.wallet.Cash * e.cashPerPosition

	// 收集所有买入信号
	buySignals := make([]string, 0)

	for _, code := range candidateCodes {
		// 已持仓的不再买入
		if _, exists := e.positions[code]; exists {
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

	// 执行买入（按资金和持仓限制）
	for _, code := range buySignals {
		if len(e.positions) >= e.maxPositions {
			break
		}

		e.executeBuy(code, dayIdx, cashPerPos)
	}
}

// executeBuy 执行买入
func (e *TimeBasedBacktestEngine) executeBuy(code string, dayIdx int, cashToUse float64) {
	dayData := e.getDayData(code, e.currentDate)
	if dayData == nil {
		return
	}

	stockInfo := e.allStockData[code]
	price := float64(dayData.PriceBegin)

	// 计算买入数量（整手）
	stockNum := int(cashToUse / price / 100) * 100
	if stockNum < 100 {
		return // 资金不足买一手
	}

	// 检查资金是否足够
	actualCost := price * float64(stockNum)
	if actualCost > e.wallet.Cash {
		return
	}

	// 扣除资金
	e.wallet.Cash -= actualCost

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
		Code:     code,
		Name:     stockInfo.Name,
		Action:   "buy",
		Date:     e.currentDate,
		Price:    price,
		StockNum: stockNum,
		Amount:   actualCost,
		Cash:     e.wallet.Cash,
		Reason:   "买入信号",
	})
}

// executeSell 执行卖出
func (e *TimeBasedBacktestEngine) executeSell(pos *PositionState, reason string) {
	dayData := e.getDayData(pos.Code, e.currentDate)
	if dayData == nil {
		return
	}

	price := float64(dayData.PriceBegin)
	amount := price * float64(pos.StockNum)

	// 增加资金
	e.wallet.Cash += amount

	// 记录交易
	e.tradeRecords = append(e.tradeRecords, TradeRecord{
		Code:     pos.Code,
		Name:     pos.Name,
		Action:   "sell",
		Date:     e.currentDate,
		Price:    price,
		StockNum: pos.StockNum,
		Amount:   amount,
		Cash:     e.wallet.Cash,
		Reason:   reason,
	})

	// 删除持仓
	delete(e.positions, pos.Code)
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
	InitialCash   float64
	FinalCash     float64
	FinalAssets   float64
	TotalReturn   float64
	TotalReturnPct float64

	TotalTrades   int
	BuyCount      int
	SellCount     int
	WinCount      int
	LoseCount     int
	WinRate       float64

	DailyEquity   []DailyEquity
	TradeRecords  []TradeRecord

	// 新增统计
	MaxDrawdown   float64
	SharpeRatio   float64
	MaxPositions  int
	AvgHoldDays   float64
}

// generateResult 生成回测结果
func (e *TimeBasedBacktestEngine) generateResult() *TimeBasedBacktestResult {
	result := &TimeBasedBacktestResult{
		InitialCash:  e.initialCash,
		FinalCash:    e.wallet.Cash,
		FinalAssets:  e.wallet.TotalAssets,
		DailyEquity:  e.dailyEquity,
		TradeRecords: e.tradeRecords,
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
	fmt.Printf("\n========================================\n")
	fmt.Printf("回测总结\n")
	fmt.Printf("========================================\n")
	fmt.Printf("初始资金: %.2f\n", result.InitialCash)
	fmt.Printf("最终资金: %.2f\n", result.FinalCash)
	fmt.Printf("最终总资产: %.2f\n", result.FinalAssets)
	fmt.Printf("总收益: %.2f (%.2f%%)\n", result.TotalReturn, result.TotalReturnPct)
	fmt.Printf("最大回撤: %.2f%%\n", result.MaxDrawdown)
	fmt.Printf("\n交易统计:\n")
	fmt.Printf("  总交易次数: %d\n", result.TotalTrades)
	fmt.Printf("  买入次数: %d\n", result.BuyCount)
	fmt.Printf("  卖出次数: %d\n", result.SellCount)
	fmt.Printf("  盈利次数: %d\n", result.WinCount)
	fmt.Printf("  亏损次数: %d\n", result.LoseCount)
	fmt.Printf("  胜率: %.2f%%\n", result.WinRate)
	fmt.Printf("========================================\n")
}
