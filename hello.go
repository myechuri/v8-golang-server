package main

import (
	"C"
	"fmt"
	"github.com/ry/v8worker"
	"runtime"
)

func main() {
}

func DiscardSendSync(msg string) string { return "" }

//export GoMain
func GoMain() {
	fmt.Println("Hello, 世界")
	fmt.Println("Go version:", runtime.Version())

	worker := v8worker.New(func(msg string) {
		println("recv cb", msg)
		if msg != "hello" {
			fmt.Println("bad msg", msg)
		}
	}, DiscardSendSync)

	code := ` $print("ready"); `
	err := worker.Load("code.js", code)
	if err != nil {
		fmt.Println(err.Error())
	}
}
