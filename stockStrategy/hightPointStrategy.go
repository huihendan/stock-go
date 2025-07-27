package stockStrategy

import (
	"stock/globalConfig"
	"stock/logger"
	"stock/painter"
	"stock/stockData"
)

func HighPointStrategy(stockCode string) (isHighPoint bool, dataStr string) {
	if stockCode == "" {
		logger.Warnf("股票代码为空")
		return false, ""
	}

	stock := stockData.GetstockBycode(stockCode)
	if stock == nil {
		logger.Warnf("无法获取股票 %s 数据", stockCode)
		return false, ""
	}
	
	stockSessionLen := len(stock.Datas.DayDatas)

	if stockSessionLen < globalConfig.STOCK_SESSION_LEN {
		logger.Infof("stock %s session len is %d < %d", stockCode, stockSessionLen, globalConfig.STOCK_SESSION_LEN)
		return false, ""
	}

	for _, highPoint := range stock.Datas.HighPoints {
		if highPoint == nil {
			logger.Warnf("股票 %s 存在空的高点数据", stockCode)
			continue
		}

		// indexDesc 为当前点距离当前时间的距离
		indexDesc := stockSessionLen - highPoint.Index

		if indexDesc > globalConfig.STOCK_SESSION_LEN {
			continue
		}

		// 如果最大值距今超过 15天，则不满足条件
		if indexDesc > globalConfig.STOCK_SESSION_HIGHTPOINT_LEN {
			break
		}

		if indexDesc < globalConfig.STOCK_SESSION_HIGHTPOINT_LEN {
			sessionStartIndex := stockSessionLen - globalConfig.STOCK_SESSION_LEN
			if sessionStartIndex < 0 || sessionStartIndex >= len(stock.Datas.DayDatas) {
				logger.Warnf("股票 %s 数组索引越界", stockCode)
				continue
			}
			
			sessionStartData := stock.Datas.DayDatas[sessionStartIndex]
			if sessionStartData == nil {
				logger.Warnf("股票 %s 起始数据为空", stockCode)
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

	if stockSessionLen < globalConfig.STOCK_SESSION_LEN {
		logger.Infof("stock %s session len is %d < %d", stockCode, stockSessionLen, globalConfig.STOCK_SESSION_LEN)
		return
	}

	// 使用单调递减队列维护滑动窗口的最大值
	// queue 存储的是索引，对应的价格是单调递减的
	queue := make([]int, 0)

	// 先处理前 STOCK_SESSION_LEN 天的数据，建立初始队列
	for i := 0; i < globalConfig.STOCK_SESSION_LEN; i++ {
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
			logger.Infof("HighPointStrategyDayByDay stockCode:%s, DateStr:%s, PriceA:%.2f, is max in session",
				stockCode, currentDay.DataStr, currentDay.PriceA)

			painter.PaintStockKline(stockCode)
			break
		}
	}
}
