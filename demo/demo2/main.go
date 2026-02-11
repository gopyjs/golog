package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

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

	// Ensure Close() is called on exit to flush buffers and release resources
	defer func() {
		Info("Shutting down, closing logger...")
		if err := Close(); err != nil {
			fmt.Printf("Error closing logger: %v\n", err)
		}
	}()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Simulate web service work
	i := 0
	for {
		select {
		case sig := <-sigChan:
			Info("Received signal", sig)
			return
		default:
			i++
			if i > 1000 {
				return
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
	}
}
