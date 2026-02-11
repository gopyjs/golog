package golog

import (
	"context"
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	SetLevel(InfoLevel)
	SetCallerShort(true)
	SetOutputJson(true)
	SetName("log_demo")
	SetIsOutputStdout(true)
	SetOutputFile("./log", "demo")
	SetFileRotate(100, 1000, 15)
	AddFieldFunc(func(ctx context.Context, m map[string]interface{}) {
		m["diy_filed"] = ctx.Value("diy")
	})
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

func TestDebug(t *testing.T) {
	SetLevel(DebugLevel)
	SetName("log_demo")
	SetIsOutputStdout(true)
	SetOutputFile("./log", "demo")
	AddFieldFunc(func(ctx context.Context, m map[string]interface{}) {
		m["diy_filed"] = ctx.Value("diy")
	})
	InitLogger()

	ctx := context.WithValue(context.Background(), "diy", []interface{}{"ahhahahahahh"})
	DebugContext(ctx, "tpl:%s", "adAD, DebugContext")
	DebugContext(ctx, "tpl:%s", "adAD, DebugContext")
	DebugContext(ctx, "tpl:%s", "adAD, DebugContext")
	InfoContext(ctx, "tpl:%s", "adAD, InfoContext")
	InfoContext(ctx, "tpl:%s", "adAD, InfoContext")
	ErrorContext(ctx, "tpl:%s", "adAD, ErrorContext")
	// FatalContext(ctx, "tpl:%s", "adAD, FatalContext")
	// PanicContext(ctx, "tpl:%s", "adAD, PanicContext")
}
