package cookies

import (
	"errors"
	"github.com/x1m3/elixir/games/cookies/messages"
	"math/rand"
)

type Game struct {
	gSessions *gameSessions
	world     *world
	width     float64
	height    float64
}

// New returns a new cookies game.
func New(widthX, widthY float64) *Game {

	gameSessions := newGameSessions()

	return &Game{
		gSessions: gameSessions,
		world:     NewWorld(gameSessions, widthX, widthY, 10, 60, 45, 70, 5000),
		width:     widthX,
		height:    widthY,
	}
}

func (g *Game) Init() {
	g.world.createWorld()
	go g.world.runSimulation( 4, 1)
}

func (g *Game) NewSession() uint64 {
	return g.gSessions.add()
}

func (g *Game) UserJoin(sessionID uint64, req *messages.UserJoinRequest) (*messages.UserJoinResponse, error) {

	if err := g.gSessions.session(sessionID).login(req.Username); err != nil {
		return nil, err
	}

	// TODO: Change type name and inform if login was ok
	return messages.NewUserJoinResponse(true, nil), nil
}

func (g *Game) CreateCookie(sessionID uint64, req *messages.CreateCookieRequest) (*messages.CreateCookieResponse, error) {

	session := g.gSessions.session(sessionID)

	if !session.inLoggedState() {
		return nil, errors.New("not logged user wants to play")
	}

	x := float64(300 + rand.Intn(int(g.width-300)))
	y := float64(300 + rand.Intn(int(g.height-300)))

	session.setBox2DBody(g.world.addCookieToWorld(x, y, session))

	if err := session.startPlaying(); err != nil {
		return nil, err
	}
	return messages.NewCreateCookieResponse(sessionID, session.getScore(), float32(x), float32(y), 10), nil
}

func (g *Game) ViewPortRequest(sessionID uint64) (*messages.ViewportResponse, error) {

	v, err := g.gSessions.session(sessionID).getViewport()
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
				ID:              cookie.GetUserData().(*Cookie).ID,
				Score:           cookie.GetUserData().(*Cookie).Score,
				X:               float32(pos.X),
				Y:               float32(pos.Y),
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
	g.gSessions.session(sessionID).updateViewPort(req.X, req.Y, req.XX, req.YY, req.Angle, req.Turbo)
}
