package main

import (
	"context"
	"fmt"

	. "github.com/gopyjs/golog/v3"
)

func main() {
	SetName("log_demo")
	SetLevel(InfoLevel)

	SetCallerShort(true).SetOutputJson(true)

	AddFieldFunc(func(ctx context.Context, m map[string]interface{}) {
		m["diy_filed"] = ctx.Value("diy")
	})

	SetOutputFile("./log", "demo").SetFileRotate(1, 5, 10)
	SetIsOutputStdout(true)
	InitLogger()

	i := 0
	for {
		i = i + 1
		if i > 1000 {
			break
		}
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
	}

	err := Sync()
	if err != nil {
		fmt.Println(err.Error())
	}
}
