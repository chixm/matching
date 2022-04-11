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
	ID    RoomID         `json:"roomId"`
	Users []*User        `json:"users"`
	Event chan RoomEvent // イベント検知チャンネル
}

// 部屋を作成
func NewRoom() (*Room, error) {
	mux.Lock()
	defer mux.Unlock()
	newRoom := Room{}
	roomAutoInc++
	newRoom.ID = RoomID(roomAutoInc)
	newRoom.Event = make(chan RoomEvent)
	// 新しいルームを登録
	currentRooms[newRoom.ID] = &newRoom
	return &newRoom, nil
}

// 部屋を削除
func RemoveRoom(roomID RoomID) {
	mux.Lock()
	defer mux.Unlock()
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
