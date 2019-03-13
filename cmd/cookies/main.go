package main

import (
	"fmt"
	"github.com/x1m3/elixir/games/cookies/bots"

	"net/http"
	"time"
	"github.com/gorilla/mux"
	"io"
	"html/template"
	"github.com/gorilla/websocket"
	"log"
	"github.com/x1m3/elixir/games/cookies"

	"github.com/x1m3/elixir/games/cookies/codec/json"
	"github.com/x1m3/elixir/games/cookies/messages"
	_ "net/http/pprof"
)

const (
	updateClientPeriod         = 100 * time.Millisecond
	pixels2Meters              = 10
	gameWidthMeters            = 2000
	gameHeightMeters           = 2000
	NumCookies                 = 100
	virtualHost                = ""
	port                       = 8000
	serverHTTPReadTimeOut      = 30 * time.Second // Maximum time to read the full http request
	serverHTTPWriteTimeout     = 30 * time.Second // Maximum time to write the full http request
	serverHTTPKeepAliveTimeout = 5 * time.Second  // Keep alive timeout. Time to close an idle connection if keep alive is enable
)

var game *cookies.Game

func main() {

	game = cookies.New(gameWidthMeters, gameHeightMeters)

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

	for i := 0; i < 1000; i++ {
		bot := bots.New(game, bots.NewDummyBotAgent(100, 100))
		go func(i int) {
			log.Println("Bot started", i)
			if err := bot.Run(); err != nil {
				log.Println(err)
				bot.Destroy()
				return
			}

		}(i)
	}

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

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

	go manageRemoteView(transport, sessionID, updateClientPeriod)

	log.Println("New Connection")
}

func manageRemoteView(transport *cookies.Transport, sessionID uint64, updatePeriod time.Duration) {

	for {
		time.Sleep(updatePeriod)
		req, err := game.ViewPortRequest(sessionID)
		if err != nil {
			// Do nothing. Probably, game is not in playing state
			continue
		}

		err = transport.Send(req)
		if err != nil {
			log.Printf("Socket broken while writing. Closing connection. Err:<%v>", err)
			transport.Close()
			return
		}
	}
}

func handleWSRequests(transport *cookies.Transport, sessionID uint64) {
	var resp messages.Message
	var errResp error

	for {
		msg, err := transport.Receive()
		if err != nil {
			log.Printf("Closing conection. Err:<%v>", err)
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
