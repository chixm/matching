package matching

import (
	"errors"
	"sync"
)

// 全体ロック
var mux = sync.Mutex{}

//カウントアップするルーム番号
var roomAutoInc = 0

// 現在作られている部屋
var currentRooms map[RoomID]*Room

// 部屋番号
type RoomID int

// 部屋のイベント
type RoomEvent int

const (
	RoomUserJoined = RoomEvent(1)
	RoomUserLeft   = RoomEvent(2)
)

// 部屋に関する構造体
type Room struct {
	mux   sync.Mutex
	ID    RoomID         `json:"roomId"`
	Users []*User        `json:"users"`
	Event chan RoomEvent // イベント検知チャンネル
}

// 新しい部屋に入る
func (m *Room) JoinUser(user *User) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	user.mux.Lock()
	defer user.mux.Unlock()
	m.Users = append(m.Users, user)
	user.JoinedRoom = m
	return nil
}

// 現在所属してる部屋から出る
func (m *Room) LeaveRoom(user *User) error {
	if m == nil {
		return errors.New(`room does not exists`)
	}
	m.mux.Lock()
	defer m.mux.Unlock()
	// リスト作り直し
	var newList []*User
	for _, ou := range m.Users {
		if ou.ID == user.ID {
			continue
		}
		newList = append(newList, ou)
	}
	// 誰もいない部屋は片づける
	if len(newList) == 0 {
		RemoveRoom(user.JoinedRoom.ID)
		return nil
	}
	m.Users = newList
	return nil
}

// 部屋を作成
func NewRoom(u *User) (*Room, error) {
	mux.Lock()
	defer mux.Unlock()
	newRoom := Room{}
	roomAutoInc++
	newRoom.ID = RoomID(roomAutoInc)
	newRoom.Event = make(chan RoomEvent)
	newRoom.Users = append(newRoom.Users, u) //部屋を作成した人は入る
	// 新しいルームを登録
	currentRooms[newRoom.ID] = &newRoom
	u.JoinedRoom = &newRoom
	return &newRoom, nil
}

// 部屋を削除
func RemoveRoom(roomID RoomID) {
	mux.Lock()
	defer mux.Unlock()
	for _, u := range currentRooms[roomID].Users {
		// 全ユーザ削除
		u.mux.Lock()
		u.JoinedRoom = nil
		u.mux.Unlock()
	}
	delete(currentRooms, roomID)
}

// 現在のルーム情報を取得する
func GetCurrentRooms() (map[RoomID]*Room, error) {
	mux.Lock()
	defer mux.Unlock()
	if currentRooms != nil {
		return currentRooms, nil
	}
	return nil, errors.New(`no room info found`)
}
