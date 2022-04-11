package matching

import (
	"errors"
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

// 新しい部屋に入る
func (m *User) JoinRoom(room *Room) error {
	if room == nil {
		return errors.New(`room not found`)
	}
	m.mux.Lock()
	defer m.mux.Unlock()
	m.JoinedRoom = room
	return nil
}

// 現在所属してる部屋から出る
func (m *User) LeaveRoom() error {
	if m.JoinedRoom == nil {
		return errors.New(`user has not joined room`)
	}
	m.mux.Lock()
	defer m.mux.Unlock()
	m.JoinedRoom = nil
	return nil
}

// ユーザを作成する 同時に現存ユーザmapに登録
func NewUser(id UserID, conn *websocket.Conn) *User {
	u := User{ID: id, conn: conn}
	mux.Lock()
	defer mux.Unlock()
	currentUsers[id] = &u
	return &u
}
