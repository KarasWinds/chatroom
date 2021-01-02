package main

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/KarasWinds/chatroom/logic"
	"nhooyr.io/websocket/wsjson"

	"flag"
	"time"

	"nhooyr.io/websocket"
)

var (
	userNum       int           // 使用者數
	loginInterval time.Duration // 使用者登入時間間隔
	msgInterval   time.Duration // 同一使用者發送訊息間隔
)

func init() {
	flag.IntVar(&userNum, "u", 500, "登入使用者數")
	flag.DurationVar(&loginInterval, "l", 5e9, "使用者陸續登入時間間隔")
	flag.DurationVar(&msgInterval, "m", 1*time.Minute, "使用者發送訊息時間間隔")
}

func main() {
	flag.Parse()

	for i := 0; i < userNum; i++ {
		go UserConnect("user" + strconv.Itoa(i))
		time.Sleep(loginInterval)
	}

	select {}
}

func UserConnect(nickname string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, "ws://127.0.0.1:2022/ws?nickname="+nickname, nil)
	if err != nil {
		log.Println("Dial error:", err)
		return
	}
	defer conn.Close(websocket.StatusInternalError, "內部錯誤!")

	go sendMessage(conn, nickname)

	ctx = context.Background()

	for {
		var message logic.Message
		err = wsjson.Read(ctx, conn, &message)
		if err != nil {
			log.Println("receive msg error:", err)
			continue
		}
		if message.ClientSendTime.IsZero() {
			continue
		}
		if d := time.Now().Sub(message.ClientSendTime); d > 1*time.Second {
			fmt.Printf("接收到服務端回應(%d) : %#v\n", d.Milliseconds(), message)
		}
	}
	conn.Close(websocket.StatusNormalClosure, "")
}

func sendMessage(conn *websocket.Conn, nickname string) {
	ctx := context.Background()
	i := 1
	for {
		msg := map[string]string{
			"content":   "來自" + nickname + "的消息" + strconv.Itoa(i),
			"send_time": strconv.FormatInt(time.Now().UnixNano(), 10),
		}
		err := wsjson.Write(ctx, conn, msg)
		if err != nil {
			log.Println("send msg error:", err, "nickname:", nickname, "no", i)
		}
		i++

		time.Sleep(msgInterval)
	}
}
