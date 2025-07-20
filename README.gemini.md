# Stock-Go: 股票数据分析与可视化平台

[![Go Version](https://img.shields.io/badge/Go-1.18+-blue.svg)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

`Stock-Go` 是一个使用 Golang 构建的股票数据分析和可视化工具。它旨在提供一个完整的解决方案，从数据获取、处理、策略分析到结果可视化和通知。

## 核心功能

- **HTTP 服务**: 内置一个轻量级的 Web 服务器，通过 API 接口提供股票数据查询服务。
- **数据可视化**: 能够生成股票的 K-线图、走势图等，便于直观分析。
- **策略分析**: 包含可扩展的股票分析策略模块，例如已实现的“高点策略”。
- **自动化任务**: 提供独立的每日执行程序，用于自动化更新股票数据和执行每日分析。
- **Python 数据脚本**: 使用 Python 脚本从网络获取原始股票数据，并进行预处理。
- **微信通知**: 集成了微信消息推送功能，可以将分析结果或警报发送给用户。

## 项目结构

```
/
├── Data/                  # Python 数据处理脚本
│   ├── getAllStockDatas.py  # 获取全量股票数据
│   └── updateDayDatas.py    # 更新每日股票数据
├── exec/                  # 独立的、可每日执行的 Go 程序
│   ├── analyseDataEveryDay/ # 每日分析任务
│   └── updateDataEveryDay/  # 每日数据更新任务
├── globalConfig/          # 全局配置文件
├── http/                  # HTTP 服务器相关代码
│   ├── server.go            # 服务器启动与路由
│   └── stockHandler.go      # 股票相关的 HTTP Handler
├── logger/                # 日志模块
├── painter/               # 图表绘制与可视化模块
├── stockData/             # 股票数据定义与处理核心逻辑
│   ├── defindStock.go       # 股票数据结构定义
│   └── loadData.go          # 数据加载与解析
├── stockStrategy/         # 股票分析策略
│   └── hightPointStrategy.go # 高点策略实现
├── utils/                 # 通用工具类
│   └── wechat.go            # 微信消息推送工具
├── stockServer.go         # Web 服务器主入口
├── go.mod                 # Go 模块依赖
└── vendor/                # 项目依赖库
```

## 安装与配置

### 1. 环境要求

- **Go**: 1.18 或更高版本
- **Python**: 3.x
- **Make** (可选, 用于简化命令)

### 2. 安装步骤

**a. 克隆项目**
```bash
git clone <your-repo-url>
cd stock-go
```

**b. 安装 Go 依赖**

项目使用了 `go mod` 和 `vendor`。执行以下命令来同步依赖：
```bash
go mod tidy
go mod vendor
```

**c. 安装 Python 依赖**

数据获取脚本依赖于一些 Python 库。请检查 `Data/` 目录下的 `.py` 文件中的 `import` 语句，并使用 `pip` 安装所需的库。常见的库可能包括：
```bash
pip install requests pandas
```

### 3. 数据准备

Go 程序依赖于由 Python 脚本生成的 CSV 数据文件。在运行 Go 应用之前，必须先执行这些脚本。

**a. 获取所有股票列表和基础数据**
```bash
python Data/getAllStockDatas.py
```

**b. 更新最新的日线数据**
```bash
python Data/updateDayDatas.py
```
建议将 `updateDayDatas.py` 设置为每日定时任务。

### 4. 配置

修改 `globalConfig/globalConfig.go` 文件，根据需要配置数据库连接、文件路径、微信推送的 `token` 等信息。

## 如何运行

### 1. 启动 Web 服务器

Web 服务器提供了数据查询的 API 接口。
```bash
go run stockServer.go
```
启动后，您可以访问 `http://localhost:8080` (或您配置的端口) 来与服务交互。

### 2. 执行每日自动化任务

项目提供了两个独立的程序用于每日的自动化处理。

**a. 每日更新数据**

此程序会调用 `stockData` 包的功能，整理和更新本地数据。
```bash
go run exec/updateDataEveryDay/updateDataEveryDay.go
```

**b. 每日执行策略分析**

此程序会加载最新数据，并运行 `stockStrategy` 中定义的策略。
```bash
go run exec/analyseDataEveryDay/analyseDataEveryDay.go
```
建议使用 `crontab` 或其他调度工具来每日自动执行这两个命令。

## 主要模块详解

### `stockData` 模块

这是项目的核心数据层。它负责定义股票的数据结构（如 `Stock`、`DayLine`），并处理数据的加载（从 CSV）、解析和存储。

### `painter` 模块

该模块是项目的可视化核心。它利用 `gonum/plot`、`go-echarts` 等库将股票数据渲染成图表（如 PNG 图片）。`painter/line.go` 是一个很好的例子，展示了如何生成一只股票的走势图。

### `stockStrategy` 模块

这是策略分析引擎。您可以在此目录下添加新的策略文件。每个策略都应该实现一个统一的接口（例如，一个接收 `Stock` 对象并返回分析结果的函数）。`hightPointStrategy` 提供了一个寻找阶段性高点的示例。

### `http` 模块

基于 Go 内置的 `net/http` 包构建。`server.go` 中定义了路由规则，将不同的 URL 路径映射到 `stockHandler.go` 中的处理函数。这些处理函数负责调用底层模块（如 `stockData` 和 `painter`）并返回数据或图表。

### `utils/wechat.go`

一个实用的工具，用于将文本消息推送到企业微信或个人微信。这对于发送交易提醒、策略分析结果等非常有用。

## 主要依赖库

- [gonum/plot](https://gonum.org/v1/plot): 用于生成高质量图表的科学计算库。
- [go-echarts/go-echarts](https://github.com/go-echarts/go-echarts): 一个功能强大的图表库，用于生成交互式图表。
- [Arafatk/glot](https://github.com/Arafatk/glot): Gnuplot 的一个 Go 语言封装。
- [wcharczuk/go-chart](https://github.com/wcharczuk/go-chart): 基础图表库。

## 贡献

欢迎提交 Pull Request 或 Issue 来改进此项目。

## 许可证

本项目采用 [MIT License](https://opensource.org/licenses/MIT) 开源。
