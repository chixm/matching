package matching

import "github.com/rs/xid"

// ユニークな文字列を生成する機構
// Create Unique Hash to define User
func makeUniqueID() string {
	idHash := xid.New()
	return idHash.String()
}
