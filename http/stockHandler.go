package http

import (
	"encoding/json"
	"github.com/apache/dubbo-go/common/logger"
	"io/ioutil"
	"log"
	"net/http"
	"stock/stockData"
)

type stockParam struct {
	Code       string   `json:"code"`
	IsReLoad   bool     `json:"isReLoad"`
	IsReadFast bool     `json:"isReadFast"`
	CmdCount   int      `json:"cmdCount"`
	IsPlot     bool     `json:"IsPlot"`
	methods    []string `json:"methods"`
}

func stock(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err1 := ioutil.ReadAll(r.Body)
	if err1 != nil {
		panic(err1)
	}
	logger.Infof("rec %s", string(bodyBytes))
	var param stockParam
	err := json.NewDecoder(r.Body).Decode(&param)
	param.CmdCount = 0
	logger.Infof("rec param %v", param)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	methodCost(&param)

	log.Printf("rev param %v", param)
}

func methodCost(param *stockParam) {
	doMethod(param)

	param.CmdCount++
	if param.CmdCount < len(param.methods) {
		methodCost(param)
	}
}

func doMethod(param *stockParam) {

	method := param.methods[param.CmdCount]
	switch method {
	case "Sock.ReadDayData":

	case "Sock.AnalyzePaintSections":
		stockData.DealStocksPoints()
	default:
		log.Printf("method:[%s] is not support!", method)
	}
}
