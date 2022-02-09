package main

import (
	"fmt"
	hessian "github.com/apache/dubbo-go-hessian2"
	"os"
	"os/signal"
	"stock/pkg"
	"stock/stockData"
	"syscall"
	"time"
)

import (
	_ "github.com/apache/dubbo-go/cluster/cluster_impl"
	_ "github.com/apache/dubbo-go/cluster/loadbalance"
	"github.com/apache/dubbo-go/common/logger"
	_ "github.com/apache/dubbo-go/common/proxy/proxy_factory"
	"github.com/apache/dubbo-go/config"
	_ "github.com/apache/dubbo-go/filter/filter_impl"
	_ "github.com/apache/dubbo-go/protocol/dubbo"
	_ "github.com/apache/dubbo-go/registry/nacos"
	_ "github.com/apache/dubbo-go/registry/protocol"
)

var (
	survivalTimeout = int(3e9)
)

// need to setup environment variable "CONF_PROVIDER_FILE_PATH" to "conf/server.yml" before run
func main() {

	hessian.RegisterPOJO(&pkg.User{})
	config.SetProviderService(new(pkg.UserProvider))
	config.Load()
	stockData.Start()
	//painter.KlineExample()
	initSignal()
}

func initSignal() {
	signals := make(chan os.Signal, 1)
	// It is not possible to block SIGKILL or syscall.SIGSTOP
	signal.Notify(signals, os.Interrupt, os.Kill, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		sig := <-signals
		logger.Infof("get signal %s", sig.String())
		switch sig {
		case syscall.SIGHUP:
			// reload()
		default:
			time.AfterFunc(time.Duration(survivalTimeout), func() {
				logger.Warnf("app exit now by force...")
				os.Exit(1)
			})

			// The program exits normally or timeout forcibly exits.
			fmt.Println("provider app exit now...")
			return
		}
	}
}
