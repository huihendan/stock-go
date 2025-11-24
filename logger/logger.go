package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	globalDefine "stock-go/globalDefine"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	once           sync.Once
	defaultHandler *CustomTextHandler // 导出处理器实例
)

// 自定义文本处理器，直接控制输出格式
type CustomTextHandler struct {
	w        io.Writer
	level    slog.Level
	mu       sync.Mutex
	logChan  chan string
	shutdown chan struct{}
}

func NewCustomTextHandler(w io.Writer, opts *slog.HandlerOptions) *CustomTextHandler {
	level := slog.LevelInfo
	if opts != nil && opts.Level != nil {
		level = opts.Level.Level()
	}

	h := &CustomTextHandler{
		w:        w,
		level:    level,
		logChan:  make(chan string, 10000), // 缓冲大小可以根据需要调整
		shutdown: make(chan struct{}),
	}

	// 启动后台处理协程
	go h.processLogs()

	return h
}

// 后台处理日志的协程
func (h *CustomTextHandler) processLogsSync(logMsg string) {
	h.mu.Lock()
	_, _ = h.w.Write([]byte(logMsg))
	h.mu.Unlock()
}

// 后台处理日志的协程
func (h *CustomTextHandler) processLogs() {
	// 创建一个定时器，每秒触发一次
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// 用于批量处理的日志缓冲区
	var logBuffer []string

	// 处理日志
	for {
		select {
		case logMsg := <-h.logChan:
			// 添加到缓冲区
			logBuffer = append(logBuffer, logMsg)

			// 如果缓冲区达到一定大小，立即写入
			if len(logBuffer) >= 100 {
				h.writeLogsToFile(logBuffer)
				logBuffer = logBuffer[:0] // 清空缓冲区
			}

		case <-ticker.C:
			// 每秒定时刷新，即使日志数量不多
			if len(logBuffer) > 0 {
				h.writeLogsToFile(logBuffer)
				logBuffer = logBuffer[:0] // 清空缓冲区
			}

		case <-h.shutdown:
			// 处理剩余的日志
			close(h.logChan)

			// 先处理缓冲区中的日志
			if len(logBuffer) > 0 {
				h.writeLogsToFile(logBuffer)
			}

			// 处理通道中剩余的日志
			for logMsg := range h.logChan {
				h.mu.Lock()
				_, _ = h.w.Write([]byte(logMsg))
				h.mu.Unlock()
			}
			return
		}
	}
}

// 将日志批量写入文件
func (h *CustomTextHandler) writeLogsToFile(logs []string) {
	if len(logs) == 0 {
		return
	}

	// 合并多条日志，减少锁的获取次数
	var sb strings.Builder
	for _, log := range logs {
		sb.WriteString(log)
	}

	// 一次性写入所有日志
	h.mu.Lock()
	_, _ = h.w.Write([]byte(sb.String()))
	h.mu.Unlock()
}

func (h *CustomTextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *CustomTextHandler) Handle(ctx context.Context, r slog.Record) error {
	// 构建自定义格式的日志
	var sb strings.Builder
	sb.WriteString("[")

	// 1. 添加时间 (不包含时区)
	sb.WriteString(r.Time.Format("2006-01-02T15:04:05.000"))
	sb.WriteString(" ")

	// 2. 添加日志级别
	sb.WriteString(r.Level.String())
	sb.WriteString(" ")

	// 3. 添加文件位置信息（如果存在）
	fileFound := false
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "file" {
			if s, ok := a.Value.Any().(string); ok {
				sb.WriteString(s)
				sb.WriteString(" ")
				fileFound = true
			}
		}
		return true
	})

	// 如果没有找到文件信息，添加一个占位符
	if !fileFound {
		sb.WriteString("??? ")
	}

	// 4. 添加消息
	sb.WriteString(r.Message)

	// 5. 添加其他属性
	r.Attrs(func(a slog.Attr) bool {
		if a.Key != "file" { // 跳过已处理的文件属性
			sb.WriteString(" ")
			sb.WriteString(a.Key)
			sb.WriteString(" ")

			// 简单处理值
			switch v := a.Value.Any().(type) {
			case string:
				sb.WriteString(v)
			default:
				sb.WriteString(fmt.Sprintf("%v", a.Value.Any()))
			}
		}
		return true
	})

	sb.WriteString("]")

	// 添加换行符
	sb.WriteString("\n")

	h.processLogsSync(sb.String())

	// // 将日志发送到通道
	// select {
	// case h.logChan <- sb.String():
	// 	// 成功发送到通道
	// default:
	// 	// 通道已满，直接写入（避免阻塞）
	// 	h.mu.Lock()
	// 	_, err := h.w.Write([]byte(sb.String()))
	// 	h.mu.Unlock()
	// 	return err
	// }

	return nil
}

// 关闭处理器，确保所有日志都被写入
func (h *CustomTextHandler) Close() {
	close(h.shutdown)
}

func (h *CustomTextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// 简单实现，实际应用中可能需要更复杂的处理
	return h
}

func (h *CustomTextHandler) WithGroup(name string) slog.Handler {
	// 简单实现，实际应用中可能需要更复杂的处理
	return h
}

// LineHandler 添加行号的自定义 Handler
type LineHandler struct {
	slog.Handler
}

func (h *LineHandler) Handle(ctx context.Context, r slog.Record) error {
	// 尝试不同的调用栈深度
	for skip := 3; skip <= 10; skip++ {
		_, file, line, ok := runtime.Caller(skip)
		// 找到第一个非logger包的调用者
		if ok && !strings.Contains(file, "logger/") && !strings.Contains(file, "slog") {
			// 只获取文件名，不要路径
			_, filename := filepath.Split(file)

			// 添加行号属性
			r.AddAttrs(slog.String("file", filename+":"+strconv.Itoa(line)))
			break
		}
	}

	return h.Handler.Handle(ctx, r)
}

// WithAttrs 实现 slog.Handler 接口
func (h *LineHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &LineHandler{Handler: h.Handler.WithAttrs(attrs)}
}

// WithGroup 实现 slog.Handler 接口
func (h *LineHandler) WithGroup(name string) slog.Handler {
	return &LineHandler{Handler: h.Handler.WithGroup(name)}
}

// NewLineHandler 创建带行号的 Handler
func NewLineHandler(h slog.Handler) *LineHandler {
	return &LineHandler{Handler: h}
}

// Init 初始化日志系统，确保只初始化一次
func Init() {
	once.Do(func() {
		// 获取进程PID
		processPID := os.Getpid()

		// 获取启动时间（当前时间）
		startTime := time.Now().Format("01021504") // MMDDHHMM 格式

		// 构建日志文件名
		logFileName := fmt.Sprintf("stock_%d_%s.log", processPID, startTime)

		logFile, err := os.OpenFile(globalDefine.LOG_PATH+logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Printf("无法打开日志文件: %v\n", err)
			return
		}

		// 创建一个多输出写入器，同时写入文件和标准输出
		multiWriter := io.MultiWriter(logFile, os.Stdout)

		// 创建自定义文本处理器
		textHandler := NewCustomTextHandler(multiWriter, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
		defaultHandler = textHandler // 保存处理器实例

		// 添加行号信息
		handler := NewLineHandler(textHandler)

		// 设置全局 logger
		slog.SetDefault(slog.New(handler))

		// 输出一条测试日志，验证行号是否正常
		slog.Info("日志系统初始化完成", "logFile", logFileName)
	})
}

// Close 关闭日志处理器，确保所有日志都被写入
func Close() {
	if defaultHandler != nil {
		defaultHandler.Close()
	}
}

func init() {
	// 在包被导入时自动初始化日志系统
	Init()
}

// 提供一些便捷的日志函数
func Debug(msg string, args ...any) {
	slog.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	slog.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	slog.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	slog.Error(msg, args...)
}

// 添加格式化版本的日志函数
func Debugf(format string, args ...any) {
	slog.Debug(fmt.Sprintf(format, args...))
}

func Infof(format string, args ...any) {
	slog.Info(fmt.Sprintf(format, args...))
}

func Warnf(format string, args ...any) {
	slog.Warn(fmt.Sprintf(format, args...))
}

func Errorf(format string, args ...any) {
	slog.Error(fmt.Sprintf(format, args...))
}
