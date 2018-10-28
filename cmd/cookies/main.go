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
	wsSessions[conn]= sessionID
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
	for {
		m := make(map[string]interface{})
		conn.ReadMessage()
		err := conn.ReadJSON(&m)
		if err != nil {
			log.Printf("Socket broken while reading. Closing connection: <%s>\n", err)
			conn.Close()
			return
		}
		// TODO: Verify that sessionID provided by client is the same that we have in memory

		switch m["t"] {
		case "v":
			viewport := m["d"].(map[string]interface{})
			game.UpdateViewPortRequest(sessionID, viewport["x"].(float64), viewport["y"].(float64), viewport["xx"].(float64), viewport["yy"].(float64))
		default:
			log.Printf("got unknown message type <%v>", m)
		}
	}
}
