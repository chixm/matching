package matching

import (
	"errors"
	"fmt"
	"sync"

	"github.com/chixm/matching/logger"
)

// 全体ロック
var mux = sync.Mutex{}

//カウントアップするルーム番号
var roomAutoInc = 0

// 現在作られている部屋
var currentRooms map[RoomID]*Room

// 部屋番号
type RoomID int

// 部屋に関する構造体
type Room struct {
	mux   sync.Mutex
	ID    RoomID     `json:"roomId"`
	Users []*User    `json:"users"`
	Event chan Event // イベント検知チャンネル
}

// 新しい部屋に入る
func (m *Room) JoinUser(user *User) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.Event <- Event{ev: RoomUserJoined, user: user}
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
	m.Event <- Event{ev: RoomUserLeft, user: user}
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
	newRoom.Event = make(chan Event)
	newRoom.Users = append(newRoom.Users, u) //部屋を作成した人は入る
	// 新しいルームを登録
	currentRooms[newRoom.ID] = &newRoom
	u.JoinedRoom = &newRoom
	go detectRoomEvents(&newRoom)
	return &newRoom, nil
}

// 部屋を削除
func RemoveRoom(roomID RoomID) {
	mux.Lock()
	defer mux.Unlock()
	room, ok := currentRooms[roomID]
	if !ok {
		logger.Infoln(fmt.Sprintf(`room %d not found`, roomID))
		return
	}
	for _, u := range room.Users {
		// 全ユーザ削除
		u.mux.Lock()
		u.JoinedRoom = nil
		u.mux.Unlock()
	}
	// 管理goroutineに終了通知
	room.Event <- Event{ev: RoomDismiss}
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

// 部屋で起こったことを各ユーザに知らせる
func detectRoomEvents(r *Room) {
	defer logger.Infoln(`end of room event detection`)
roomEvent:
	for e := range r.Event {
		switch e.ev {
		case RoomUserJoined:
			for _, u := range r.Users {
				// tell all users someone joined the room
				u.conn.WriteJSON(annouce{Message: string(e.user.ID) + ` joined the room`, User: e.user})
			}
		case RoomUserLeft:
			for _, u := range r.Users {
				// tell all users someone left the room
				u.conn.WriteJSON(annouce{Message: string(e.user.ID) + ` left the room`, User: e.user})
			}
		case RoomDismiss: //解散：監視解除
			break roomEvent
		}
	}
}

// ルーム内に拡散させる情報
type annouce struct {
	Message string `json:"message"`
	User    *User  `json:"user"`
}
