package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/x1m3/corona/internal/bots"
	"github.com/x1m3/corona/internal/codec/json"
	"github.com/x1m3/corona/internal/corona"
	"github.com/x1m3/corona/internal/messages"
)

const (
	updateClientPeriod         = 100 * time.Millisecond
	pixels2Meters              = 10
	gameWidthMeters            = 2000
	gameHeightMeters           = 2000
	virtualHost                = ""
	port                       = 8000
	serverHTTPReadTimeOut      = 10 * time.Second // Maximum time to read the full http request
	serverHTTPWriteTimeout     = 10 * time.Second // Maximum time to write the full http request
	serverHTTPKeepAliveTimeout = 5 * time.Second  // Keep alive timeout. Time to close an idle connection if keep alive is enable
)

var game *corona.Game

func main() {

	game = corona.New(gameWidthMeters, gameHeightMeters, updateClientPeriod)

	router := &mux.Router{}
	router.NotFoundHandler = func() http.HandlerFunc {
		return func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusNotFound)
			resp.Header().Set("Content-Type", "text/html")
			_, _ = io.WriteString(resp, "Not Found.")
		}
	}()
	router.HandleFunc("/", indexAction).Methods("GET")
	router.HandleFunc("/ws/", wsAction).Methods("GET")
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/")))).Methods("GET")

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", virtualHost, port),
		Handler:      router,
		ReadTimeout:  serverHTTPReadTimeOut,
		WriteTimeout: serverHTTPWriteTimeout,
		IdleTimeout:  serverHTTPKeepAliveTimeout,
	}

	go game.Init()
	log.Println("Starting Server")

	botsManager := bots.NewManager(game)
	go botsManager.Init()

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	server.ListenAndServe()
}

func indexAction(resp http.ResponseWriter, req *http.Request) {

	home := "templates/index.tpl.html"
	index, err := template.ParseFiles(home)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		cwd, _ := os.Getwd()
		log.Printf("Error loading index. <working_dir:%s> <Error:%s", cwd, err.Error())
		resp.Header().Set("Content-Type", "text/html")
		_, _ = resp.Write([]byte("Error loading url"))
		return
	}
	resp.WriteHeader(http.StatusOK)
	resp.Header().Set("Content-Type", "text/html")

	tplData := struct {
		UpdateClientPeriod float64
		PixelsToMeters     int
		GameWidth          int
		GameHeight         int
	}{
		UpdateClientPeriod: float64(updateClientPeriod) / float64(time.Second),
		PixelsToMeters:     pixels2Meters,
		GameWidth:          gameWidthMeters,
		GameHeight:         gameHeightMeters,
	}

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

	sessionID, responses, endOfGame := game.NewSession()

	transport := corona.NewTransport(json.Codec, corona.NewWebsocketConnection(conn))

	go handleWSRequests(transport, sessionID)

	go manageRemoteView(transport, sessionID, responses, endOfGame)

	log.Println("New Connection")
}

func manageRemoteView(transport *corona.Transport, sessionID uint64, responses chan interface{}, endOfGame chan interface{}) {
	for {
		select {
		case req, ok := <-responses:
			if !ok {
				return
			}
			if err := transport.Send(req.(messages.Message)); err != nil {
				log.Printf("Socket broken while writing. Closing connection. Err:<%v>", err)
				game.Logout(sessionID)
				transport.Close()
				return
			}
		case _, ok := <-endOfGame:
			if !ok {
				return
			}
		}
	}
}

func handleWSRequests(transport *corona.Transport, sessionID uint64) {
	var resp messages.Message
	var errResp error

	for {
		msg, err := transport.Receive()
		if err != nil {
			log.Printf("Closing conection. Err:<%v>", err)
			game.Logout(sessionID)
			transport.Close()
			return
		}

		errResp = nil
		resp = nil

		switch msg.GetType() {
		case messages.ViewPortRequestType:
			game.UpdateViewPortRequest(sessionID, msg.(*messages.ViewPortRequest))

		case messages.UserJoinRequestType: // join user
			resp, errResp = game.UserJoin(sessionID, msg.(*messages.UserJoinRequest))

		case messages.CreateCookieRequestType:
			resp, errResp = game.CreateCookie(sessionID, msg.(*messages.CreateCookieRequest))

		default:
			log.Printf("got unknown message type <%v>", msg)
		}

		if errResp != nil {
			log.Printf("Error: <%s>", err)
			continue
		}
		if resp != nil {
			if err := transport.Send(resp); err != nil {
				log.Printf("Error sending response, Err:<%v>", err)
			}
		}
	}
}
