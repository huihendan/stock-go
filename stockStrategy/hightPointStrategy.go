package stockStrategy

import (
	globalDefine "stock-go/globalDefine"
	"stock-go/logger"
	"stock-go/painter"
	"stock-go/stockData"
)

func HighPointStrategy(stockCode string) (isHighPoint bool, dataStr string) {
	if stockCode == "" {
		logger.Warnf("stock代码为空")
		return false, ""
	}

	stock := stockData.GetstockBycode(stockCode)
	if stock == nil {
		logger.Warnf("无法获取stock %s 数据", stockCode)
		return false, ""
	}

	stockSessionLen := len(stock.Datas.DayDatas)

	if stockSessionLen < globalDefine.STOCK_SESSION_LEN {
		logger.Infof("stock %s session len is %d < %d", stockCode, stockSessionLen, globalDefine.STOCK_SESSION_LEN)
		return false, ""
	}

	for _, highPoint := range stock.Datas.HighPoints {
		if highPoint == nil {
			logger.Warnf("stock %s 存在空的高点数据", stockCode)
			continue
		}

		// indexDesc 为当前点距离当前时间的距离
		indexDesc := stockSessionLen - highPoint.Index

		if indexDesc > globalDefine.STOCK_SESSION_LEN {
			continue
		}

		// 如果最大值距今超过 15天，则不满足条件
		if indexDesc > globalDefine.STOCK_SESSION_HIGHTPOINT_LEN {
			break
		}

		if indexDesc < globalDefine.STOCK_SESSION_HIGHTPOINT_LEN {
			sessionStartIndex := stockSessionLen - globalDefine.STOCK_SESSION_LEN
			if sessionStartIndex < 0 || sessionStartIndex >= len(stock.Datas.DayDatas) {
				logger.Warnf("stock %s 数组索引越界", stockCode)
				continue
			}

			sessionStartData := stock.Datas.DayDatas[sessionStartIndex]
			if sessionStartData == nil {
				logger.Warnf("stock %s 起始数据为空", stockCode)
				continue
			}

			if highPoint.PriceA < sessionStartData.PriceA {
				continue
			}

			beginDataStr := sessionStartData.DataStr
			logger.Infof("HighPointStrategy stockCode:%s indexDesc:%d, DateStr:%s, sessionBeginDate:%s", stockCode, indexDesc, highPoint.DataStr, beginDataStr)

			painter.PaintStockKline(stockCode)
			return true, highPoint.DataStr
		}
	}
	return false, ""
}

// HighPointStrategyDayByDay 从 STOCK_SESSION_LEN 开始，对每一天的数据进行判断，检查是否是当前区间的最大值
func HighPointStrategyDayByDay(stockCode string) {
	stock := stockData.GetstockBycode(stockCode)
	stockSessionLen := len(stock.Datas.DayDatas)

	if stockSessionLen < globalDefine.STOCK_SESSION_LEN {
		logger.Infof("stock %s session len is %d < %d", stockCode, stockSessionLen, globalDefine.STOCK_SESSION_LEN)
		return
	}

	// 使用单调递减队列维护滑动窗口的最大值
	// queue 存储的是索引，对应的价格是单调递减的
	queue := make([]int, 0)

	// 先处理前 STOCK_SESSION_LEN 天的数据，建立初始队列
	for i := 0; i < globalDefine.STOCK_SESSION_LEN; i++ {
		// 移除队列中所有小于当前价格的价格对应的索引
		for len(queue) > 0 && stock.Datas.DayDatas[queue[len(queue)-1]].PriceA < stock.Datas.DayDatas[i].PriceA {
			queue = queue[:len(queue)-1]
		}
		queue = append(queue, i)
	}

	// 从第 STOCK_SESSION_LEN 天开始遍历
	for i := globalDefine.STOCK_SESSION_LEN; i < stockSessionLen; i++ {
		currentDay := stock.Datas.DayDatas[i]

		// 移除队列中已经超出滑动窗口范围的索引
		for len(queue) > 0 && queue[0] <= i-globalDefine.STOCK_SESSION_LEN {
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
			logger.Infof("HighPointStrategyDayByDay stockCode:%s, DateStr:%s, PriceA:%.2f, is max in session",
				stockCode, currentDay.DataStr, currentDay.PriceA)

			painter.PaintStockKline(stockCode)
			break
		}
	}
}

// HighPointStrategyLast 判断当前（最后一天）的数值是否是 STOCK_SESSION_LEN 内的最大值
// 并且过去30天内都不满足最大值要求，今天是第一次满足条件
func HighPointStrategyLast(stockCode string) (isHighPoint bool, dataStr string) {
	if stockCode == "" {
		logger.Warnf("stock代码为空")
		return false, ""
	}

	stock := stockData.GetstockBycode(stockCode)
	if stock == nil {
		logger.Warnf("无法获取stock %s 数据", stockCode)
		return false, ""
	}

	stockSessionLen := len(stock.Datas.DayDatas)

	if stockSessionLen < globalDefine.STOCK_SESSION_LEN {
		logger.Infof("stock %s session len is %d < %d", stockCode, stockSessionLen, globalDefine.STOCK_SESSION_LEN)
		return false, ""
	}

	// 检查是否有足够的数据进行30天历史检查
	if stockSessionLen < globalDefine.STOCK_SESSION_LEN+30 {
		logger.Infof("stock %s session len is %d, not enough for 30-day historical check", stockCode, stockSessionLen)
		return false, ""
	}

	// 获取最后一天的数据
	lastDayData := stock.Datas.DayDatas[stockSessionLen-1]
	if lastDayData == nil {
		logger.Warnf("stock %s 最后一天数据为空", stockCode)
		return false, ""
	}

	// 计算当前的会话开始索引
	currentSessionStartIndex := stockSessionLen - globalDefine.STOCK_SESSION_LEN
	if currentSessionStartIndex < 0 {
		currentSessionStartIndex = 0
	}

	// 检查最后一天是否是当前会话区间内的最大值
	maxPrice := lastDayData.PriceA
	for i := currentSessionStartIndex; i < stockSessionLen-1; i++ {
		if stock.Datas.DayDatas[i] != nil && stock.Datas.DayDatas[i].PriceA > maxPrice {
			// 最后一天不是最大值
			return false, ""
		}
	}

	// 检查过去30天内是否有任何一天满足在其对应的STOCK_SESSION_LEN区间内为最大值的条件
	// 从倒数第2天开始检查，向前检查30天
	for dayOffset := 1; dayOffset <= 30; dayOffset++ {
		checkDayIndex := stockSessionLen - 1 - dayOffset
		if checkDayIndex < globalDefine.STOCK_SESSION_LEN-1 {
			break // 没有足够的历史数据
		}

		checkDayData := stock.Datas.DayDatas[checkDayIndex]
		if checkDayData == nil {
			continue
		}

		// 计算这一天对应的会话开始索引
		checkSessionStartIndex := checkDayIndex + 1 - globalDefine.STOCK_SESSION_LEN
		if checkSessionStartIndex < 0 {
			checkSessionStartIndex = 0
		}

		// 检查这一天是否是其对应会话区间内的最大值
		isMaxInSession := true
		checkMaxPrice := checkDayData.PriceA
		for i := checkSessionStartIndex; i <= checkDayIndex; i++ {
			if i != checkDayIndex && stock.Datas.DayDatas[i] != nil && stock.Datas.DayDatas[i].PriceA > checkMaxPrice {
				isMaxInSession = false
				break
			}
		}

		if isMaxInSession {
			// 过去30天内有一天满足了最大值条件，不满足首次满足的要求
			logger.Infof("HighPointStrategyLast stockCode:%s, past day %s already satisfied max condition",
				stockCode, checkDayData.DataStr)
			return false, ""
		}
	}

	// 最后一天是区间内的最大值，且过去30天内都不满足条件，今天是第一次满足
	sessionStartData := stock.Datas.DayDatas[currentSessionStartIndex]
	beginDataStr := ""
	if sessionStartData != nil {
		beginDataStr = sessionStartData.DataStr
	}

	logger.Infof("HighPointStrategyLast stockCode:%s, DateStr:%s, PriceA:%.2f, is first time max in session in 30 days, sessionBeginDate:%s",
		stockCode, lastDayData.DataStr, lastDayData.PriceA, beginDataStr)

	return true, lastDayData.DataStr
}
