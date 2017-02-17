package main

import (
	"C"
	"fmt"
	"net/http"
	"runtime"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
        msg := "Hello world from Golang"
	w.Write([]byte(msg))
}

func main() {
}

//export GoMain
func GoMain() {
	fmt.Println("Go version:", runtime.Version())
	fmt.Println("Server listening on port 8014")

	http.HandleFunc("/", rootHandler)
	http.ListenAndServe(":8014", nil)
}
