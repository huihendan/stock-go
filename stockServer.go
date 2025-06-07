package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"stock/http"
	"stock/painter"
	"stock/stockData"
	"syscall"
	"time"
)

var (
	survivalTimeout = int(3e9)
)

func init() {
	logFile, err := os.OpenFile("stock.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("无法打开日志文件: %v\n", err)
		return
	}

	// 创建一个多输出写入器，同时写入文件和标准输出
	multiWriter := io.MultiWriter(logFile, os.Stdout)

	// 创建一个文本处理器，将日志输出到多输出写入器
	handler := slog.NewTextHandler(multiWriter, nil)

	// 设置默认的日志记录器
	slog.SetDefault(slog.New(handler))
}

// need to setup environment variable "CONF_PROVIDER_FILE_PATH" to "conf/server.yml" before run
func main() {
	go http.StartServer()

	go stockData.Start()

	painter.PaintStockKline("sz.000001")
	initSignal()
}

func initSignal() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	for {
		sig := <-signals
		slog.Info("get signal", "signal", sig.String())
		switch sig {
		case syscall.SIGHUP:
			// reload()
		default:
			time.AfterFunc(time.Duration(survivalTimeout), func() {
				slog.Warn("app exit now by force...")
				os.Exit(1)
			})

			// The program exits normally or timeout forcibly exits.
			fmt.Println("provider app exit now...")
			return
		}
	}
}
