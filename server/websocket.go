package server

import (
	"log"
	"net/http"

	"github.com/KarasWinds/chatroom/logic"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func WebSocketHandleFunc(w http.ResponseWriter, req *http.Request) {
	// Accept 從客戶端接受 WebSocket 驗證，並將連接升級到 WebSocket。
	// 如果 Origin 域與主機不同，Accept 將拒絕驗證，除非設置了 InsecureSkipVerify 選項（通過第三個參數 AcceptOptions 設置）。
	// 換句話說，默認情況下，它不允許跨源請求。如果發生錯誤，Accept 將始終寫入適當的響應
	conn, err := websocket.Accept(w, req, nil)
	if err != nil {
		log.Println("websocket accept error:", err)
		return
	}

	// 1. 新使用者進來，建置該使用者的實例
	token := req.FormValue("token")
	nickname := req.FormValue("nickname")
	if l := len(nickname); l < 2 || l > 20 {
		log.Println("nickname illegal: ", nickname)
		wsjson.Write(req.Context(), conn, logic.NewErrorMessage("非法暱稱，暱稱長度:2-20"))
		conn.Close(websocket.StatusUnsupportedData, "nickname illegal!")
		return
	}

	if !logic.Broadcaster.CanEnterRoom(nickname) {
		log.Println("暱稱已存在:", nickname)
		wsjson.Write(req.Context(), conn, logic.NewErrorMessage("該暱稱已存在!"))
		conn.Close(websocket.StatusUnsupportedData, "nickname exists!")
		return
	}

	userHasToken := logic.NewUser(conn, token, nickname, req.RemoteAddr)

	// 2. 開啟給使用者發送訊息的goroution
	go userHasToken.SendMessage(req.Context())

	// 3. 給新使用者發送歡迎資訊
	userHasToken.MessageChannel <- logic.NewWelcomeMessage(userHasToken)

	// 避免 token 洩露
	tmpUser := *userHasToken
	user := &tmpUser
	user.Token = ""

	// 向所有使用者告知新使用者到來
	msg := logic.NewUserEnterMessage(user)
	logic.Broadcaster.Broadcast(msg)

	// 4. 將該使用者加入廣播器的使用者列表中
	logic.Broadcaster.UserEntering(user)
	log.Println("user: ", nickname, "joins chat")

	// 5. 接收使用者訊息
	err = user.ReceiveMessage(req.Context())

	// 6. 使用者離開
	logic.Broadcaster.UserLeaving(user)
	msg = logic.NewUserLeaveMessage(user)
	logic.Broadcaster.Broadcast(msg)
	log.Println("user:", nickname, "leaves chat")

	// 根據讀取時的錯誤執行不同的Close
	if err == nil {
		conn.Close(websocket.StatusNormalClosure, "")
	} else {
		log.Println("read from client error:", err)
		conn.Close(websocket.StatusInternalError, "Read from client error")
	}

}
