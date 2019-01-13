package main

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

var match = &MatchingWaitRoom{waitingUsers: make(map[string]*UserConnection)}

var roomNoAutoInc = 0 //カウントアップするルーム番号

var roomNoMutex sync.Mutex

//メッセージコード
const (
	MATCHING_SUCCEED = 0
	MATCHING_STILL   = 50
	MATCHING_FAILED  = 100
)

//マッチング機能
type MatchingWaitRoom struct {
	mutex        sync.Mutex
	waitingUsers map[string]*UserConnection
}

type UserConnection struct {
	mutex          sync.Mutex
	UserId         string
	Conn           *websocket.Conn
	MatchingStatus bool
	sig            chan int // クローズフラグ
}

type MatchingMessage struct {
	code      int
	connectTo string // 接続先URL
	roomNo    int
	ExtraMsg  string // 追加情報メッセージ
}

func webSocketHandler(h websocket.Handler) http.Handler {
	s := websocket.Server{
		Handler:   h,
		Handshake: nil,
	}
	return s
}

//　ユーザからのwebsocket接続の受付
func Matching(ws *websocket.Conn) {
	defer ws.Close()
	defer RecoverFromPanic()

	ws.SetDeadline(time.Now().Add(1 * time.Minute)) //最大1分待つ
	// ユーザID
	userId := ws.Request().URL.Query().Get(`userId`)

	var endSig = make(chan int)
	defer close(endSig)
	userConnection := &UserConnection{UserId: userId, Conn: ws, MatchingStatus: false, sig: make(chan int)}

	match.addUser(userConnection)
	defer match.removeUser(userConnection)

	go waitForMatchingSuccess(endSig, userConnection)

	<-endSig
	log.Println(`End of Matching leaving wait room ` + userId)
}

type MatchingCommand struct {
	CMD int
}

const (
	MATCHER_ADDED = iota
	MATCHER_REMOVED
)

// サーバー起動時にmatchingDetectionコルーチンを立て、そこでユーザ追加や離脱を監視する
func matchingDetection(fin chan int) chan MatchingCommand {
	var cmd = make(chan MatchingCommand)

	go func() {
		defer close(cmd)
		for {
			select {
			case v := <-cmd:
				switch v.CMD {
				case MATCHER_ADDED:
					log.Println(`User Added to matching room`)
					matchingFreeUsers()
				case MATCHER_REMOVED:
					log.Println(`User Removed from matching room`)
				}
			default:
				//log.Println(`waiting for match`)
				time.Sleep(20 * time.Millisecond)
				matchingFreeUsers()
			}
		}

		fin <- 0 //終了したことを通知
	}()
	return cmd
}

//　対戦相手を選択する
func matchingFreeUsers() {
	match.matchUsers()
}

// 毎秒誰か来ないかチェックする
func waitForMatchingSuccess(endSig chan int, user *UserConnection) {
	t := time.NewTicker(3 * time.Second)
	defer t.Stop()

LOOP_FIND:
	for {
		select {
		case <-t.C:
			// まだ待っていることをユーザに通知
			err := sendMatchWaitingMessage(user)
			if err != nil {
				break LOOP_FIND
			}
		case room := <-user.sig:
			log.Println(`Invited to RoomNo` + strconv.Itoa(room) + ` [` + user.UserId + `]`)
			// マッチングスレッドから呼び出された
			if room < 0 {
				log.Println(`Error on Matching Thread`)
				continue
			} else {
				sendMatchMessage(room, user)
				break LOOP_FIND
			}
		}
	}
	endSig <- 0
	log.Println(`End of Matching Finder for [` + user.UserId + `]`)
}

func sendMatchWaitingMessage(user *UserConnection) error {
	msg := &MatchingMessage{code: MATCHING_STILL, ExtraMsg: `Waiting for Match waiting >>> ` + strconv.Itoa(match.waitingUsersCount())}
	err := websocket.JSON.Send(user.Conn, msg)
	if err != nil {
		log.Println(`Error::` + err.Error())
		return err
	}
	return nil
}

func sendMatchMessage(roomNo int, matchedUsers ...*UserConnection) error {
	msg := &MatchingMessage{code: MATCHING_SUCCEED, roomNo: roomNo, connectTo: `http://localhost:8080/cardgame`}

	var succeededUser []*UserConnection
	for _, t := range matchedUsers {
		// ユーザごとの接続先を決定
		msg.connectTo = msg.connectTo + `?roomNo=` + strconv.Itoa(msg.roomNo) + `&userId` + t.UserId
		err := websocket.JSON.Send(t.Conn, msg)
		if err == nil {
			succeededUser = append(succeededUser, t)
		} else {
			log.Println(err)
		}
	}
	// 失敗した旨を送信
	if len(matchedUsers) != len(succeededUser) {
		for _, s := range succeededUser {
			msg.code = MATCHING_FAILED
			websocket.JSON.Send(s.Conn, msg)
		}
		return errors.New(`Failed to send Message to All members`)
	} else { //成功
		//マッチング待ちから抜く
		for _, s := range succeededUser {
			match.removeUser(s)
			s.mutex.Lock()
			s.MatchingStatus = true
			s.mutex.Unlock()
			log.Println(`Sending Match End Signal to [` + s.UserId + `]`)
		}
	}
	return nil
}

func autoIncrementRoomNo() int {
	roomNoMutex.Lock()
	defer roomNoMutex.Unlock()
	roomNoAutoInc++
	return roomNoAutoInc
}

func (m *MatchingWaitRoom) addUser(user *UserConnection) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	match.waitingUsers[user.UserId] = user

	// 追加したことを送信
	matchingChan <- MatchingCommand{CMD: MATCHER_ADDED}
	log.Println(`Added ` + user.UserId + " to waiting Users.[" + strconv.Itoa(len(match.waitingUsers)) + "]")
}

func (m *MatchingWaitRoom) removeUser(user *UserConnection) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(match.waitingUsers, user.UserId)
	// 削除したことを送信
	matchingChan <- MatchingCommand{CMD: MATCHER_REMOVED}
}

// 対戦するユーザを選択する
func (m *MatchingWaitRoom) findUser(user *UserConnection) (*UserConnection, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for id, roomUser := range m.waitingUsers {
		log.Println(id + ` is in Waiting room`)
		if id == user.UserId {
			continue
		}
		return roomUser, nil
	}
	return nil, errors.New(`Matching Failed`)
}

func (m *MatchingWaitRoom) matchUsers() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	// 対戦するペア
	var pair []*UserConnection
	for _, roomUser := range m.waitingUsers {
		if !roomUser.MatchingStatus {
			pair = append(pair, roomUser)
			roomUser.MatchingStatus = true
			if len(pair) == 2 {
				var copyPair = pair[:]
				go notifyWaitingUsers(copyPair)
				pair = pair[:0] //初期化
			}
		}
	}
	pair = pair[:0]
}

func (m *MatchingWaitRoom) waitingUsersCount() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return len(m.waitingUsers)
}

func notifyWaitingUsers(users []*UserConnection) {
	newRoomNo := autoIncrementRoomNo()
	var f string
	for _, user := range users {
		f += `[` + user.UserId + `]`
		user.sig <- newRoomNo
	}
	log.Println(`Match ` + f)
}
