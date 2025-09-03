package plot

import (
	"stock/globalConfig"
	"stock/logger"
	"stock/stockData"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

func PlotPoints(stockCode string) {
	stock := stockData.GetstockBycode(stockCode)
	if stock == nil {
		logger.Error("PaintStockKline failed, stock data not exist", "code", stockCode)
		return
	}

	// 创建 plotter.XYs 类型的数据
	points := make(plotter.XYs, len(stock.Datas.DayDatas))
	for index, data := range stock.Datas.DayDatas {
		points[index].X = float64(index)
		points[index].Y = float64(data.PriceA)
	}

	p := plot.New()
	p.Title.Text = "stock价格走势图 - " + stockCode
	p.X.Label.Text = "时间索引"
	p.Y.Label.Text = "价格"

	err := plotutil.AddLinePoints(p, "价格走势", points)
	if err != nil {
		panic(err)
	}

	// 计算合适的图片宽度，确保每个数据点都有足够的显示空间
	width := vg.Length(len(points)) * 0.1 * vg.Inch
	if width < 4*vg.Inch {
		width = 4 * vg.Inch
	}
	// if width > 20*vg.Inch {
	// 	width = 20 * vg.Inch
	// }

	p.Save(width, 4*vg.Inch, globalConfig.LOG_PATH+"/plot/"+stockCode+".png")
}
