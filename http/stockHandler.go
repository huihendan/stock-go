package http

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"log/slog"
	"net/http"
	"stock/stockData"
)

type stockParam struct {
	Code       string   `json:"code"`
	IsReLoad   bool     `json:"isReLoad"`
	IsReadFast bool     `json:"isReadFast"`
	CmdCount   int      `json:"cmdCount"`
	IsPlot     bool     `json:"IsPlot"`
	Methods    []string `json:"Methods"`
}

func stock(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err1 := ioutil.ReadAll(r.Body)
	if err1 != nil {
		panic(err1)
	}
	slog.Info("received request", "body", string(bodyBytes))
	var param stockParam
	//err := json.NewDecoder(bodyBytes).Decode(&param)
	err := json.Unmarshal([]byte(bodyBytes), &param)

	param.CmdCount = 0
	slog.Info("received param", "param", param)

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
	if param.CmdCount < len(param.Methods) {
		methodCost(param)
	}
}

func doMethod(param *stockParam) {

	method := param.Methods[param.CmdCount]
	switch method {
	case "Sock.ReadDayData":
		stockData.LoadDataByCode(param.Code)
	case "Sock.AnalyzePaintSections":
		stockData.DealAllStocksPoints()
	default:
		log.Printf("method:[%s] is not support!", method)
	}
}
