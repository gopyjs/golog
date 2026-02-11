package main

import (
	"fmt"

	. "github.com/gopyjs/golog/v3"
)

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
	err := Sync()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

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
	err = Sync()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

}
