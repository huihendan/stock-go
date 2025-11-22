package stockStrategy

import (
	"stock-go/stockData"
)

// ===== 选股器接口 =====
// StockSelector 负责从全市场筛选出符合条件的股票代码列表
type StockSelector interface {
	// SelectStocks 选择股票代码列表
	// 参数: allCodes - 全市场股票代码
	// 返回: 筛选后的代码列表
	SelectStocks(allCodes []string) []string

	// GetName 获取选股器名称
	GetName() string
}

// ===== 交易信号生成器接口 =====
// SignalGenerator 负责对单只股票的历史数据，逐日判断买入/卖出信号
type SignalGenerator interface {
	// Reset 初始化策略状态(每只股票开始回测前调用)
	// 用于清空历史数据，确保不同股票之间的状态隔离
	Reset()

	// ProcessDay 处理单日数据,返回交易信号
	// 参数:
	//   - dayData: 当天的K线数据
	//   - dateIndex: 数据索引(从0开始)
	//   - position: 当前持仓状态(nil表示空仓)
	// 返回:
	//   - signal: 1=买入, -1=卖出, 0=无操作
	ProcessDay(dayData *stockData.StockDataDay, dateIndex int, position *Position) int

	// GetName 获取信号生成器名称
	GetName() string
}

// ===== 完整策略接口 =====
// Strategy 完整的交易策略，由选股器和信号生成器组成
type Strategy interface {
	// GetSelector 获取选股器
	GetSelector() StockSelector

	// GetSignalGenerator 获取信号生成器
	GetSignalGenerator() SignalGenerator

	// GetName 获取策略名称
	GetName() string
}

// ===== 持仓状态 =====
// Position 表示当前的持仓状态
type Position struct {
	StockCode string  // 股票代码
	StockName string  // 股票名称
	StockNum  int     // 持有股数
	BuyPrice  float32 // 买入价格
	BuyDate   string  // 买入日期
	BuyIndex  int     // 买入时的数据索引
	HoldDays  int     // 持有天数
}
