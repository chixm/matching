package matching

// 初期化処理
func init() {
	// 現状のデータを保存
	currentRooms = make(map[RoomID]*Room)
	currentUsers = make(map[UserID]*User)
}
