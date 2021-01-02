package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/KarasWinds/chatroom/server"
)

var (
	addr   = ":2022"
	banner = `
    ____                _____
   |     |    |    /\     |
   |     |____|   /  \    |
   |     |    |  /----\   |
   |____ |    | /      \  |
Golang-Practice：ChatRoom，start on：%s
`
)

func main() {
	fmt.Printf(banner+"\n", addr)

	server.RegisterHandle()

	log.Fatal(http.ListenAndServe(addr, nil))

}
