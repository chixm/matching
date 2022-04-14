package matching

import (
	"sync"

	"github.com/gorilla/websocket"
)

// 現在ログインしているユーザ
var currentUsers map[UserID]*User

// ユーザID型エイリアス
type UserID string

// ユーザ情報
type User struct {
	mux        sync.Mutex
	conn       *websocket.Conn // ユーザの接続
	ID         UserID          `json:"id"` //一意にユーザを識別するID
	JoinedRoom *Room           // 参加しているルーム
}

// ユーザを作成する 同時に現存ユーザmapに登録
func NewUser(id UserID, conn *websocket.Conn) *User {
	u := User{ID: id, conn: conn}
	mux.Lock()
	defer mux.Unlock()
	currentUsers[id] = &u
	return &u
}
