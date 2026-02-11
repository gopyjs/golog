# Log the world very easy

[![GitHub forks](https://img.shields.io/github/forks/gopyjs/golog.svg?style=social&label=Forks)](https://github.com/gopyjs/golog/v3/forks)
[![GitHub stars](https://img.shields.io/github/stars/gopyjs/golog.svg?style=social&label=Stars)](https://github.com/gopyjs/golog/v3/stargazers)
[![GitHub last commit](https://img.shields.io/github/last-commit/gopyjs/golog.svg)](https://github.com/gopyjs/golog/v3)
[![GitHub issues](https://img.shields.io/github/issues/gopyjs/golog.svg)](https://github.com/gopyjs/golog/v3/issues)
[![License](https://img.shields.io/badge/license-Apache%202-4EB1BA.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)

Thanks To Uber ZapLog! Log to console or file very easy and fast!

[中文说明](README_ZH.md)

## How to use

Simple：

```
go get -v github.com/gopyjs/golog/v3
```

## Demo

default logger is InfoLevel, and has long func caller.

### Demo1

```go
package main

import . "github.com/gopyjs/golog/v3"

func main() {
	// use default log
	Info("now is Info", 2, " good")
	Debug("now is Debug", 2, " good")
	Warn("now is Warn", 2, " good")
	Error("now is Error", 2, " good")
	Infof("now is Infof: %d,%s", 2, "good")
	Debugf("now is Debugf: %d,%s", 2, "good")
	Warnf("now is Warnf: %d,%s", 2, "good")
	Errorf("now is Errorf: %d,%s", 2, "good")
	Sync()

	// config log
	SetLevel(DebugLevel).SetCallerShort(true).SetOutputJson(true).InitLogger()

	Info("now is Info", 2, " good")
	Debug("now is Debug", 2, " good")
	Warn("now is Warn", 2, " good")
	Error("now is Error", 2, " good")
	Infof("now is Infof: %d,%s", 2, "good")
	Debugf("now is Debugf: %d,%s", 2, "good")
	Warnf("now is Warnf: %d,%s", 2, "good")
	Errorf("now is Errorf: %d,%s", 2, "good")
	Sync()

}
```

Output:

```
2021-08-27T11:16:10.455+0800    INFO    /Users/pika/Documents/code/github/golog/demo/demo1/main.go:7    main.main       now is Info2 good
2021-08-27T11:16:10.455+0800    WARN    /Users/pika/Documents/code/github/golog/demo/demo1/main.go:9    main.main       now is Warn2 good
2021-08-27T11:16:10.455+0800    ERROR   /Users/pika/Documents/code/github/golog/demo/demo1/main.go:10   main.main       now is Error2 good
2021-08-27T11:16:10.455+0800    INFO    /Users/pika/Documents/code/github/golog/demo/demo1/main.go:11   main.main       now is Infof: 2,good
2021-08-27T11:16:10.455+0800    WARN    /Users/pika/Documents/code/github/golog/demo/demo1/main.go:13   main.main       now is Warnf: 2,good
2021-08-27T11:16:10.455+0800    ERROR   /Users/pika/Documents/code/github/golog/demo/demo1/main.go:14   main.main       now is Errorf: 2,good
{"l":"info","t":"2021-08-27T11:16:10.455+0800","caller":"demo1/main.go:19","func":"main.main","msg":"now is Info2 good"}
{"l":"debug","t":"2021-08-27T11:16:10.455+0800","caller":"demo1/main.go:20","func":"main.main","msg":"now is Debug2 good"}
{"l":"warn","t":"2021-08-27T11:16:10.455+0800","caller":"demo1/main.go:21","func":"main.main","msg":"now is Warn2 good"}
{"l":"error","t":"2021-08-27T11:16:10.455+0800","caller":"demo1/main.go:22","func":"main.main","msg":"now is Error2 good"}
{"l":"info","t":"2021-08-27T11:16:10.455+0800","caller":"demo1/main.go:23","func":"main.main","msg":"now is Infof: 2,good"}
{"l":"debug","t":"2021-08-27T11:16:10.455+0800","caller":"demo1/main.go:24","func":"main.main","msg":"now is Debugf: 2,good"}
{"l":"warn","t":"2021-08-27T11:16:10.455+0800","caller":"demo1/main.go:25","func":"main.main","msg":"now is Warnf: 2,good"}
{"l":"error","t":"2021-08-27T11:16:10.455+0800","caller":"demo1/main.go:26","func":"main.main","msg":"now is Errorf: 2,good"}
```

### Demo2

you can config log to file and auto rotate.

```go
package main

import (
	"context"
	"fmt"
	. "github.com/gopyjs/golog/v3"
	"time"
)

func main() {
	SetName("log_demo")
	SetLevel(InfoLevel)

	SetCallerShort(true).SetOutputJson(true)

	AddFieldFunc(func(ctx context.Context, m map[string]interface{}) {
		m["diy_filed"] = ctx.Value("diy")
	})

	SetOutputFile("./log", "demo").SetFileRotate(100, 1000, 15)
	SetIsOutputStdout(true)
	InitLogger()

	Info("now is Info", 2, " good")
	Debug("now is Debug", 2, " good")
	Warn("now is Warn", 2, " good")
	Error("now is Error", 2, " good")
	Infof("now is Infof: %d,%s", 2, "good")
	Debugf("now is Debugf: %d,%s", 2, "good")
	Warnf("now is Warnf: %d,%s", 2, "good")
	Errorf("now is Errorf: %d,%s", 2, "good")

	ctx := context.WithValue(context.Background(), "diy", []interface{}{"ahhahahahahh"})
	InfoContext(ctx, "InfoContext")
	InfoContext(ctx, "InfoContext, %s:InfoContext, %d", "ss", 333)
	InfoWithFields(map[string]interface{}{"k1": "sss"}, "InfoWithFields:%s，%d", "sss", 33333)
	InfoWithFields(map[string]interface{}{"k1": "sss"}, "InfoWithFields")

	err := Sync()
	if err != nil {
		fmt.Println(err.Error())
	}
}
```

Output is in the file dir `·/log`, due to `SetIsOutputStdout` is true it also output to console.

### Demo3: Split Log by Level

By default, all logs are written to a single file (`app.log`). You can also split logs into different files by level:

```go
package main

import (
    . "github.com/gopyjs/golog/v3"
)

func main() {
    SetName("log_demo")
    SetLevel(DebugLevel)
    SetOutputFile("./log", "demo")
    SetFileRotate(100, 1000, 15)

    // Enable split by level (default is false)
    // Each level will be written to its own file only (no duplication)
    SetSplitByLevel(true)

    InitLogger()

    Debug("debug message")   // Written to demo_debug.log only
    Info("info message")     // Written to demo_info.log only
    Warn("warn message")     // Written to demo_warn.log only
    Error("error message")   // Written to demo_error.log only

    Sync()
}
```

When `SetSplitByLevel(true)`:

- Each log level is written to its own file only
- No IO duplication, better performance
- Files: `demo_debug.log`, `demo_info.log`, `demo_warn.log`, `demo_error.log`, etc.

When `SetSplitByLevel(false)` (default):

- All logs are written to a single file
- Better for high-throughput scenarios
- File: `demo.log`

## Usage

very easy to understand.

```go
type LoggerInterface interface {
	SetOutputFile(logPath, fileName string) LoggerInterface
	SetFileRotate(maxSizeMB int, maxBackups int, maxAgeDay int) LoggerInterface
	SetLevel(level Level) LoggerInterface
	SetCallerShort(short bool) LoggerInterface
	SetName(name string) LoggerInterface
	SetIsOutputStdout(isOutputStdout bool) LoggerInterface
	SetCallerSkip(skip int) LoggerInterface
	SetOutputJson(json bool) LoggerInterface
	SetSplitByLevel(split bool) LoggerInterface

	GetOutputFile() (logPath, fileName string)
	GetSplitByLevel() (split bool)
	GetFileRotate() (maxSizeMB int, maxBackups int, maxAgeDay int)
	GetLevel() (level Level)
	GetCallerShort() (short bool)
	GetName() (name string)
	GetIsOutputStdout() (isOutputStdout bool)
	GetCallerSkip() (skip int)
	GetOutputJson() bool

	// InitLogger init logger should call this when change config
	InitLogger()
	// Sync terminal the logger should call this to flush
	Sync() error

	Panicf(template string, args ...interface{})
	Fatalf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Debugf(template string, args ...interface{})

	Panic(args ...interface{})
	Fatal(args ...interface{})
	Error(args ...interface{})
	Warn(args ...interface{})
	Info(args ...interface{})
	Debug(args ...interface{})

	PanicWithFields(fields map[string]interface{}, template string, args ...interface{})
	FatalWithFields(fields map[string]interface{}, template string, args ...interface{})
	ErrorWithFields(fields map[string]interface{}, template string, args ...interface{})
	WarnWithFields(fields map[string]interface{}, template string, args ...interface{})
	InfoWithFields(fields map[string]interface{}, template string, args ...interface{})
	DebugWithFields(fields map[string]interface{}, template string, args ...interface{})

	PanicContext(ctx context.Context, template string, args ...interface{})
	FatalContext(ctx context.Context, template string, args ...interface{})
	ErrorContext(ctx context.Context, template string, args ...interface{})
	WarnContext(ctx context.Context, template string, args ...interface{})
	InfoContext(ctx context.Context, template string, args ...interface{})
	DebugContext(ctx context.Context, template string, args ...interface{})

	PanicContextWithFields(ctx context.Context, fields map[string]interface{}, template string, args ...interface{})
	FatalContextWithFields(ctx context.Context, fields map[string]interface{}, template string, args ...interface{})
	ErrorContextWithFields(ctx context.Context, fields map[string]interface{}, template string, args ...interface{})
	WarnContextWithFields(ctx context.Context, fields map[string]interface{}, template string, args ...interface{})
	InfoContextWithFields(ctx context.Context, fields map[string]interface{}, template string, args ...interface{})
	DebugContextWithFields(ctx context.Context, fields map[string]interface{}, template string, args ...interface{})

	// AddFieldFunc filter deal the fields
	AddFieldFunc(func(context.Context, map[string]interface{}))

	GetZapLogger() *zap.Logger
	GetZapSugaredLogger() *zap.SugaredLogger
}
```

# License

```
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

Copyright [2021-2021] [github.com/hunterhug] v2
Copyright [2026-2030] [github.com/gopyjs] v3
```
