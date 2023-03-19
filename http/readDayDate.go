package http

import (
	"encoding/json"
	"log"
	"net/http"
)

type readParam struct {
	Code       string `json:"code"`
	IsReLoad   bool   `json:"isReLoad"`
	IsReadFast bool   `json:"isReadFast"`
	CmdCount   int    `json:"cmdCount"`
	IsPlot     bool   `json:"IsPlot"`
}

func readDayDate(w http.ResponseWriter, r *http.Request) {
	var param readParam
	err := json.NewDecoder(r.Body).Decode(&param)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("rev param %v", param)
}
