package golog

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder        // 设置时间格式
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // 设置大写字母记录日志级别,以及不同颜色
	// encoderConfig.EncodeCaller = zapcore.FullCallerEncoder       //显示完整文件路径
	return zapcore.NewConsoleEncoder(encoderConfig)
}

// customCallerEncoder 根据 verbose 级别格式化调用者信息
// verbose = 0: 相对于 go mod 的短路径，如 "util/help.go:85"
// verbose = 1: 完整包名路径，如 "github.com/u/repo/util/help.go:85"
// verbose = 2: 绝对文件路径，如 "/home/u/repo/src/util/help.go:85"
func customCallerEncoder(verbose int) zapcore.CallerEncoder {
	return func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		switch verbose {
		case 0:
			//; 只显示相对路径（去掉模块名前缀）
			enc.AppendString(caller.TrimmedPath())
		case 2:
			//; 显示完整绝对路径
			enc.AppendString(caller.File)
		default:
			//; verbose = 1 或其他值，显示完整包路径
			enc.AppendString(caller.FullPath())
		}
	}
}

// InitZapLoggerEx 初始化 zap 日志（扩展版），同时输出到控制台和文件
// logLevel: debug 或 info（其他值默认为 info）
// logFilePath: 日志文件路径，如果为空则只输出到控制台
// verbose: 0=短路径, 1=完整包路径(默认), 2=绝对路径
//
// ;@NOTE: 如需更简单的初始化（仅控制台），使用 InitZapLogger()
func InitZapLoggerEx(logLevel string, verbose int, stdout bool, logFilePath string) (*zap.Logger, error) {
	//; 创建日志文件（如果指定了路径）
	var logFile *os.File
	var err error
	if logFilePath != "" {
		logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf(fmt.Sprintf("failed to open zap log file, path = %s, err = %v", logFilePath, err))
		}
	}

	//; 配置编码器（控制台 + 文件使用相同格式）
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoderConfig.EncodeCaller = customCallerEncoder(verbose)
	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	//; 设置日志级别
	var level zapcore.Level
	if strings.ToLower(logLevel) == "debug" {
		level = zapcore.DebugLevel
	} else {
		level = zapcore.InfoLevel
	}

	//; 创建多输出 Syncer（控制台 + 文件）
	writeSyncers := []zapcore.WriteSyncer{}
	// writeSyncers := []zapcore.WriteSyncer{zapcore.AddSync(os.Stdout)}
	if !stdout {
		writeSyncers = append(writeSyncers, zapcore.AddSync(os.Stdout))
	}
	if logFile != nil {
		writeSyncers = append(writeSyncers, zapcore.AddSync(logFile))
	}
	multiWriteSyncer := zapcore.NewMultiWriteSyncer(writeSyncers...)

	//; 创建核心和 logger
	core := zapcore.NewCore(encoder, multiWriteSyncer, level)
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	zap.ReplaceGlobals(logger)

	if logFilePath != "" {
		logger.Info(fmt.Sprintf("zap logger initialized, log file: %s", logFilePath))
	} else {
		logger.Info("zap logger initialized, console only")
	}

	return logger, nil
}
