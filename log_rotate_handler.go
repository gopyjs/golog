package golog

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// LogRotateHandler 日志文件大小轮换管理器
// ; 由于 golog 不支持按文件大小轮换，我们使用定时任务监控日志文件大小
// ; 当日志文件超过指定大小时，手动触发轮换
type LogRotateHandler struct {
	logDir    string        // 日志目录
	globExp   string        // 日志文件名模式
	maxSizeMB int           // 最大文件大小（字节）
	checkMins int           // 检查周期, 单位分钟
	stopChan  chan struct{} // 停止信号
}

// InitLogRotateHandler 初始化日志轮换监控器
// ; 在 InitGolog() 之后调用
func InitLogRotateHandler(logDir string, maxSizeMB int, checkMins int) *LogRotateHandler {
	// sample: InitLogRotateHandler("logs", 50, 15)
	//; 示例: 配置：最大 50MB，每 15 分钟检查一次
	handler := NewLogRotateHandler(logDir, "*.log", maxSizeMB, checkMins)
	handler.Start()
	return handler
}

// NewLogRotateHandler 创建日志轮换监控器
func NewLogRotateHandler(logDir string, globExp string, maxSizeMB int, checkMins int) *LogRotateHandler {
	return &LogRotateHandler{
		logDir:    logDir,
		globExp:   globExp,
		maxSizeMB: maxSizeMB,
		checkMins: checkMins,
		stopChan:  make(chan struct{}),
	}
}

// Start 启动监控
func (m *LogRotateHandler) Start() {
	go m.run()
	Infof("LogRotateHandler started: dir=%s, globExp=%s, maxSize=%dMB, checkMins=%d",
		m.logDir, m.globExp, m.maxSizeMB/(1024*1024), m.checkMins)
}

// Stop 停止监控
func (m *LogRotateHandler) Stop() {
	close(m.stopChan)
	Infof("LogRotateHandler stopped")
}

// run 监控循环
func (m *LogRotateHandler) run() {
	checkPeriod := time.Duration(m.checkMins) * time.Minute
	ticker := time.NewTicker(checkPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.checkAndRotate()
		}
	}
}

// checkAndRotate 检查并轮换日志文件
func (m *LogRotateHandler) checkAndRotate() {
	//; 查找所有日志文件
	pattern := filepath.Join(m.logDir, m.globExp)
	maxSize := int64(m.maxSizeMB) * 1024 * 1024
	checkSs := time.Duration(m.checkMins) * time.Minute

	files, err := filepath.Glob(pattern)
	if err != nil {
		Errorf(fmt.Sprintf("Failed to glob log files: %v", err))
		return
	}

	for _, file := range files {
		// //; 跳过已经轮换的文件（带时间戳的）
		// if filepath.Ext(file) != ".log" {
		// 	continue
		// }

		//; 检查文件大小
		info, err := os.Stat(file)
		if err != nil {
			Errorf(fmt.Sprintf("Failed to stat log file %s: %v", file, err))
			continue
		}

		//; 检查文件是否在最近周期内修改
		if info.ModTime().Add(checkSs).Before(time.Now()) {
			continue
		}

		if info.Size() >= maxSize {
			//; 文件超过大小限制，触发轮换
			m.rotateFile(file, info.Size())
		}
	}
}

// rotateFile 轮换日志文件 (copy + truncate 模式)
// ; 使用 copy+truncate 而非 rename，避免上游持有的文件描述符失效导致日志写入失败
func (m *LogRotateHandler) rotateFile(oldFp string, fileSize int64) {
	//; 生成备份文件名（添加时间戳）
	timestamp := time.Now().Format("20060102_150405")
	logdir := filepath.Dir(oldFp)
	logFn := filepath.Base(oldFp)
	ext := filepath.Ext(logFn)
	name := logFn[:len(logFn)-len(ext)]
	//; 提取基础文件名（去掉之前的时间戳）
	name = extractBaseFileName(name)

	backupFp := filepath.Join(logdir, fmt.Sprintf("%s.%s%s", name, timestamp, ext))

	//; 打开原文件（保持文件描述符不变）
	file, err := os.OpenFile(oldFp, os.O_RDWR, 0644)
	if err != nil {
		Errorf("Failed to open log file for rotate: %v", err)
		return
	}
	defer file.Close()

	//; copy 内容到备份文件
	backupFile, err := os.Create(backupFp)
	if err != nil {
		Errorf("Failed to create backup file %s: %v", backupFp, err)
		return
	}

	_, err = io.Copy(backupFile, file)
	backupFile.Close()
	if err != nil {
		Errorf("Failed to copy log content to backup: %v", err)
		return
	}

	//; truncate 原文件为 0 字节（清空内容，但保持 fd 有效）
	err = file.Truncate(0)
	if err != nil {
		Errorf("Failed to truncate log file: %v", err)
		return
	}

	//; 重置文件指针到开头
	_, err = file.Seek(0, 0)
	if err != nil {
		Errorf("Failed to seek log file: %v", err)
		return
	}

	Infof("Log file rotated: %s -> %s (size: %.2fMB)",
		oldFp, backupFp, float64(fileSize)/(1024*1024))
}

// extractBaseFileName 提取基础文件名
// ; 从可能包含多次时间戳的文件名中提取基础名称
// ; 例如: "gorm-sql-20260328_1811.20260329_021133" -> "gorm-sql"
func extractBaseFileName(name string) string {
	//; 匹配时间戳模式: .YYYYMMDD_HHMMSS 或 -YYYYMMDD_HHMMSS
	//; 支持多种分隔符: . - _
	tsPattern1 := regexp.MustCompile(`[._-]\d{8}_\d{6}`)
	tsPattern2 := regexp.MustCompile(`[._-]\d{6}_\d{6}`)

	//; 移除所有时间戳后缀
	base := tsPattern1.ReplaceAllString(name, "")
	base = tsPattern2.ReplaceAllString(base, "")

	//; 如果结果为空，返回原始名称
	if strings.TrimSpace(base) == "" {
		return name
	}

	//; 移除尾部可能残留的 . 或 -
	base = strings.TrimRight(base, ".-")

	return base
}
