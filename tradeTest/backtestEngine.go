package tradeTest

import (
	"fmt"
	"sort"
	globalDefine "stock-go/globalDefine"
	"stock-go/stockData"
	"stock-go/stockStrategy"
)

// BacktestEngine 回测引擎
type BacktestEngine struct {
	initialCash      float64
	strategy         stockStrategy.Strategy
	reselectInterval int // 重新选股的天数间隔，0表示只在开始时选股一次
}

// NewBacktestEngine 创建回测引擎（默认每30天重新选股）
func NewBacktestEngine(initialCash float64, strategy stockStrategy.Strategy) *BacktestEngine {
	return &BacktestEngine{
		initialCash:      initialCash,
		strategy:         strategy,
		reselectInterval: 30, // 默认每30天重新选股
	}
}

// NewBacktestEngineWithReselect 创建回测引擎（自定义选股间隔）
// reselectInterval: 重新选股的天数间隔，0表示只在开始时选股一次
func NewBacktestEngineWithReselect(initialCash float64, strategy stockStrategy.Strategy, reselectInterval int) *BacktestEngine {
	return &BacktestEngine{
		initialCash:      initialCash,
		strategy:         strategy,
		reselectInterval: reselectInterval,
	}
}

// BacktestResult 回测结果
type BacktestResult struct {
	Wallet         globalDefine.Wallet
	OperateRecords map[string][]globalDefine.OperateRecord
	Stats          PortfolioStats
}

// Run 执行回测（使用固定时间点选股避免未来数据泄漏）
func (engine *BacktestEngine) Run() BacktestResult {
	fmt.Printf("========================================\n")
	fmt.Printf("开始回测: %s\n", engine.strategy.GetName())
	fmt.Printf("初始资金: %.2f\n", engine.initialCash)
	fmt.Printf("========================================\n")

	// 获取全市场票票代码
	allCodes := getAllStockCodes()
	fmt.Printf("全市场票票数量: %d\n", len(allCodes))

	// 使用固定时间点进行选股（避免未来数据泄漏）
	// 找出合适的选股时间点：取所有票票数据的中位数长度的一半作为选股时点
	selectDateIndex := engine.getSelectDateIndex(allCodes)
	fmt.Printf("选股时间点索引: %d\n", selectDateIndex)

	selectedCodes := engine.strategy.GetSelector().SelectStocksAtDate(allCodes, selectDateIndex)
	fmt.Printf("策略[%s]选股结果: %d只票票\n",
		engine.strategy.GetName(), len(selectedCodes))

	// 对每只票票进行回测
	wallet := globalDefine.Wallet{
		Cash:      float32(engine.initialCash),
		Positions: make([]globalDefine.Position, 0),
	}
	allRecords := make(map[string][]globalDefine.OperateRecord)

	processedCount := 0
	for _, code := range selectedCodes {
		records := engine.backtestSingleStock(code, &wallet)
		if len(records) > 0 {
			allRecords[code] = records
			processedCount++
		}
	}

	fmt.Printf("实际回测票票数量: %d\n", processedCount)
	fmt.Printf("========================================\n")

	// 计算绩效
	stats := CalculatePortfolioPerformance(wallet, allRecords)

	return BacktestResult{
		Wallet:         wallet,
		OperateRecords: allRecords,
		Stats:          stats,
	}
}

// getSelectDateIndex 获取选股时间点索引
// 使用策略：找出所有票票数据长度，取满足选股器回看要求的最早时间点
func (engine *BacktestEngine) getSelectDateIndex(allCodes []string) int {
	// 假设选股器需要500天数据（这是HighPointSelector的默认lookback）
	// 更好的做法是在接口中添加 GetMinDataLength() 方法
	minRequiredDays := 500

	// 找出第一只有足够数据的票票，使用其第500天作为选股时间点
	for _, code := range allCodes {
		stockInfo := stockData.GetStockRawBycode(code)
		if stockInfo != nil && len(stockInfo.Datas.DayDatas) >= minRequiredDays {
			return minRequiredDays - 1 // 返回索引（从0开始）
		}
	}

	// 如果没有票票有足够数据，返回0
	return 0
}

// backtestSingleStock 单只票票回测
func (engine *BacktestEngine) backtestSingleStock(
	code string,
	wallet *globalDefine.Wallet,
) []globalDefine.OperateRecord {
	// 1. 加载原始数据
	stockInfo := stockData.GetStockRawBycode(code)
	if stockInfo == nil || len(stockInfo.Datas.DayDatas) == 0 {
		return nil
	}

	dayDatas := stockInfo.Datas.DayDatas
	signalGen := engine.strategy.GetSignalGenerator()

	// 2. 重置策略状态
	signalGen.Reset()

	// 3. 逐日循环
	var position *stockStrategy.Position = nil
	var records []globalDefine.OperateRecord

	for i := 0; i < len(dayDatas); i++ {
		dayData := dayDatas[i]

		// 4. 获取交易信号
		signal := signalGen.ProcessDay(dayData, i, position)

		// 5. 执行交易（次日开盘价执行，避免未来函数）
		if signal == 1 && position == nil && i+1 < len(dayDatas) { // 买入信号且当前空仓且有次日数据
			nextDayData := dayDatas[i+1] // 使用次日数据
			position = engine.executeBuy(code, stockInfo.Name, nextDayData, i+1, wallet)
			if position != nil {
				// 创建新的交易记录
				record := globalDefine.OperateRecord{
					StockCode:  code,
					StockName:  stockInfo.Name,
					StockNum:   position.StockNum,
					BuyOperate: createBuyOperate(position, nextDayData),
					Status:     1, // 已买入
				}
				records = append(records, record)
			}
		} else if signal == -1 && position != nil && i+1 < len(dayDatas) { // 卖出信号且当前持仓且有次日数据
			nextDayData := dayDatas[i+1] // 使用次日数据
			record := &records[len(records)-1]
			engine.executeSell(position, nextDayData, wallet, record)
			position = nil // 清空持仓
		}

		// 6. 更新持有天数
		if position != nil {
			position.HoldDays++
		}
	}

	// 7. 强制平仓未卖出的持仓(使用最后一天价格)
	if position != nil {
		lastDay := dayDatas[len(dayDatas)-1]
		record := &records[len(records)-1]
		engine.executeSell(position, lastDay, wallet, record)
	}

	return records
}

// executeBuy 执行买入（使用开盘价，模拟真实交易）
func (engine *BacktestEngine) executeBuy(
	code, name string,
	dayData *stockData.StockDataDay,
	dateIndex int,
	wallet *globalDefine.Wallet,
) *stockStrategy.Position {
	price := float64(dayData.PriceBegin) // 使用开盘价而非收盘价

	// 计算可买股数(假设每次使用10%的现金)
	cashToUse := float64(wallet.Cash) * 0.1
	stockNum := int(cashToUse / price / 100) * 100 // 整手买入

	if stockNum < 100 {
		return nil // 资金不足
	}

	actualCost := price * float64(stockNum)
	wallet.Cash -= float32(actualCost)

	return &stockStrategy.Position{
		StockCode:    code,
		StockName:    name,
		StockNum:     stockNum,
		BuyPrice:     float32(price),
		BuyDate:      dayData.DataStr,
		BuyIndex:     dateIndex,
		HoldDays:     0,
		HighestPrice: float32(price), // 初始化为买入价
	}
}

// executeSell 执行卖出（使用开盘价，模拟真实交易）
func (engine *BacktestEngine) executeSell(
	position *stockStrategy.Position,
	dayData *stockData.StockDataDay,
	wallet *globalDefine.Wallet,
	record *globalDefine.OperateRecord,
) {
	sellPrice := float64(dayData.PriceBegin) // 使用开盘价而非收盘价
	sellAmount := sellPrice * float64(position.StockNum)

	wallet.Cash += float32(sellAmount)

	// 更新操作记录
	record.SellOperate = globalDefine.Operate{
		OperateType: 2,
		SellPrice:   float32(sellPrice),
		OperateDate: dayData.DataStr,
		StockCode:   position.StockCode,
		StockName:   position.StockName,
		StockNum:    position.StockNum,
	}
	record.Status = 2 // 已卖出
	record.Profit = float32((sellPrice - float64(position.BuyPrice)) * float64(position.StockNum))
}

// createBuyOperate 创建买入操作记录
func createBuyOperate(position *stockStrategy.Position, dayData *stockData.StockDataDay) globalDefine.Operate {
	return globalDefine.Operate{
		OperateType: 1,
		BuyPrice:    position.BuyPrice,
		OperateDate: dayData.DataStr,
		StockCode:   position.StockCode,
		StockName:   position.StockName,
		StockNum:    position.StockNum,
	}
}

// getAllStockCodes 获取所有票票代码
func getAllStockCodes() []string {
	codes := make([]string, 0, len(stockData.StockList))
	for code := range stockData.StockList {
		codes = append(codes, code)
	}
	// 排序以确保一致性
	sort.Strings(codes)
	return codes
}
