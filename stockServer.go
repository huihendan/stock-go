package main

// 使用特殊注释防止导入被重新排序
import (
	_ "stock/logger" // 确保logger最先初始化

	"fmt"
	"os"
	"os/signal"
	"stock/http"
	"stock/logger"
	"stock/painter"
	"stock/stockData"
	"syscall"
	"time"
)

var (
	survivalTimeout = int(3e9)
)

// need to setup environment variable "CONF_PROVIDER_FILE_PATH" to "conf/server.yml" before run
func main() {
	// 确保在程序退出时关闭日志处理器
	defer logger.Close()

	go http.StartServer()

	go stockData.Start()

	painter.PaintStockKline("sz.000001")
	initSignal()
}

func initSignal() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	for {
		sig := <-signals
		logger.Infof("get signal %s", sig.String())
		switch sig {
		case syscall.SIGHUP:
			// reload()
		default:
			time.AfterFunc(time.Duration(survivalTimeout), func() {
				logger.Warn("app exit now by force...")
				os.Exit(1)
			})

			// The program exits normally or timeout forcibly exits.
			fmt.Println("provider app exit now...")
			// 确保所有日志都被写入
			logger.Close()
			return
		}
	}
}
