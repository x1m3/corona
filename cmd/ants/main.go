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
	"github.com/x1m3/elixir/games/ants"
)

const (
	UpcateClientPeriod = 100 * time.Millisecond
	PixelsToMeters     = 10
	GameWidth          = 300
	GameHeight         = 300
	NAnts              = 200
)

const virtualHost = ""
const port = 8000

// Maximum time to read the full http request
const SERVER_HTTP_READTIMEOUT = 30 * time.Second

// Maximum time to write the full http request
const SERVER_HTTP_WRITETIMEOUT = 30 * time.Second

// Keep alive timeout. Time to close an idle connection if keep alive is enable
const SERVER_HTTP_IDLETIMEOUT = 5 * time.Second

type GameSession struct {
	sync.RWMutex
	viewportX  float64
	viewportY  float64
	viewportXX float64
	viewportYY float64
}

func (s *GameSession) SetViewport(x, y, xx, yy float64) {
	s.Lock()
	s.viewportX, s.viewportY, s.viewportXX, s.viewportYY = x, y, xx, yy
	s.Unlock()

}
func (s *GameSession) GetViewport() (float64, float64, float64, float64) {
	s.RLock()
	defer s.RUnlock()
	return s.viewportX, s.viewportY, s.viewportXX, s.viewportYY
}

type GameSessions map[*websocket.Conn]*GameSession

var gSessions GameSessions
var game *ants.Game

func main() {

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
		ReadTimeout:  SERVER_HTTP_READTIMEOUT,
		WriteTimeout: SERVER_HTTP_WRITETIMEOUT,
		IdleTimeout:  SERVER_HTTP_IDLETIMEOUT,
	}

	gSessions = make(map[*websocket.Conn]*GameSession)

	game = ants.New(GameWidth, GameHeight, NAnts)

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
		UpdateClientPeriod: float64(UpcateClientPeriod) / float64(time.Second),
		PixelsToMeters:     PixelsToMeters,
		GameWidth:          GameWidth,
		GameHeight:         GameHeight,
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
	session := &GameSession{
		viewportX: 0,
		viewportY: 0,
	}

	gSessions[conn] = session

	go handleWSRequest(conn)
	go sendAnts(conn, UpcateClientPeriod)
	log.Println("New Connection")
}

func sendAnts(conn *websocket.Conn, updatePeriod time.Duration) {
	var request ants.ViewPortRequest
	session := gSessions[conn]
	for {
		// TODO: Avoid this timer implementing a ticker
		time.Sleep(updatePeriod)
		request.X, request.Y, request.XX, request.YY = session.GetViewport()
		ants := game.ProcessCommand(&request).(*ants.ViewportResponse).Ants
		err := conn.WriteJSON(ants)
		if err != nil {
			log.Println("Socket broken while writing. Closing connection")
			conn.Close()
			return
		}
	}
}

func handleWSRequest(conn *websocket.Conn) {
	for {
		m := make(map[string]interface{})

		err := conn.ReadJSON(&m)
		if err != nil {
			log.Printf("Socket broken while reading. Closing connection: <%s>\n", err)
			conn.Close()
			return
		}
		switch m["t"] {
		case "v":
			viewport := m["d"].(map[string]interface{})
			session := gSessions[conn]
			session.SetViewport(viewport["x"].(float64), viewport["y"].(float64), viewport["xx"].(float64), viewport["yy"].(float64))
		default:
			log.Printf("got unknown message type <%v>", m)
		}

	}
}
