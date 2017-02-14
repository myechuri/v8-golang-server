package main

import (
        "runtime"
        "fmt"
	"C"
        "github.com/lazytiger/go-v8"
)

func main() {
}

//export GoMain
func GoMain() {
	fmt.Println("Hello, 世界");
	fmt.Println("Go version:", runtime.Version());
            engine := v8.NewEngine()
    script := engine.Compile([]byte("'Hello ' + 'World!'"), nil, nil)
    context := engine.NewContext(nil)

    context.Scope(func(cs v8.ContextScope) {
        result := script.Run()
        println(result.ToString())
    })
}
