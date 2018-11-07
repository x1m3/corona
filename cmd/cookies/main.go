package main

import (
	"fmt"
	"net/http"
	"time"
	"github.com/gorilla/mux"
	"io"
	"html/template"
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"github.com/x1m3/elixir/games/cookies"
	"github.com/nu7hatch/gouuid"
	"encoding/json"
	"github.com/davecgh/go-spew/spew"
	"github.com/x1m3/elixir/games/command"
)

const (
	updateClientPeriod         = 100 * time.Millisecond
	pixels2Meters              = 10
	gameWidthMeters            = 2000
	gameHeightMeters           = 2000
	NumCookies                 = 200
	virtualHost                = ""
	port                       = 8000
	serverHTTPReadTimeOut      = 30 * time.Second // Maximum time to read the full http request
	serverHTTPWriteTimeout     = 30 * time.Second // Maximum time to write the full http request
	serverHTTPKeepAliveTimeout = 5 * time.Second  // Keep alive timeout. Time to close an idle connection if keep alive is enable
)

var game *cookies.Game
var wsSessions map[*websocket.Conn]uuid.UUID
var wsSessionsMutex sync.RWMutex

func main() {

	game = cookies.New(gameWidthMeters, gameHeightMeters, NumCookies)
	wsSessions = make(map[*websocket.Conn]uuid.UUID)

	router := &mux.Router{}
	router.NotFoundHandler = func() http.HandlerFunc {
		return func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusNotFound)
			resp.Header().Set("Content-Type", "text/html")
			io.WriteString(resp, "Not Found.")
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

	sessionID := game.NewSession()
	wsSessionsMutex.Lock()
	wsSessions[conn] = sessionID
	wsSessionsMutex.Unlock()

	go handleWSRequests(conn, sessionID)

	go manageRemoteView(conn, sessionID, updateClientPeriod)

	log.Println("New Connection")
}

func manageRemoteView(conn *websocket.Conn, sessionID uuid.UUID, updatePeriod time.Duration) {

	for {
		time.Sleep(updatePeriod)
		req := game.ViewPortRequest(sessionID)

		err := conn.WriteJSON(req)
		if err != nil {
			log.Println("Socket broken while writing. Closing connection")
			conn.Close()
			return
		}
	}

}

func handleWSRequests(conn *websocket.Conn, sessionID uuid.UUID) {

	req := cookies.Message{}
	for {

		// Reading the message
		t, data, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: <%s>", err)
			conn.Close()
			return
		}
		if t != websocket.TextMessage && t != websocket.BinaryMessage {
			continue
		}

		if err := json.Unmarshal(data, &req); err != nil {
			log.Printf("Error parsing message: <%s>", err)
			conn.Close()
			return
		}

		switch req.Type {
		case "v": // viewportRequest
			viewPortRequest := &cookies.ViewPortRequest{}
			if err := json.Unmarshal(req.Data, viewPortRequest); err != nil {
				if err := conn.WriteJSON(nil); err != nil {
					log.Printf("ViewportRequest bad request <%s>", err)
				}
			}
			game.UpdateViewPortRequest(sessionID, viewPortRequest)

		case "j": // join user
			userDataRequest := &cookies.UserJoinRequest{}
			if err := json.Unmarshal(req.Data, userDataRequest); err != nil {
				if err := conn.WriteJSON(nil); err != nil {
					log.Printf("UserJoin Bad Request <%s>", err)
				}
			}
			spew.Dump(userDataRequest)

			if err := conn.WriteJSON(game.UserJoin(sessionID, userDataRequest)); err != nil {
				log.Printf("UserJoin error: <%s>", err)
			}

		default:
			log.Printf("got unknown message type <%v>", req)
		}
	}
}

