package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	globalDefine "stock/globalDefine"
	"stock/logger"
	"stock/utils"

	"syscall"
	"time"
)

func main() {

	executeTime := globalDefine.ExecuteUpdataDataTime
	utils.DoWorkEveryDayOnce(doworkEveryDay, &executeTime)

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Infof("收到中断信号，程序退出")
}

func doworkEveryDay() {
	// 更新数据
	updateStockData()
}

func updateStockData() error {

	logger.Infof("开始执行数据更新任务 - %s", time.Now().Format("2006-01-02 15:04:05"))
	defer logger.Infof("数据更新任务完成 - %s", time.Now().Format("2006-01-02 15:04:05"))

	// 检查 DATA_PATH 是否存在
	if _, err := os.Stat(globalDefine.DATA_PATH); os.IsNotExist(err) {
		logger.Errorf("数据目录不存在: %s", globalDefine.DATA_PATH)
		return fmt.Errorf("数据目录不存在: %s", globalDefine.DATA_PATH)
	}

	// 构建 Python 脚本的完整路径
	scriptPath := filepath.Join(globalDefine.DATA_PATH, "updateDayDatas.py")

	// 检查脚本文件是否存在
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		logger.Errorf("Python 脚本不存在: %s", scriptPath)
		return fmt.Errorf("Python 脚本不存在: %s", scriptPath)
	}

	// 检查 Python 解释器是否可用
	pythonPath, err := exec.LookPath("python3")
	if err != nil {
		logger.Errorf("未找到 python3 解释器: %v", err)
		return fmt.Errorf("未找到 python3 解释器: %v", err)
	}

	cmd := exec.Command(pythonPath, scriptPath)
	cmd.Dir = globalDefine.DATA_PATH // 设置工作目录

	// 捕获输出
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Errorf("执行 Python 脚本失败: %v", err)
		logger.Errorf("脚本输出: %s", string(output))
		return fmt.Errorf("执行 Python 脚本失败: %v, 输出: %s", err, string(output))
	}

	if len(output) > 0 {
		logger.Infof("脚本输出:\n%s", string(output))
	}

	return nil
}
