package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"io"
	"fmt"
	"time"
	"html/template"
	"log"
	"github.com/gorilla/websocket"
)
type config struct {
	VirtualHost string
	Port int
	HttpReadTimeOut time.Duration
	HttpWriteTimeOut time.Duration
	HttpIdleTimeout time.Duration
}

func hardcodedConfig() *config {
	return &config {
		VirtualHost:"",
		Port: 8000,
		HttpReadTimeOut: 30 * time.Second,
		HttpWriteTimeOut: 30 * time.Second,
		HttpIdleTimeout: 5 * time.Second,
	}
}

func main() {
	config := hardcodedConfig()

	router := &mux.Router{}
	router.NotFoundHandler = func() http.HandlerFunc {
		return func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusNotFound)
			resp.Header().Set("Content-Type", "text/html")
			io.WriteString(resp, "Because some body could we wrong.")
		}
	}()

	router.HandleFunc("/", indexAction).Methods("GET")
	router.HandleFunc("/channel/", wsAction).Methods("GET")
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/")))).Methods("GET")

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.VirtualHost, config.Port),
		Handler:      router,
		ReadTimeout:  config.HttpReadTimeOut,
		WriteTimeout: config.HttpWriteTimeOut,
		IdleTimeout:  config.HttpIdleTimeout,
	}

	server.ListenAndServe()
}

func indexAction(resp http.ResponseWriter, req *http.Request) {
	index, err := template.ParseFiles("templates/index.tpl.html")
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Header().Set("Content-Type", "text/html")
		io.WriteString(resp, err.Error())
		return
	}
	resp.WriteHeader(http.StatusOK)
	resp.Header().Set("Content-Type", "text/html")

	tplData := struct {}{}

	index.Execute(resp, &tplData)
}

func wsAction(resp http.ResponseWriter, req *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(resp, req, nil)
	if err != nil {
		log.Println(err)
		return
	}
	_ = conn

	log.Println("New Connection")
}
