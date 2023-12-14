package main

import (
	"time"

	"github.com/1f604/util"
)

func main() {
	f := func() {
		println("Hello! Sleeping for 2 seconds!")
		time.Sleep(2 * time.Second)
	}
	go util.RunFuncEveryXSeconds(f, 4)
	time.Sleep(1 * time.Hour)
}
