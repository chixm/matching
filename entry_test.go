package matching_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chixm/matching"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

// ユーザのルーム作成テスト
func TestUserEntry(t *testing.T) {
	// 入力
	in := matching.Message{
		Act:    `createRoom`,
		UserID: `abc`,
	}
	msg, _ := json.Marshal(in)
	conn := getConnection()

	matching.WebsocketTextMessageReceiver(conn, msg)

	// 一人がインしてるルームがある
	rooms, _ := matching.GetCurrentRooms()

	assert.Equal(t, len(rooms[1].Users), 1)
}

func TestUserJoinOtherUser(t *testing.T) {
	// 入力
	in := matching.Message{
		Act:    `createRoom`,
		UserID: `eric`,
	}
	msg, _ := json.Marshal(in)
	conn := getConnection()

	matching.WebsocketTextMessageReceiver(conn, msg)

	otherUser := matching.MessageJoinRoom{in, matching.RoomID(1)}
	otherUser.Act = `joinRoom`
	otherUser.UserID = `samuel`

	msg, _ = json.Marshal(otherUser)
	conn2 := getConnection()
	matching.WebsocketTextMessageReceiver(conn2, msg)

	// 2人がインしてるルームがある
	rooms, _ := matching.GetCurrentRooms()

	assert.Equal(t, len(rooms[1].Users), 2)
}

func getConnection() *websocket.Conn {
	var up websocket.Upgrader
	w := httptest.NewRecorder()
	r := httptest.NewRequest(`POST`, `http://localhost:8080/ws`, nil)
	h := http.Header{}
	conn, _ := up.Upgrade(w, r, h)
	return conn
}
