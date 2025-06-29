package stockData

import (
	"sort"
	"stock/globalConfig"
	"stock/logger"
)

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

// 列举峰谷点
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

		if dayDataYes.PointType == POINT_PEAK {
			stock.Datas.HighPoints = append(stock.Datas.HighPoints, dayDataYes)
		} else {
			stock.Datas.LowPoints = append(stock.Datas.LowPoints, dayDataYes)
		}
		//TODO 过滤峰谷点波动
		//stock.DealPointsLen(2)
	}

	// 对HighPoints按照价格从高到低排序
	sort.Slice(stock.Datas.HighPoints, func(i, j int) bool {
		return stock.Datas.HighPoints[i].PriceA > stock.Datas.HighPoints[j].PriceA
	})

	// 对LowPoints按照价格从低到高排序
	sort.Slice(stock.Datas.LowPoints, func(i, j int) bool {
		return stock.Datas.LowPoints[i].PriceA < stock.Datas.LowPoints[j].PriceA
	})

}

// 从峰谷点中获取Section
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

func (stock *StockInfo) DealSessionHighPointDayByDay2() (highPoint bool) {

	highPoint = false
	stockSessionLen := len(stock.Datas.DayDatas)
	if stockSessionLen < globalConfig.STOCK_SESSION_LEN {
		logger.Infof("stock %s session len is %d < %d", stock.Code, stockSessionLen, globalConfig.STOCK_SESSION_LEN)
		return
	}

	// 使用单调递减队列维护滑动窗口的最大值
	// queue 存储的是索引，对应的价格是单调递减的
	queue := make([]int, 0)

	// 先处理前 STOCK_SESSION_LEN 天的数据，建立初始队列
	for i := 0; i < stockSessionLen; i++ {
		// 移除队列中所有小于当前价格的价格对应的索引
		for len(queue) > 0 && stock.Datas.DayDatas[queue[len(queue)-1]].PriceA < stock.Datas.DayDatas[i].PriceA {
			queue = queue[:len(queue)-1]
		}
		queue = append(queue, i)
	}

	// 从第 STOCK_SESSION_LEN 天开始遍历
	for i := globalConfig.STOCK_SESSION_LEN; i < stockSessionLen; i++ {
		currentDay := stock.Datas.DayDatas[i]

		// 移除队列中已经超出滑动窗口范围的索引
		for len(queue) > 0 && queue[0] <= i-globalConfig.STOCK_SESSION_LEN {
			queue = queue[1:]
		}

		// 移除队列中所有小于当前价格的价格对应的索引
		for len(queue) > 0 && stock.Datas.DayDatas[queue[len(queue)-1]].PriceA < currentDay.PriceA {
			queue = queue[:len(queue)-1]
		}

		// 将当前索引加入队列
		queue = append(queue, i)

		// 如果队列的第一个元素（最大值）就是当前索引，说明当前天是区间内的最大值
		if len(queue) > 0 && queue[0] == i {
			logger.Infof("HighPointStrategy stockCode:%s, DateStr:%s, PriceA:%.2f, is max in session",
				stock.Code, currentDay.DataStr, currentDay.PriceA)
			stock.Datas.SessionHighPoints = append(stock.Datas.SessionHighPoints, currentDay)
			highPoint = true
			i = i + 30
			if i >= stockSessionLen {
				break
			}
		}
	}
	return
}

func (stock *StockInfo) DealSessionHighPointDayByDay() (highPoint bool) {

	highPoint = false
	stockSessionLen := len(stock.Datas.DayDatas)
	if stockSessionLen < globalConfig.STOCK_SESSION_LEN {
		logger.Infof("stock %s session len is %d < %d", stock.Code, stockSessionLen, globalConfig.STOCK_SESSION_LEN)
		return
	}

	// 从第 STOCK_SESSION_LEN 天开始遍历
	for i := globalConfig.STOCK_SESSION_LEN; i < stockSessionLen; i++ {
		currentDay := stock.Datas.DayDatas[i]

		// 检查当前天是否是前 STOCK_SESSION_LEN 天内的最大值
		isMaxInSession := true
		sessionStartIndex := i - globalConfig.STOCK_SESSION_LEN

		// 遍历前 STOCK_SESSION_LEN 天的数据，与当前天进行比较
		for j := sessionStartIndex; j < i; j++ {
			if stock.Datas.DayDatas[j].PriceA > currentDay.PriceA {
				isMaxInSession = false
				break
			}
		}

		// 如果当前天是区间内的最大值，则记录并继续
		if isMaxInSession {
			logger.Infof("HighPointStrategy stockCode:%s, DateStr:%s, PriceA:%.2f, is max in session",
				stock.Code, currentDay.DataStr, currentDay.PriceA)
			stock.Datas.SessionHighPoints = append(stock.Datas.SessionHighPoints, currentDay)
			highPoint = true
			i = i + 30
			if i >= stockSessionLen {
				break
			}
		}
	}
	return
}
