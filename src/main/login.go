package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/rs/xid"
)

func Login(w http.ResponseWriter, r *http.Request) {

	if r.Method != `POST` {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Request's Content-Type must be application/json
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get Content length of Request body
	length, err := strconv.Atoi(r.Header.Get("Content-Length"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//Read body data to parse json
	body := make([]byte, length)
	length, err = r.Body.Read(body)
	if err != nil && err != io.EOF {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Println(`Login Request Body::` + string(body))

	var loginInfo LoginInfo
	err = json.Unmarshal(body, &loginInfo)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Println(`Try login::` + loginInfo.UserId + `/` + loginInfo.Password)

	// validate LoginInfo
	if loginInfo.validLoginInfo() {
		// add login hash key to use loggedIn info later.
		res, err := json.Marshal(&LoginResult{Result: true, Message: `succeeded to login`, LoginToken: loginInfo.makeGuid()})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(res)
	} else {
		res, err := json.Marshal(&LoginResult{Result: false, Message: `failed to login`})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(res)
	}
}

type LoginInfo struct {
	UserId   string `json:"userId"`
	Password string `json:"password"`
}

// Validate Login User and Password by anyway.
func (m *LoginInfo) validLoginInfo() bool {
	if m.UserId == `` {
		return false
	}
	if m.Password == `` {
		return false
	}

	// Validate UserID and Password here.

	return true
}

// Create Unique Hash to define User
func (m *LoginInfo) makeGuid() string {
	idHash := xid.New()
	loggedInUsers.GuidMap[m.UserId] = idHash.String()
	loggedInUsers.UserIdMap[idHash.String()] = m.UserId
	return idHash.String()
}

type LoginResult struct {
	Result     bool   `json:"result"`
	Message    string `json:"message"`
	LoginToken string `json:"loginToken"`
}

var loggedInUsers *LoggedInUsers = new(LoggedInUsers)

// Keep the Logged in users to define them later
type LoggedInUsers struct {
	GuidMap   map[string]string
	UserIdMap map[string]string
}

func InitializeLoginFunc() {
	loggedInUsers.GuidMap = make(map[string]string)
	loggedInUsers.UserIdMap = make(map[string]string)
}
