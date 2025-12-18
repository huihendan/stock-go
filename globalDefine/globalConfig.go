package globaldefine

import "runtime"

const (
	STOCK_SESSION_LEN            = 500
	STOCK_SESSION_HIGHTPOINT_LEN = 15
)

const (
	STOCK_DATA_LOAD_PCT = 50
	STOCK_DATA_LOAD_MOD = 1
)

var DATA_PATH = "../Data/"
var LOG_PATH = "../Log/"

var ExecuteUpdataDataTime = "19:00"
var ExecuteAnalyseDataTime = "19:30"

func init() {

	sysType := runtime.GOOS
	if sysType == "linux" {
		DATA_PATH = "/home/beven/gits/stockData/"
		LOG_PATH = "../Log/"
	}

	if sysType == "darwin" {
		DATA_PATH = "/Users/beven/Item/stockData/"
		LOG_PATH = "/Users/beven/Item/stock-go/Log/"
	}

	if sysType == "windows" {
		DATA_PATH = "D:\\Data\\"
		LOG_PATH = "D:\\Log\\"
	}
}
