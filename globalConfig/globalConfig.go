package globalConfig

import "runtime"

const (
	STOCK_SESSION_LEN            = 500
	STOCK_SESSION_HIGHTPOINT_LEN = 15
)

const (
	STOCK_DATA_LOAD_PCT = 500
	STOCK_DATA_LOAD_MOD = 0
)

var DATA_PATH = "../Data/"
var LOG_PATH = "../Log/"

func init() {

	sysType := runtime.GOOS
	if sysType == "linux" {
		DATA_PATH = "../Data/"
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
