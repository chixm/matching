package main

import (
	"log"
	"net/http"
	"strconv"
)

var server *http.Server
var serverMutex *http.ServeMux

// Matching Server with Websocket Client for Online Games
func main() {
	log.Println(`Server Start`)
	var exit = make(chan interface{})
	// サーバー設定
	server, serverMutex = CreateServer(8080)

	//　マッチングスレッド
	runMatchingThread()

	// サーバー起動
	go LaunchServer(exit, server, serverMutex)

	<-exit
}

func LaunchServer(exit chan interface{}, server *http.Server, mutex *http.ServeMux) {
	log.Println(`Launching Server`)
	UrlSettings(mutex)

	err := server.ListenAndServe()
	if err != nil {
		log.Println(err.Error())
		exit <- 1
	}
}

// マッチング読み込みチャネル
var matchingChan chan MatchingCommand

// マッチングスレッド監視
func runMatchingThread() {
	var finishDetection = make(chan int) //マッチングスレッドは終了してはいけない。終了検知時は再起動する
	matchingChan = matchingDetection(finishDetection)
	go func() { //matchingDetectionのゴルーチン監視
		defer close(finishDetection)
		<-finishDetection
		log.Println(`Matching Thread Finished... Restarting Matching Thread.`)
		runMatchingThread()
	}()
}

func UrlSettings(m *http.ServeMux) {
	// Login URI for matching, all other websocket uri requires login token.
	InitializeLoginFunc()
	m.HandleFunc(`/login`, Login)
	m.Handle(`/match`, webSocketHandler(Matching))
}

func CreateServer(port int) (*http.Server, *http.ServeMux) {
	m := &http.ServeMux{}
	s := &http.Server{Addr: `localhost:` + strconv.Itoa(port), Handler: m}
	return s, m
}

func RecoverFromPanic() {
	if r := recover(); r != nil {
		log.Println(r)
	}
}
