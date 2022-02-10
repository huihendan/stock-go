package painter

import (
	"github.com/apache/dubbo-go/common/logger"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"io"
	"os"
	. "stock/stockData"
)

func LineExample() {
	page := components.NewPage()
	line := charts.NewLine()

	items := make([]opts.LineData, 0)
	pointsH := make([]opts.ScatterData, 0)
	pointsL := make([]opts.ScatterData, 0)
	fruits := make([]string, 0)
	var tiele string
	for _, stock := range Stocks {
		for _, data := range stock.Datas.DayDatas {
			items = append(items, opts.LineData{
				Value:      data.PriceA,
				SymbolSize: 1,
			})
			fruits = append(fruits, data.DataStr)
			if data.PointType == POINT_PEAK {
				pointsH = append(pointsH, opts.ScatterData{
					Value:        data.PriceA,
					Symbol:       "pin",
					SymbolSize:   10,
					SymbolRotate: 10,
				})
			} else {
				pointsH = append(pointsH, opts.ScatterData{})
			}

			if data.PointType == POINT_BOTTOM {
				pointsL = append(pointsL, opts.ScatterData{
					Value:        data.PriceA,
					Symbol:       "pin",
					SymbolSize:   10,
					SymbolRotate: 3,
				})
			} else {
				pointsL = append(pointsL, opts.ScatterData{})
			}

		}
		tiele = stock.Code + "_" + stock.Name
		break
	}
	line.SetXAxis(fruits)
	line.AddSeries("Category A", items)

	scatter := charts.NewScatter()
	scatter.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "basic scatter example"}),
	)
	scatter.SetXAxis(fruits)
	//SeriesOpts
	scatter.AddSeries("Point B", pointsH)
	scatter.AddSeries("Point A", pointsL)
	line.Overlap(scatter)

	//支持X,Y轴缩放
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: tiele,
		}),
		charts.WithColorsOpts(opts.Colors{"#5470c6", "red", "#25E712"}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:       "slider",
			Start:      50,
			End:        100,
			XAxisIndex: []int{0},
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:       "inside",
			Start:      50,
			End:        100,
			XAxisIndex: []int{0},
		}))
	page.AddCharts(line)

	f, err := os.Create("line.html")
	if err != nil {
		panic(err)
	}
	page.Render(io.MultiWriter(f))
}

func KlineExample() {
	page := components.NewPage()
	kline := charts.NewKLine()

	y := make([]opts.KlineData, 0)
	x := make([]string, 0)
	var tiele string
	for _, stock := range Stocks {
		for _, data := range stock.Datas.DayDatas {
			y = append(y, opts.KlineData{Value: [4]float32{data.PriceBegin, data.PriceEnd, data.PriceHigh, data.PriceLow}})
			x = append(x, data.DataStr)
		}
		tiele = stock.Code + "_" + stock.Name
		break
	}
	kline.SetXAxis(x)
	kline.AddSeries("Category A", y)

	//支持X,Y轴缩放
	kline.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: tiele,
		}),
		charts.WithColorsOpts(opts.Colors{"#5470c6", "red", "#25E712"}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:       "slider",
			Start:      50,
			End:        100,
			XAxisIndex: []int{0},
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:       "inside",
			Start:      50,
			End:        100,
			XAxisIndex: []int{0},
		}))
	page.AddCharts(kline)

	f, err := os.Create("kline.html")
	if err != nil {
		panic(err)
	}
	page.Render(io.MultiWriter(f))
}

func PaintStock(stock *StockInfo) {
	page := components.NewPage()
	line := charts.NewLine()

	items := make([]opts.LineData, 0)
	fruits := make([]string, 0)
	for _, data := range stock.Datas.DayDatas {
		items = append(items, opts.LineData{Value: data.PriceA})
		fruits = append(fruits, data.DataStr)
	}
	tiele := stock.Code + "_" + stock.Name

	line.SetXAxis(fruits)
	line.AddSeries("Category A", items)

	//支持X,Y轴缩放
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: tiele,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:       "slider",
			Start:      50,
			End:        100,
			XAxisIndex: []int{0},
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:       "inside",
			Start:      50,
			End:        100,
			XAxisIndex: []int{0},
		}))
	page.AddCharts(line)
	f, err := os.Create(stock.Code + ".html")
	if err != nil {
		panic(err)
	}
	page.Render(io.MultiWriter(f))
}

func PaintStockKline(stock *StockInfo) {
	page := components.NewPage()
	kline := charts.NewKLine()
	y := make([]opts.KlineData, 0)
	x := make([]string, 0)
	var tiele string
	for _, data := range stock.Datas.DayDatas {
		x = append(x, data.DataStr)
		y = append(y, opts.KlineData{Value: [4]float32{data.PriceBegin, data.PriceEnd, data.PriceHigh, data.PriceLow}})
	}
	kline.SetXAxis(x)
	kline.AddSeries("Category A", y)

	//支持X,Y轴缩放
	kline.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: tiele,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:       "slider",
			Start:      50,
			End:        100,
			XAxisIndex: []int{0},
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:       "inside",
			Start:      50,
			End:        100,
			XAxisIndex: []int{0},
		}))

	sectionH := make([]opts.ScatterData, 0)
	sectionL := make([]opts.ScatterData, 0)
	j := 0
	for i := 0; i < len(stock.Datas.Sections); i++ {
		stockDay := stock.Datas.Sections[i]
		for ; j < stockDay.Index-1; j++ {
			sectionH = append(sectionH, opts.ScatterData{})
			sectionL = append(sectionL, opts.ScatterData{})
		}
		if stockDay.PointType == POINT_PEAK {
			sectionH = append(sectionH, opts.ScatterData{
				Value:        stockDay.PriceA,
				Symbol:       "pin",
				SymbolSize:   30,
				SymbolRotate: 10,
			})
			sectionL = append(sectionL, opts.ScatterData{})
		} else {
			sectionL = append(sectionL, opts.ScatterData{
				Value:        stockDay.PriceA,
				Symbol:       "pin",
				SymbolSize:   30,
				SymbolRotate: 10,
			})
			sectionH = append(sectionH, opts.ScatterData{})
		}
		j++
	}

	scatter := charts.NewScatter()
	scatter.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "basic scatter example"}),
	)
	scatter.SetXAxis(x)
	//SeriesOpts
	scatter.AddSeries("Point B", sectionH)
	scatter.AddSeries("Point A", sectionL)
	kline.Overlap(scatter)

	page.AddCharts(kline)

	f, err := os.Create(stock.Code + ".html")
	if err != nil {
		panic(err)
	}
	page.Render(io.MultiWriter(f))
	logger.Infof("finish")
}
