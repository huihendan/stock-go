package globaldefine

type Wallet struct {
	Cash      float64
	Positions []Position
}

type Position struct {
	StockCode string
	StockName string
	StockNum  int
	Profit    float64
	BuyPrice  float64
	SellPrice float64
	BuyDate   string
}

type Operate struct {
	OperateType int
	BuyPrice    float64
	SellPrice   float64
	OperateDate string
	OperateTime string
	StockCode   string
	StockName   string
	StockNum    int
}

type OperateRecord struct {
	StockCode   string
	StockName   string
	StockNum    float64
	BuyOperate  Operate
	SellOperate Operate
	Status      int
	Profit      float64
	OperateDate string
	OperateTime string
}
