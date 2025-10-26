package globaldefine

type Wallet struct {
	Cash      float32
	Positions []Position
}

type Position struct {
	StockCode string
	StockName string
	StockNum  int
	Profit    float32
	BuyPrice  float32
	SellPrice float32
	BuyDate   string
	Status    int // 1-持仓 2-空仓
	BuyIndex  int
	SellIndex int
}

type Operate struct {
	OperateType int
	BuyPrice    float32
	SellPrice   float32
	OperateDate string
	OperateTime string
	StockCode   string
	StockName   string
	StockNum    int
}

type OperateRecord struct {
	StockCode   string
	StockName   string
	StockNum    int
	BuyOperate  Operate
	SellOperate Operate
	Status      int
	Profit      float32
	OperateDate string
	OperateTime string
}
