package matching

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

// ユーザからのメッセージを最初に受け取る箇所
func WebsocketTextMessageReceiver(conn *websocket.Conn, msg []byte) {
	var req message
	// 届いたメッセージにIDを振る
	req.ID = makeUniqueID()

	if err := json.Unmarshal(msg, &req); err != nil {
		panic(err)
	} else {
		var u *User
		// 既にいればそのユーザをいなければ作る
		if existingUser, ok := currentUsers[UserID(req.ID)]; ok {
			u = existingUser
			if u.conn != conn { // 接続切れの場合後からつないだconnectionで上書き
				u.conn = conn
			}
		} else {
			u = NewUser(UserID(req.UserID), conn)
		}

		// ユーザのアクションの処理
		switch req.Act {
		case `joinRoom`:
			var joinMsg messageJoinRoom
			err = json.Unmarshal(msg, &joinMsg)
			if err != nil {
				panic(err)
			}
			u.JoinRoom(currentRooms[joinMsg.RoomID])
		case `leaveRoom`:
			u.LeaveRoom()
		case `createRoom`:
			_, err := NewRoom(u)
			if err != nil {
				panic(err)
			}
		}
	}
}

// ユーザから送信されてくるメッセージの基本形
type message struct {
	ID     string `json:"messageId"` //サーバ内部で一意のメッセージを見分けるためのID
	Act    string `json:"action"`    // ユーザが行いたい行動内容
	UserID string `json:"userId"`    // ユーザを一意に識別する
}

type messageJoinRoom struct {
	message
	RoomID RoomID `json:"roomId"`
}
