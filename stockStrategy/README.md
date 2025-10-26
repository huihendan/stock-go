# 策略模块

## 文件结构

```
stockStrategy/
├── strategy.go                      # 旧策略实现（向后兼容，已重构Mode 1）
├── types.go                         # 新增：策略接口和基础类型定义
├── buyHighSellLowStrategy.go       # 新增：追涨杀跌策略1（新实现）
├── buyHighSellLowStrategy_test.go  # 新增：策略1测试文件
├── hightPointStrategy.go            # 高点策略实现
├── hightPointStrategy_test.go       # 高点策略测试
└── README.md                        # 本文件
```

## 新旧代码关系

### 旧代码（strategy.go）
- 保留用于**向后兼容**
- `Strategy_Mode_1` 已自动重定向到新实现 `BuyHighSellLowStrategy`
- `Strategy_Mode_2-3` 保留旧实现（⚠️ 存在未来函数问题）
- `Strategy_Mode_4-6` 待实现
- 辅助函数（`calculateMA`、`calculateRSI` 等）保留供各策略共用

### 新代码
- **types.go**: 定义策略接口 `StockStrategy` 和基础类型
- **buyHighSellLowStrategy.go**: 策略1的新实现
  - ✅ 无未来函数问题
  - ✅ 实现 `StockStrategy` 接口
  - ✅ 参数可配置
  - ✅ 模拟实时交易

### 迁移说明

如果你的代码使用了旧的 `DealStrategys` 函数：

```go
// 旧用法（仍然有效）
operates := stockStrategy.DealStrategys("000001", stockStrategy.Strategy_Mode_1)
```

现在会自动调用新实现，无需修改代码。

如果你想直接使用新接口：

```go
// 新用法（推荐）
strategy := stockStrategy.NewBuyHighSellLowStrategy()
operates := strategy.DealStrategy("000001")
```

## 新增文件说明

### types.go
定义策略模块的核心类型和接口：
- `Wallet`: 钱包状态结构，用于跟踪持仓信息
- `StockStrategy`: 策略接口，定义了所有策略必须实现的方法

### buyHighSellLowStrategy.go
实现追涨杀跌策略1：
- `BuyHighSellLowStrategy`: 策略实现结构
- `NewBuyHighSellLowStrategy()`: 创建默认配置的策略实例
- `NewBuyHighSellLowStrategyWithConfig()`: 创建自定义配置的策略实例
- `DealStrategy()`: 执行策略的主方法
- `DealStrategyBuy()`: 判断买入条件
- `DealStrategySell()`: 判断卖出条件

### buyHighSellLowStrategy_test.go
策略测试文件：
- `TestBuyHighSellLowStrategy`: 基本功能测试
- `TestBuyHighSellLowStrategyCustomConfig`: 自定义配置测试
- `BenchmarkBuyHighSellLowStrategy`: 性能测试

## 使用示例

```go
package main

import (
    "fmt"
    "stock/stockStrategy"
    "stock/stockData"
)

func main() {
    // 加载票票数据
    stockData.LoadPreStockList()
    stockData.LoadDataOneByOne()
    
    // 创建策略实例
    strategy := stockStrategy.NewBuyHighSellLowStrategy()
    
    // 执行策略
    operates := strategy.DealStrategy("000001")
    
    // 处理交易记录
    for key, record := range operates {
        if record.Status == 2 { // 已完成交易
            fmt.Printf("交易: %s\n", key)
            fmt.Printf("  买入: %s %.2f\n", 
                record.BuyOperate.OperateDate, 
                record.BuyOperate.BuyPrice)
            fmt.Printf("  卖出: %s %.2f\n", 
                record.SellOperate.OperateDate, 
                record.SellOperate.SellPrice)
            fmt.Printf("  收益: %.2f\n", record.Profit)
        }
    }
}
```

## 运行测试

```bash
# 运行策略测试
go test -v -run TestBuyHighSellLowStrategy

# 运行自定义配置测试
go test -v -run TestBuyHighSellLowStrategyCustomConfig

# 运行性能测试
go test -bench=BenchmarkBuyHighSellLowStrategy
```

## 设计特点

1. **防止未来函数**: 策略方法只接收当前价格参数，通过内部维护的历史价格队列进行判断
2. **模拟实时交易**: 按时间顺序遍历历史数据，确保回测结果可靠
3. **参数可配置**: 支持自定义回看天数、止损比例、持有时间等参数
4. **接口统一**: 实现 `StockStrategy` 接口，便于扩展新策略

## 详细设计文档

参见：`ReadMe/策略模块设计.md`

