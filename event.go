package matching

// 部屋のイベント
type RoomEvent int

const (
	RoomUserJoined = RoomEvent(1)
	RoomUserLeft   = RoomEvent(2)
	RoomDismiss    = RoomEvent(3)
)

// ルームで発生するイベント
type Event struct {
	ev   RoomEvent
	user *User
	room *Room
}
