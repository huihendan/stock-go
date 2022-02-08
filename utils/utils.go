package utils

import (
	"github.com/apache/dubbo-go/common/logger"
	"time"
)

func CostTime(start time.Time) {
	elapsed := time.Since(start)
	//fmt.Println("cost:%v",elapsed)
	logger.Infof("cost:%v", elapsed)
}
