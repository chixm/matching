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
	var exit = make(chan interface{})
	server, serverMutex = CreateServer(8080)

	go LaunchServer(exit, server, serverMutex)

	<-exit
}

func LaunchServer(exit chan interface{}, server *http.Server, mutex *http.ServeMux) {

	UrlSettings(mutex)

	err := server.ListenAndServe()
	if err != nil {
		log.Println(err.Error())
		exit <- 1
	}
}

func UrlSettings(m *http.ServeMux) {
	// Login URI for matching, all other websocket uri requires login token.
	InitializeLoginFunc()
	m.HandleFunc(`/login`, Login)

}

func CreateServer(port int) (*http.Server, *http.ServeMux) {
	m := &http.ServeMux{}
	s := &http.Server{Addr: `localhost:` + strconv.Itoa(port), Handler: m}
	return s, m
}
