package utils

import (
	"log/slog"
	"time"
)

func CostTime(start time.Time) {
	elapsed := time.Since(start)
	//fmt.Println("cost:%v",elapsed)
	slog.Info("cost time", "elapsed", elapsed)
}
