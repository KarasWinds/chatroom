package logic

import (
	"log"

	"github.com/KarasWinds/chatroom/global"
)

// 廣播器
type broadcaster struct {
	// 所有聊天室使用者
	users map[string]*User

	// 所有channel統一管理，可以避免亂用
	enteringChannel chan *User
	leavingChannel  chan *User
	messageChannel  chan *Message

	// 判斷該暱稱使用者是否可進入聊天室(重複與否?):true(能)、false(不能)
	checkUserChannel      chan string
	checkUserCanInChannel chan bool

	// 獲取使用者列表
	requestUsersChannel chan struct{}
	usersChannel        chan []*User
}

var Broadcaster = &broadcaster{
	users: make(map[string]*User),

	enteringChannel: make(chan *User),
	leavingChannel:  make(chan *User),
	messageChannel:  make(chan *Message, global.MessageQueueLen),

	checkUserChannel:      make(chan string),
	checkUserCanInChannel: make(chan bool),

	requestUsersChannel: make(chan struct{}),
	usersChannel:        make(chan []*User),
}

// Start 啟動廣播器
// 需要在一個新的goroutine中執行，因為它無回傳
func (b *broadcaster) Start() {
	for {
		select {
		case user := <-b.enteringChannel:
			// 新使用者進入
			b.users[user.NickName] = user

			OfflineProcessor.Send(user)
		case user := <-b.leavingChannel:
			// 使用者離開
			delete(b.users, user.NickName)
			// 避免goroutine洩漏
			user.CloseMessageChannel()
		case msg := <-b.messageChannel:
			// 給線上所有使用者發送訊息
			for _, user := range b.users {
				if user.UID == msg.User.UID {
					continue
				}
				user.MessageChannel <- msg
			}
			// OfflineProcessor.Save(msg)
		case nickname := <-b.checkUserChannel:
			if _, ok := b.users[nickname]; ok {
				b.checkUserCanInChannel <- false
			} else {
				b.checkUserCanInChannel <- true
			}
		case <-b.requestUsersChannel:
			userList := make([]*User, 0, len(b.users))
			for _, user := range b.users {
				userList = append(userList, user)
			}

			b.usersChannel <- userList
		}
	}
}

func (b *broadcaster) Broadcast(msg *Message) {
	if len(b.messageChannel) >= global.MessageQueueLen {
		log.Println("broadcast queue 满了")
	}
	b.messageChannel <- msg
}

func (b *broadcaster) UserEntering(user *User) {
	b.enteringChannel <- user
}

func (b *broadcaster) UserLeaving(user *User) {
	b.leavingChannel <- user
}

func (b *broadcaster) Broadcaster(msg *Message) {
	b.messageChannel <- msg
}

func (b *broadcaster) CanEnterRoom(nickname string) bool {
	b.checkUserChannel <- nickname

	return <-b.checkUserCanInChannel
}

func (b *broadcaster) GetUserList() []*User {
	b.requestUsersChannel <- struct{}{}
	return <-b.usersChannel
}
