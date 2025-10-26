package stockStrategy

import globalDefine "stock/globalDefine"

// // Wallet 钱包状态
// type Wallet struct {
// 	Status   int     // 1-持仓 2-空仓
// 	BuyPrice float32 // 买入价格
// 	BuyDate  string  // 买入日期
// 	BuyIndex int     // 买入的数据索引
// }

// StockStrategy 股票交易策略接口
type StockStrategy interface {
	//执行策略，通过GetstockBycode内部获取数据
	DealStrategy(code string) (operates map[string]OperateRecord)

	//判断是否符合买入条件
	//只传入当前价格，防止访问未来数据
	DealStrategyBuy(price float32, dateIndex int) (buy bool)

	//判断是否符合卖出条件
	//只传入当前价格，防止访问未来数据
	DealStrategySell(price float32, dataIndex int, wallet *globalDefine.Wallet) (sell bool)
}
