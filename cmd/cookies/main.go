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
	"github.com/x1m3/elixir/games/cookies"
	"github.com/nu7hatch/gouuid"

	"github.com/davecgh/go-spew/spew"

	"github.com/x1m3/elixir/games/cookies/codec/json"
	"github.com/x1m3/elixir/games/cookies/messages"
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

func main() {

	game = cookies.New(gameWidthMeters, gameHeightMeters, NumCookies)

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
	log.Println("Starting Server")

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

	transport := cookies.NewTransport(json.Codec, cookies.NewWebsocketConnection(conn))

	go handleWSRequests(transport, sessionID)

	time.Sleep(2 * time.Second)

	go manageRemoteView(transport, sessionID, updateClientPeriod)

	log.Println("New Connection")
}

func manageRemoteView(transport *cookies.Transport, sessionID uuid.UUID, updatePeriod time.Duration) {

	for {
		time.Sleep(updatePeriod)
		req := game.ViewPortRequest(sessionID)

		err := transport.Send(req)
		if err != nil {
			log.Printf("Socket broken while writing. Closing connection. Err:<%v>", err)
			transport.Close()
			return
		}
	}
}

func handleWSRequests(transport *cookies.Transport, sessionID uuid.UUID) {

	for {

		msg, err := transport.Receive()
		if err != nil {
			log.Printf("Closing conection. Err:<%v>", err)
			transport.Close()

			return
		}

		switch msg.GetType() {
		case messages.ViewPortRequestType:
			game.UpdateViewPortRequest(sessionID, msg.(*messages.ViewPortRequest))

		case messages.UserJoinRequestType: // join user
			req := msg.(*messages.UserJoinRequest)
			spew.Dump(req)

			if err := transport.Send(game.UserJoin(sessionID, req)); err != nil {
				log.Printf("UserJoin error: <%s>", err)
			}

		default:
			log.Printf("got unknown message type <%v>", msg)
		}
	}
}
