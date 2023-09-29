package http

import (
	"log"
	"net/http"
)

func init() {
	http.HandleFunc("/stock", stock)
	http.HandleFunc("/readDayDate", readDayDate)

}

func StartServer() {
	// 启动HTTP服务器，监听本地8080端口
	log.Fatal(http.ListenAndServe(":8080", nil))
}
