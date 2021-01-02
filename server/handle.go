package server

import (
	"net/http"

	"github.com/KarasWinds/chatroom/logic"
)

func RegisterHandle() {
	go logic.Broadcaster.Start()

	http.HandleFunc("/", homeHandleFunc)
	http.HandleFunc("/ws", WebSocketHandleFunc)
}
