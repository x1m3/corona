package cookies

import (
	"errors"
	"fmt"
	"github.com/ByteArena/box2d"
	"github.com/x1m3/elixir/games/cookies/messages"
	"github.com/x1m3/elixir/games/cookies/sessionmanager"
	"log"
	"math/rand"
)

type Game struct {
	gSessions *sessionmanager.Sessions
	world     *world
	width     float64
	height    float64
}

// New returns a new cookies game.
func New(widthX, widthY float64) *Game {

	gameSessions := sessionmanager.New()

	return &Game{
		gSessions: gameSessions,
		world:     NewWorld(gameSessions, widthX, widthY, 10, 45, 45, 70, 10000),
		width:     widthX,
		height:    widthY,
	}
}

func (g *Game) Init() {
	g.world.createWorld()
	go g.world.runSimulation(4, 1)
}

func (g *Game) NewSession() uint64 {
	return g.gSessions.Add()
}

func (g *Game) UserJoin(sessionID uint64, req *messages.UserJoinRequest) (*messages.UserJoinResponse, error) {

	if err := g.gSessions.Login(sessionID, req.Username); err != nil {
		return nil, err
	}

	return messages.NewUserJoinResponse(true, nil), nil
}

func (g *Game) Logout(sessionID uint64) {
	var body *box2d.B2Body
	var err error

	if body, err = g.gSessions.GetCookieBody(sessionID); err != nil {
		log.Printf("Error on Logout. <%s>", err)
		return
	}

	if body != nil {
		g.world.removeCookie(body)
	}
	if err := g.gSessions.Close(sessionID); err != nil {
		log.Printf("Error on Logout. <%s>", err)
		return
	}
}

func (g *Game) CreateCookie(sessionID uint64, req *messages.CreateCookieRequest) (*messages.CreateCookieResponse, error) {

	isLogged, err := g.gSessions.IsLogged(sessionID)
	if err != nil {
		log.Printf("Error with inconsistent session state. <%s>", err)
		return nil, err
	}

	if !isLogged {
		return nil, errors.New("not logged user wants to play")
	}

	x := float64(300 + rand.Intn(int(g.width-300)))
	y := float64(300 + rand.Intn(int(g.height-300)))

	score, err := g.gSessions.GetScore(sessionID)
	if err != nil {
		log.Printf("Error getting session score. <%s>", err)
		return nil, err
	}

	err = g.gSessions.SetCookieBody(sessionID, g.world.addCookieToWorld(x, y, sessionID, score))
	if err != nil {
		log.Printf("Error adding cookie to session, <%s>", err)
	}

	if err := g.gSessions.StartPlaying(sessionID); err != nil {
		return nil, err
	}
	return messages.NewCreateCookieResponse(sessionID, score, float32(x), float32(y)), nil
}

func (g *Game) ViewPortRequest(sessionID uint64) (*messages.ViewportResponse, error) {

	v, err := g.gSessions.GetViewport(sessionID)
	if err != nil {
		return nil, err
	}

	cookies, food := g.world.viewPort(v)

	response := messages.ViewportResponse{}
	response.Type = messages.ViewPortResponseType

	response.Cookies = make([]*messages.CookieInfo, 0, len(cookies))
	response.Food = make([]*messages.FoodInfo, 0, len(food))

	for _, cookie := range cookies {
		pos := cookie.GetPosition()
		response.Cookies = append(
			response.Cookies,
			&messages.CookieInfo{
				ID:    cookie.GetUserData().(*Cookie).ID,
				Score: cookie.GetUserData().(*Cookie).Score,
				X:     float32(pos.X),
				Y:     float32(pos.Y),
			})
	}
	for _, f := range food {
		pos := f.GetPosition()
		response.Food = append(
			response.Food,
			&messages.FoodInfo{
				ID:    f.GetUserData().(*Food).ID,
				Score: f.GetUserData().(*Food).Score,
				X:     float32(pos.X),
				Y:     float32(pos.Y),
			})
	}

	return &response, nil
}

func (g *Game) UpdateViewPortRequest(sessionID uint64, req *messages.ViewPortRequest) {
	err := g.gSessions.SetViewport(sessionID, req.X, req.Y, req.XX, req.YY, req.Angle, req.Turbo)
	if err != nil {
		fmt.Printf("Error updating viewport <%s>", err)
	}
}
