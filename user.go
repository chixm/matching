package matching

import (
	"errors"
	"sync"

	"github.com/gorilla/websocket"
)

// ユーザID型エイリアス
type UserID string

// ユーザ情報
type User struct {
	mux        sync.Mutex
	conn       *websocket.Conn // ユーザの接続
	ID         UserID          `json:"id"` //一意にユーザを識別するID
	JoinedRoom *Room           // 参加しているルーム
}

// 部屋に入る
func (m *User) JoinRoom(room *Room) error {
	if room == nil {
		return errors.New(`room not found`)
	}
	m.mux.Lock()
	defer m.mux.Unlock()
	m.JoinedRoom = room
	return nil
}

// 部屋から出る
func (m *User) LeaveRoom() error {
	if m.JoinedRoom == nil {
		return errors.New(`user has not joint room`)
	}
	m.mux.Lock()
	defer m.mux.Unlock()
	m.JoinedRoom = nil
	return nil
}
