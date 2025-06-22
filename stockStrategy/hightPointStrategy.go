package stockStrategy

import (
	"stock/globalConfig"
	"stock/logger"
	"stock/painter"
	"stock/stockData"
)

func HighPointStrategy(stockCode string) {

	stock := stockData.GetstockBycode(stockCode)
	stockSessionLen := len(stock.Datas.DayDatas)

	if stockSessionLen < globalConfig.STOCK_SESSION_LEN {

		logger.Infof("stock %s session len is %d < %d", stockCode, stockSessionLen, globalConfig.STOCK_SESSION_LEN)
		return
	}

	for _, highPoint := range stock.Datas.HighPoints {

		indexDesc := stockSessionLen - highPoint.Index

		if indexDesc > globalConfig.STOCK_SESSION_LEN {
			continue
		}

		if indexDesc > globalConfig.STOCK_SESSION_HIGHTPOINT_LEN {
			break
		}

		if indexDesc < globalConfig.STOCK_SESSION_HIGHTPOINT_LEN {

			if highPoint.PriceA < stock.Datas.DayDatas[stockSessionLen-globalConfig.STOCK_SESSION_LEN].PriceA {
				continue
			}

			beginDataStr := stock.Datas.DayDatas[stockSessionLen-globalConfig.STOCK_SESSION_LEN].DataStr
			logger.Infof("HighPointStrategy stockCode:%s indexDesc:%d, DateStr:%s, sessionBeginDate:%s", stockCode, indexDesc, highPoint.DataStr, beginDataStr)

			painter.PaintStockKline(stockCode)
			break
		}

	}
}
