package matching

import (
	"encoding/json"

	"github.com/chixm/matching/logger"
	"github.com/gorilla/websocket"
)

// ユーザからのメッセージを最初に受け取る箇所
func WebsocketTextMessageReceiver(conn *websocket.Conn, msg []byte) {
	var req Message
	// 届いたメッセージにIDを振る
	req.ID = makeUniqueID()
	// エラー時処理
	defer recoverFromPanic(conn, req.ID)

	if err := json.Unmarshal(msg, &req); err != nil {
		panic(err)
	} else {
		var u *User
		// 既にいればそのユーザをいなければ作る
		if existingUser, ok := currentUsers[UserID(req.UserID)]; ok {
			u = existingUser
			if u.conn != conn { // 接続切れの場合、後からつないだconnectionで上書き
				u.conn = conn
			}
		} else {
			u = NewUser(UserID(req.UserID), conn)
		}

		// ユーザのアクションの処理
		switch req.Act {
		case `joinRoom`:
			var joinMsg MessageJoinRoom
			err = json.Unmarshal(msg, &joinMsg)
			if err != nil {
				panic(err)
			}
			room, ok := currentRooms[joinMsg.RoomID]
			if !ok {
				panic(`room does not exists`)
			}
			room.Users = append(room.Users, u)
			u.JoinRoom(room)
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
type Message struct {
	ID     string `json:"messageId"` //サーバ内部で一意のメッセージを見分けるためのID
	Act    string `json:"action"`    // ユーザが行いたい行動内容
	UserID string `json:"userId"`    // ユーザを一意に識別する
}

// ユーザからのルームに参加リクエスト
type MessageJoinRoom struct {
	Message
	RoomID RoomID `json:"roomId"`
}

// サーバが送信する側のベースメッセージ
type sendingMessage struct {
	ID         string      `json:"messageId"`
	Code       int         `json:"code"`   // 成功・失敗等のコード
	Data       interface{} `json:"data"`   // サーバ側からのデータ
	ErrMessage string      `json:"errMsg"` //エラー時のメッセージ
}

// パニック時の汎用レスポンス
func recoverFromPanic(conn *websocket.Conn, messageID string) {
	if r := recover(); r != nil {
		logger.Errorln(r)
		// TODO: Code群の定義
		conn.WriteJSON(sendingMessage{ID: messageID, ErrMessage: `error`, Code: 999})
	}
}
