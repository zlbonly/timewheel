package main

import (
	"fmt"
	"time"
	"timewheel"
)

var interval = 1
var soltNum = 2

func test(data interface{}) {
	res := data.(map[string]int)
	fmt.Println("延迟任务执行 uid = ?", res["uid"])
}

func main() {
	tw := timewheel.New(1*time.Second, soltNum, test)
	tw.StartWheel()
	tw.AddTimer(3*time.Second, "ccc", map[string]int{"uid": 1020000})
	select {}

}
