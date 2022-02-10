package stockData

func (stock *StockInfo) DealPointsLen(len int) {
	var stockLast *StockDataDay
	var stockNow *StockDataDay
	//stockLast = stock.Datas.Points[0]
	for index, row := range stock.Datas.Points {
		if index == 0 {
			stockLast = row
			stockNow = row
			continue
		}
		stockLast = stockNow
		stockNow = row

		//小于最小长度，需要被过滤掉
		if stockNow.Index-stockLast.Index < len {

		}
	}
}

//列举峰谷点
func (stock *StockInfo) DealStockPoints() {

	var dayDataToday *StockDataDay
	var dayDataYes *StockDataDay
	for index, dayData := range stock.Datas.DayDatas {
		if index == 0 {
			dayDataYes = dayData
			dayDataToday = dayData
			continue
		}
		dayDataYes = dayDataToday
		dayDataToday = dayData
		if dayDataToday.PriceA > dayDataYes.PriceA {
			dayDataToday.PointType = POINT_PEAK
			dayDataToday.Trend = POINT_PEAK
		} else {
			dayDataToday.PointType = POINT_BOTTOM
			dayDataToday.Trend = POINT_BOTTOM
		}
		if dayDataToday.PointType == dayDataYes.PointType {
			dayDataYes.PointType = POINT_NORMAL
		}

		//统计所有的峰谷点
		if dayDataYes.PointType != POINT_NORMAL {
			stock.Datas.Points = append(stock.Datas.Points, dayDataYes)
		}

		//TODO 过滤峰谷点波动
		//stock.DealPointsLen(2)
	}

}

//从峰谷点中获取Section
func (stock *StockInfo) DealStockSession(index int) {
	len := len(stock.Datas.Points)
	dataDayBegin := stock.Datas.Points[index]
	dataDayH := stock.Datas.Points[index]
	dataDayL := stock.Datas.Points[index]
	for ; index < len; index++ {
		dataDay := stock.Datas.Points[index]
		if dataDay.PriceA > dataDayH.PriceA {
			dataDayH = dataDay
		} else if dataDay.PriceA < dataDayL.PriceA {
			dataDayL = dataDay
		}
		if dataDay.Index-dataDayBegin.Index > 50 {
			break
		}
	}

	if dataDayH.PriceA/dataDayL.PriceA > 1.2 {
		if dataDayH.Index < dataDayL.Index {
			stock.Datas.Sections = append(stock.Datas.Sections, dataDayH)
			stock.Datas.Sections = append(stock.Datas.Sections, dataDayL)
		} else {
			stock.Datas.Sections = append(stock.Datas.Sections, dataDayL)
			stock.Datas.Sections = append(stock.Datas.Sections, dataDayH)
		}
	}

	if index < len {
		stock.DealStockSession(index)
	}
}
