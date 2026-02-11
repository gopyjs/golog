package golog

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type logger struct {
	name         string
	zapLogger    *zap.Logger
	sugarLog     *zap.SugaredLogger
	level        Level
	short        bool
	json         bool
	addFieldFunc func(context.Context, map[string]interface{})

	logPath  string
	fileName string

	maxSizeMB  int
	maxBackups int
	maxAgeDay  int

	skip           int
	isOutputStdout bool
	isOutputFile   bool
	splitByLevel   bool
}

var _log = New()

func init() {
	// skip is 2, we wrap 2 layer
	_log.SetCallerSkip(2)
	_log.InitLogger()
}

// Logger default log which output to console
// if you want to log to file you must New() and set something then call InitLogger()
func Logger() LoggerInterface {
	return _log
}

// New can new a logger interface, you can config it by it's method
func New() LoggerInterface {
	l := new(logger)
	l.level = InfoLevel
	l.short = false
	l.json = false
	l.splitByLevel = false
	l.isOutputFile = false
	l.isOutputStdout = true
	return l
}

// ; levelConfig 定义按级别拆分日志时的级别配置表
var levelConfigs = []struct {
	level       Level
	suffix      string
	defaultName string
}{
	{DebugLevel, "_debug", "access.log"},
	{InfoLevel, "_info", "info.log"},
	{WarnLevel, "_warn", "warn.log"},
	{ErrorLevel, "_err", "error.log"},
	{DPanicLevel, "_dpanic", "dpanic.log"},
	{FatalLevel, "_fatal", "fatal.log"},
	{PanicLevel, "_panic", "panic.log"},
}

// ; buildEncoder 根据配置创建 zap encoder
func (l *logger) buildEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.LevelKey = "l"
	encoderConfig.FunctionKey = "func"
	encoderConfig.CallerKey = "caller"
	encoderConfig.TimeKey = "t"
	encoderConfig.MessageKey = "msg"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	if l.short {
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	} else {
		encoderConfig.EncodeCaller = zapcore.FullCallerEncoder
	}
	encoderConfig.LineEnding = zapcore.DefaultLineEnding

	if !l.json {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		return zapcore.NewConsoleEncoder(encoderConfig)
	}
	return zapcore.NewJSONEncoder(encoderConfig)
}

// ; createFileCore 创建文件输出的 zapcore.Core（范围匹配：>= minLevel）
func (l *logger) createFileCore(filename string, minLevel Level) zapcore.Core {
	writer := getWriter(false, filename, l.maxSizeMB, l.maxBackups, l.maxAgeDay)
	return zapcore.NewCore(
		l.buildEncoder(),
		zapcore.AddSync(writer),
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= minLevel
		}),
	)
}

// ; createExactLevelCore 创建精确级别匹配的 zapcore.Core（只接收指定级别）
func (l *logger) createExactLevelCore(filename string, exactLevel Level) zapcore.Core {
	writer := getWriter(false, filename, l.maxSizeMB, l.maxBackups, l.maxAgeDay)
	return zapcore.NewCore(
		l.buildEncoder(),
		zapcore.AddSync(writer),
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl == exactLevel
		}),
	)
}

// ; createStdoutCore 创建 stdout 输出的 zapcore.Core
func (l *logger) createStdoutCore() zapcore.Core {
	encoder := l.buildEncoder()
	return zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= l.level
		}),
	)
}

// ; buildCores 构建所有日志输出 core
func (l *logger) buildCores() []zapcore.Core {
	cores := make([]zapcore.Core, 0)

	if l.logPath == "" {
		//; 无文件输出路径，仅输出到 stdout
		return []zapcore.Core{l.createStdoutCore()}
	}

	if l.splitByLevel {
		//; 按级别拆分文件：精确匹配，每个级别只写入自己的文件（无重复）
		for _, cfg := range levelConfigs {
			if l.level <= cfg.level {
				filename := l.buildLogFileName(cfg.suffix, cfg.defaultName)
				cores = append(cores, l.createExactLevelCore(filename, cfg.level))
			}
		}
	} else {
		//; 不区分级别，所有日志写入同一个文件
		filename := l.buildLogFileName("", "app.log")
		cores = append(cores, l.createFileCore(filename, l.level))
	}

	if l.isOutputStdout {
		cores = append(cores, l.createStdoutCore())
	}

	return cores
}

// ; buildLogFileName 构建日志文件名
func (l *logger) buildLogFileName(suffix string, defaultName string) string {
	if l.fileName != "" {
		if suffix != "" {
			return filepath.Join(l.logPath, l.fileName+suffix+".log")
		}
		return filepath.Join(l.logPath, l.fileName+".log")
	}
	return filepath.Join(l.logPath, defaultName)
}

// InitLogger after config you must call this method
func (l *logger) InitLogger() {
	cores := l.buildCores()
	outCore := zapcore.NewTee(cores...)

	op1 := zap.AddCaller()

	// we wrap 1 layer
	op2 := zap.AddCallerSkip(1)

	if l.skip > 0 {
		op2 = zap.AddCallerSkip(l.skip)
	}

	zapLogger := zap.New(outCore, op1, op2)
	if l.name != "" {
		zapLogger = zapLogger.Named(l.name)
	}
	sugarLogger := zapLogger.Sugar()
	l.zapLogger = zapLogger
	l.sugarLog = sugarLogger
}

func InitLogger() {
	_log.InitLogger()
}

func getWriter(isOutputStdout bool, filename string, maxSizeMB int, maxBackups int, maxAgeDay int) io.Writer {
	maxAgeDays := 30
	if maxAgeDay > 0 {
		maxAgeDays = maxAgeDay
	}

	hook := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSizeMB,
		MaxBackups: maxBackups,
		MaxAge:     maxAgeDays,
		Compress:   false,
		LocalTime:  true,
	}

	if isOutputStdout {
		return io.MultiWriter(os.Stdout, hook)
	}

	return hook
}

func (l *logger) Sync() error {
	err := l.sugarLog.Sync()
	if !errors.Is(err, os.ErrInvalid) {
		return nil
	}

	return nil
}

func Sync() error {
	return _log.Sync()
}

func SetName(name string) LoggerInterface {
	return _log.SetName(name)
}

func (l *logger) SetName(name string) LoggerInterface {
	l.name = name
	return l
}

func GetName() (name string) {
	return _log.GetName()
}

func (l *logger) GetName() (name string) {
	return l.name
}

func SetCallerCallerSkip(skip int) LoggerInterface {
	return _log.SetCallerSkip(skip)
}

func (l *logger) SetCallerSkip(skip int) LoggerInterface {
	l.skip = skip
	return l
}

func GetCallerSkip() (skip int) {
	return _log.GetCallerSkip()
}

func (l *logger) GetCallerSkip() (skip int) {
	return l.skip
}

func SetIsOutputStdout(isOutputStdout bool) LoggerInterface {
	return _log.SetIsOutputStdout(isOutputStdout)
}

func (l *logger) SetIsOutputStdout(isOutputStdout bool) LoggerInterface {
	l.isOutputStdout = isOutputStdout
	return l
}

func GetIsOutputStdout() (isOutputStdout bool) {
	return _log.GetIsOutputStdout()
}

func (l *logger) GetIsOutputStdout() (isOutputStdout bool) {
	return l.isOutputStdout
}

func SetSplitByLevel(split bool) LoggerInterface {
	return _log.SetSplitByLevel(split)
}

func (l *logger) SetSplitByLevel(split bool) LoggerInterface {
	l.splitByLevel = split
	return l
}

func GetSplitByLevel() (split bool) {
	return _log.GetSplitByLevel()
}

func (l *logger) GetSplitByLevel() (split bool) {
	return l.splitByLevel
}

func SetFileRotate(maxSizeMB int, maxBackups int, maxAgeDay int) LoggerInterface {
	return _log.SetFileRotate(maxSizeMB, maxBackups, maxAgeDay)
}

func (l *logger) SetFileRotate(maxSizeMB int, maxBackups int, maxAgeDay int) LoggerInterface {
	if maxSizeMB <= 0 {
		maxSizeMB = 200
	}

	if maxBackups <= 0 {
		maxBackups = 1000
	}

	if maxAgeDay <= 0 {
		maxAgeDay = 7
	}

	l.maxAgeDay = maxAgeDay
	l.maxSizeMB = maxSizeMB
	l.maxBackups = maxBackups

	return l
}

func GetFileRotate() (maxSizeMB int, maxBackups int, maxAgeDay int) {
	return _log.GetFileRotate()
}

func (l *logger) GetFileRotate() (maxSizeMB int, maxBackups int, maxAgeDay int) {
	return l.maxSizeMB, l.maxBackups, l.maxAgeDay
}

func SetCallerShort(short bool) LoggerInterface {
	return _log.SetCallerShort(short)
}

func (l *logger) SetCallerShort(short bool) LoggerInterface {
	l.short = short
	return l
}

func GetCallerShort() (short bool) {
	return _log.GetCallerShort()
}

func (l *logger) GetCallerShort() (short bool) {
	return l.short
}

func SetOutputJson(json bool) LoggerInterface {
	return _log.SetOutputJson(json)
}

func (l *logger) SetOutputJson(json bool) LoggerInterface {
	l.json = json
	return l
}

func GetOutputJson() (json bool) {
	return _log.GetOutputJson()
}

func (l *logger) GetOutputJson() (json bool) {
	return l.json
}

func SetLevel(level Level) LoggerInterface {
	return _log.SetLevel(level)
}

func (l *logger) SetLevel(level Level) LoggerInterface {
	l.level = level
	return l
}

func GetLevel() (level Level) {
	return _log.GetLevel()
}

func (l *logger) GetLevel() (level Level) {
	return l.level
}

func (l *logger) SetOutputFile(logPath, fileName string) LoggerInterface {
	l.logPath = logPath
	l.fileName = fileName
	if l.maxSizeMB <= 0 || l.maxBackups <= 0 || l.maxAgeDay <= 0 {
		l.SetFileRotate(l.maxSizeMB, l.maxBackups, l.maxAgeDay)
	}
	return l
}

func SetOutputFile(logPath, fileName string) LoggerInterface {
	return _log.SetOutputFile(logPath, fileName)
}

func (l *logger) GetOutputFile() (logPath, fileName string) {
	return l.logPath, l.fileName
}

func GetOutputFile() (logPath, fileName string) {
	return _log.GetOutputFile()
}

func (l *logger) Fatalf(template string, args ...interface{}) {
	l.sugarLog.Fatalf(template, args...)
}

func Fatalf(template string, args ...interface{}) {
	_log.Fatalf(template, args...)
}

func (l *logger) Fatal(args ...interface{}) {
	l.sugarLog.Fatal(args...)
}

func Fatal(args ...interface{}) {
	_log.Fatal(args...)
}

func (l *logger) Panicf(template string, args ...interface{}) {
	l.sugarLog.With().Panicf(template, args...)
}

func Panicf(template string, args ...interface{}) {
	_log.Panicf(template, args...)
}

func (l *logger) Panic(args ...interface{}) {
	l.sugarLog.Panic(args...)
}

func Panic(args ...interface{}) {
	_log.Panic(args...)
}

func (l *logger) Errorf(template string, args ...interface{}) {
	l.sugarLog.Errorf(template, args...)
}

func Errorf(template string, args ...interface{}) {
	_log.Errorf(template, args...)
}

func (l *logger) Error(args ...interface{}) {
	l.sugarLog.Error(args...)
}

func Error(args ...interface{}) {
	_log.Error(args...)
}

func (l *logger) Warnf(template string, args ...interface{}) {
	l.sugarLog.Warnf(template, args...)
}

func Warnf(template string, args ...interface{}) {
	_log.Warnf(template, args...)
}

func (l *logger) Warn(args ...interface{}) {
	l.sugarLog.Warn(args...)
}

func Warn(args ...interface{}) {
	_log.Warn(args...)
}

func (l *logger) Infof(template string, args ...interface{}) {
	l.sugarLog.Infof(template, args...)
}

func Infof(template string, args ...interface{}) {
	_log.Infof(template, args...)
}

func (l *logger) Info(args ...interface{}) {
	l.sugarLog.Info(args...)
}

func Info(args ...interface{}) {
	_log.Info(args...)
}

func (l *logger) Debugf(template string, args ...interface{}) {
	l.sugarLog.Debugf(template, args...)
}

func Debugf(template string, args ...interface{}) {
	_log.Debugf(template, args...)
}

func (l *logger) Debug(args ...interface{}) {
	l.sugarLog.Debug(args...)
}

func Debug(args ...interface{}) {
	_log.Debug(args...)
}

func with(fields map[string]interface{}) []interface{} {
	i := make([]interface{}, 0, 2*len(fields))
	keys := make([]string, 0, len(fields))
	for k, _ := range fields {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for _, v := range keys {
		i = append(i, v, fields[v])
	}
	return i
}

func (l *logger) DebugWithFields(fields map[string]interface{}, template string, args ...interface{}) {
	if len(args) == 0 {
		l.sugarLog.With(with(fields)...).Debug(template)
		return
	}
	l.sugarLog.With(with(fields)...).Debugf(template, args...)
}

func DebugWithFields(fields map[string]interface{}, template string, args ...interface{}) {
	_log.DebugWithFields(fields, template, args...)
}

func (l *logger) InfoWithFields(fields map[string]interface{}, template string, args ...interface{}) {
	if len(args) == 0 {
		l.sugarLog.With(with(fields)...).Info(template)
		return
	}
	l.sugarLog.With(with(fields)...).Infof(template, args...)
}

func InfoWithFields(fields map[string]interface{}, template string, args ...interface{}) {
	_log.InfoWithFields(fields, template, args...)
}

func (l *logger) WarnWithFields(fields map[string]interface{}, template string, args ...interface{}) {
	if len(args) == 0 {
		l.sugarLog.With(with(fields)...).Warn(template)
		return
	}
	l.sugarLog.With(with(fields)...).Warnf(template, args...)
}

func WarnWithFields(fields map[string]interface{}, template string, args ...interface{}) {
	_log.WarnWithFields(fields, template, args...)
}

func (l *logger) ErrorWithFields(fields map[string]interface{}, template string, args ...interface{}) {
	if len(args) == 0 {
		l.sugarLog.With(with(fields)...).Error(template)
		return
	}
	l.sugarLog.With(with(fields)...).Errorf(template, args...)
}

func ErrorWithFields(fields map[string]interface{}, template string, args ...interface{}) {
	_log.ErrorWithFields(fields, template, args...)
}

func (l *logger) FatalWithFields(fields map[string]interface{}, template string, args ...interface{}) {
	if len(args) == 0 {
		l.sugarLog.With(with(fields)...).Fatal(template)
		return
	}
	l.sugarLog.With(with(fields)...).Fatalf(template, args...)
}

func FatalWithFields(fields map[string]interface{}, template string, args ...interface{}) {
	_log.FatalWithFields(fields, template, args...)
}

func (l *logger) PanicWithFields(fields map[string]interface{}, template string, args ...interface{}) {
	if len(args) == 0 {
		l.sugarLog.With(with(fields)...).Panic(template)
		return
	}
	l.sugarLog.With(with(fields)...).Panicf(template, args...)
}

func PanicWithFields(fields map[string]interface{}, template string, args ...interface{}) {
	_log.PanicWithFields(fields, template, args...)
}

func AddFieldFunc(f func(context.Context, map[string]interface{})) {
	_log.AddFieldFunc(f)
}

func (l *logger) AddFieldFunc(f func(context.Context, map[string]interface{})) {
	l.addFieldFunc = f
	return
}

func (l *logger) addField(ctx context.Context, fields map[string]interface{}) {
	//fields["service.log.name"] = l.name
	//fields["service.log.time"] = time.Now().String()

	if l.addFieldFunc != nil {
		l.addFieldFunc(ctx, fields)
	}
}

func (l *logger) DebugContextWithFields(ctx context.Context, fields map[string]interface{}, template string, args ...interface{}) {
	l.addField(ctx, fields)
	if len(args) == 0 {
		l.sugarLog.With(with(fields)...).Debug(template)
		return
	}
	l.sugarLog.With(with(fields)...).Debugf(template, args...)
}

func DebugContextWithFields(ctx context.Context, fields map[string]interface{}, template string, args ...interface{}) {
	_log.DebugContextWithFields(ctx, fields, template, args...)
}

func (l *logger) InfoContextWithFields(ctx context.Context, fields map[string]interface{}, template string, args ...interface{}) {
	l.addField(ctx, fields)
	if len(args) == 0 {
		l.sugarLog.With(with(fields)...).Info(template)
		return
	}
	l.sugarLog.With(with(fields)...).Infof(template, args...)
}

func InfoContextWithFields(ctx context.Context, fields map[string]interface{}, template string, args ...interface{}) {
	_log.InfoContextWithFields(ctx, fields, template, args...)
}

func (l *logger) WarnContextWithFields(ctx context.Context, fields map[string]interface{}, template string, args ...interface{}) {
	l.addField(ctx, fields)
	if len(args) == 0 {
		l.sugarLog.With(with(fields)...).Warn(template)
		return
	}
	l.sugarLog.With(with(fields)...).Warnf(template, args...)
}

func WarnContextWithFields(ctx context.Context, fields map[string]interface{}, template string, args ...interface{}) {
	_log.WarnContextWithFields(ctx, fields, template, args...)
}

func (l *logger) ErrorContextWithFields(ctx context.Context, fields map[string]interface{}, template string, args ...interface{}) {
	l.addField(ctx, fields)
	if len(args) == 0 {
		l.sugarLog.With(with(fields)...).Error(template)
		return
	}
	l.sugarLog.With(with(fields)...).Errorf(template, args...)
}

func ErrorContextWithFields(ctx context.Context, fields map[string]interface{}, template string, args ...interface{}) {
	_log.ErrorContextWithFields(ctx, fields, template, args...)
}

func (l *logger) FatalContextWithFields(ctx context.Context, fields map[string]interface{}, template string, args ...interface{}) {
	l.addField(ctx, fields)
	if len(args) == 0 {
		l.sugarLog.With(with(fields)...).Fatal(template)
		return
	}
	l.sugarLog.With(with(fields)...).Fatalf(template, args...)
}

func FatalContextWithFields(ctx context.Context, fields map[string]interface{}, template string, args ...interface{}) {
	_log.FatalContextWithFields(ctx, fields, template, args...)
}

func (l *logger) PanicContextWithFields(ctx context.Context, fields map[string]interface{}, template string, args ...interface{}) {
	l.addField(ctx, fields)
	if len(args) == 0 {
		l.sugarLog.With(with(fields)...).Panic(template)
		return
	}
	l.sugarLog.With(with(fields)...).Panicf(template, args...)
}

func PanicContextWithFields(ctx context.Context, fields map[string]interface{}, template string, args ...interface{}) {
	_log.PanicContextWithFields(ctx, fields, template, args...)
}

func (l *logger) DebugContext(ctx context.Context, template string, args ...interface{}) {
	fields := make(map[string]interface{})
	l.addField(ctx, fields)
	if len(args) == 0 {
		l.sugarLog.With(with(fields)...).Debug(template)
		return
	}
	l.sugarLog.With(with(fields)...).Debugf(template, args...)
}

func DebugContext(ctx context.Context, template string, args ...interface{}) {
	_log.DebugContext(ctx, template, args...)
}

func (l *logger) InfoContext(ctx context.Context, template string, args ...interface{}) {
	fields := make(map[string]interface{})
	l.addField(ctx, fields)
	if len(args) == 0 {
		l.sugarLog.With(with(fields)...).Info(template)
		return
	}
	l.sugarLog.With(with(fields)...).Infof(template, args...)
}

func InfoContext(ctx context.Context, template string, args ...interface{}) {
	_log.InfoContext(ctx, template, args...)
}

func (l *logger) WarnContext(ctx context.Context, template string, args ...interface{}) {
	fields := make(map[string]interface{})
	l.addField(ctx, fields)
	if len(args) == 0 {
		l.sugarLog.With(with(fields)...).Warn(template)
		return
	}
	l.sugarLog.With(with(fields)...).Warnf(template, args...)
}

func WarnContext(ctx context.Context, template string, args ...interface{}) {
	_log.WarnContext(ctx, template, args...)
}

func (l *logger) ErrorContext(ctx context.Context, template string, args ...interface{}) {
	fields := make(map[string]interface{})
	l.addField(ctx, fields)
	if len(args) == 0 {
		l.sugarLog.With(with(fields)...).Error(template)
		return
	}
	l.sugarLog.With(with(fields)...).Errorf(template, args...)
}

func ErrorContext(ctx context.Context, template string, args ...interface{}) {
	_log.ErrorContext(ctx, template, args...)
}

func (l *logger) FatalContext(ctx context.Context, template string, args ...interface{}) {
	fields := make(map[string]interface{})
	l.addField(ctx, fields)
	if len(args) == 0 {
		l.sugarLog.With(with(fields)...).Fatal(template)
		return
	}
	l.sugarLog.With(with(fields)...).Fatalf(template, args...)
}

func FatalContext(ctx context.Context, template string, args ...interface{}) {
	_log.FatalContext(ctx, template, args...)
}

func (l *logger) PanicContext(ctx context.Context, template string, args ...interface{}) {
	fields := make(map[string]interface{})
	l.addField(ctx, fields)
	if len(args) == 0 {
		l.sugarLog.With(with(fields)...).Panic(template)
		return
	}
	l.sugarLog.With(with(fields)...).Panicf(template, args...)
}

func PanicContext(ctx context.Context, template string, args ...interface{}) {
	_log.PanicContext(ctx, template, args...)
}

func (l *logger) GetZapLogger() *zap.Logger {
	return l.zapLogger
}

func GetZapLogger() *zap.Logger {
	return _log.GetZapLogger()
}

func GetZapSugaredLogger() *zap.SugaredLogger {
	return _log.GetZapSugaredLogger()
}

func (l *logger) GetZapSugaredLogger() *zap.SugaredLogger {
	return l.sugarLog
}
