package cookies

import (
	"github.com/x1m3/elixir/pubsub"
	"github.com/ByteArena/box2d"
	"math/rand"
	"time"
	"math"
	"sync"
	"log"

	"github.com/nu7hatch/gouuid"
	"github.com/x1m3/elixir/games/cookies/messages"
)

type Game struct {
	worldMutex sync.RWMutex
	gSessions  *gameSessions

	world      box2d.B2World
	fpsSimul   float64
	fps        float64
	widthX     float64
	widthY     float64
	nAnts      int
	events     chan pubsub.Event
	speed      int
	turboSpeed int
}

type Cookie struct {
	Id        int
	SessionID uuid.UUID
	Score     int
}

// New returns a new cookies game.
func New(widthX, widthY float64, nAnts int) *Game {

	return &Game{
		gSessions:  newGameSessions(),
		world:      box2d.MakeB2World(box2d.MakeB2Vec2(0, 0)),
		fpsSimul:   45,
		fps:        10,
		nAnts:      nAnts,
		widthX:     widthX,
		widthY:     widthY,
		events:     make(chan pubsub.Event, 10000),
		speed:      25,
		turboSpeed: 40,
	}
}

func (g *Game) Init() {
	g.createWorld()
	go g.runSimulation(time.Duration(time.Second/time.Duration(g.fpsSimul)), 8, 2)
}

func (g *Game) NewSession() uuid.UUID {
	return g.gSessions.add()
}

func (g *Game) UserJoin(sessionID uuid.UUID, req *messages.UserJoinRequest) (*messages.UserJoinResponse, error) {

	err := g.gSessions.Login(sessionID, req.Username)
	if err != nil {
		return nil, err
	}

	// TODO: Change type name and inform if login was succesful
	return messages.NewUserJoinResponse(true, nil), nil
}

func (g *Game) CreateCookie(sessionID uuid.UUID, req *messages.CreateCookieRequest) (*messages.CreateCookieResponse, error) {

	if err := g.gSessions.StartPlaying(sessionID); err != nil {
		return nil, err
	}

	c := &Cookie{Id: rand.Int(), SessionID: sessionID, Score: 100}
	x := float64(100 + rand.Intn(int(g.widthX-100)))
	y := float64(100 + rand.Intn(int(g.widthY-100)))
	g.addCookieToWorld(x, y, c)
	return messages.NewCreateCookieResponse(c.Id, c.Score, float32(x), float32(y), 10), nil
}

func (g *Game) ViewPortRequest(sessionID uuid.UUID) (*messages.ViewportResponse, error) {

	v, err := g.gSessions.viewPortRequest(sessionID)
	if err != nil {
		return nil, err
	}

	cookies := g.viewPort(v)

	response := messages.ViewportResponse{}
	response.Type = messages.ViewPortResponseType

	response.Cookies = make([]*messages.CookieInfo, 0, len(cookies))
	g.worldMutex.RLock()
	for _, ant := range cookies {
		pos := ant.GetPosition()
		response.Cookies = append(
			response.Cookies,
			&messages.CookieInfo{
				ID:              ant.GetUserData().(*Cookie).Id,
				Score:           ant.GetUserData().(*Cookie).Score,
				X:               float32(pos.X),
				Y:               float32(pos.Y),
				AngularVelocity: float32(ant.GetAngularVelocity()),
			})
	}
	g.worldMutex.RUnlock()
	return &response, nil
}

func (g *Game) UpdateViewPortRequest(sessionID uuid.UUID, req *messages.ViewPortRequest) {
	g.gSessions.UpdateViewPort(sessionID, req)
}

func (g *Game) EventListener() <-chan pubsub.Event {
	return g.events
}

func (g *Game) createWorld() {

	createWorldBoundary := func(world *box2d.B2World, centerX, centerY, widthX, widthY float64, horiz bool) {
		//Body definition
		def := box2d.MakeB2BodyDef()
		def.Position.Set(centerX, centerY)
		def.Type = box2d.B2BodyType.B2_staticBody
		def.FixedRotation = true
		def.AllowSleep = false

		// Create body
		body := world.CreateBody(&def)
		body.SetUserData(-1)

		//shape
		shape := box2d.MakeB2PolygonShape()

		shape.SetAsBox(widthX, widthY)

		//fixture
		fd := box2d.MakeB2FixtureDef()
		fd.Shape = &shape
		fd.Restitution = 4
		body.CreateFixtureFromDef(&fd)
	}

	g.world.SetGravity(box2d.MakeB2Vec2(0, 0))

	createWorldBoundary(&g.world, g.widthX, 0, g.widthX, 0.1, true)
	createWorldBoundary(&g.world, g.widthX/2, g.widthY, g.widthX, 0.1, true)
	createWorldBoundary(&g.world, 0, g.widthY/2, 0.1, g.widthY, true)
	createWorldBoundary(&g.world, g.widthX, g.widthY/2, 0.1, g.widthY, true)

	g.initCookies(g.nAnts, g.widthX, g.widthY)

}

func (g *Game) addCookieToWorld(x float64, y float64, info *Cookie) *box2d.B2Body {
	// Body definition
	def := box2d.MakeB2BodyDef()
	def.Position.Set(x, y)
	def.Type = box2d.B2BodyType.B2_dynamicBody
	def.FixedRotation = false
	def.AllowSleep = true
	def.LinearVelocity = box2d.MakeB2Vec2(float64(rand.Intn(g.speed)-10), float64(rand.Intn(g.speed)-10))
	def.LinearDamping = 0.0
	def.AngularDamping = 0.0
	def.Angle = rand.Float64() * 2 * math.Pi
	def.AngularVelocity = float64(info.Score / 10)

	// Shape
	shape := box2d.MakeB2PolygonShape()
	shape.SetAsBox(math.Sqrt(float64(info.Score)), math.Sqrt(float64(info.Score)))

	// fixture
	fd := box2d.MakeB2FixtureDef()
	fd.Shape = &shape
	fd.Density = math.Sqrt(float64(info.Score))
	fd.Restitution = 0.5
	fd.Friction = 1

	// Create body
	antBody := g.world.CreateBody(&def)
	antBody.SetUserData(info)
	antBody.CreateFixtureFromDef(&fd)

	return antBody

}

func (g *Game) initCookies(number int, maxX float64, maxY float64) []*box2d.B2Body {
	bodies := make([]*box2d.B2Body, 0, number)

	for i := 0; i < number; i++ {
		uuid, _ := uuid.NewV4()
		cookie := g.addCookieToWorld(maxX*rand.Float64(), maxY*rand.Float64(), &Cookie{Id: i, SessionID: *uuid, Score: rand.Intn(200) + 20})
		bodies = append(bodies, cookie)
	}
	return bodies
}

func (g *Game) runSimulation(timeStep time.Duration, velocityIterations int, positionIterations int) {
	timeStep64 := float64(timeStep) / float64(time.Second)

	allMap := box2d.MakeB2AABB()
	allMap.LowerBound = box2d.MakeB2Vec2(0, 0)
	allMap.UpperBound = box2d.MakeB2Vec2(g.widthX, g.widthY)
	for {
		t1 := time.Now()
		g.worldMutex.Lock()

		g.world.Step(timeStep64, velocityIterations, positionIterations)
		g.adjustSpeeds(&allMap)

		g.worldMutex.Unlock()

		elapsed := time.Since(t1)
		if elapsed < timeStep {
			time.Sleep(timeStep - elapsed)
		} else {
			log.Printf("WARNING: Cannot sustain frame rate. Expected time <%v>. Real time <%v>", timeStep, elapsed)
		}
	}
}

func (g *Game) adjustSpeeds(allMap *box2d.B2AABB) {

	g.world.QueryAABB(
		func(fixture *box2d.B2Fixture) bool {
			body := fixture.M_body
			info := body.GetUserData()

			switch info.(type) {
			case *Cookie:
				// Angular speed
				currentSpeed := body.M_angularVelocity
				expectedSpeed := float64(info.(*Cookie).Score / 10)

				var linearSpeedPenalty float64 = 0

				if currentSpeed <= 0 {
					body.SetAngularVelocity(currentSpeed + (expectedSpeed+math.Abs(currentSpeed))/50)
				} else {
					if currentSpeed < expectedSpeed {
						body.SetAngularVelocity(currentSpeed + (expectedSpeed-currentSpeed)/50)
					} else {
						body.SetAngularVelocity(currentSpeed - (currentSpeed-expectedSpeed)/50)
					}
					linearSpeedPenalty = currentSpeed / expectedSpeed /// Could be positive if it is spinning faster ;-)
				}

				// Linear speed, based on configuration, but also on spining angular speed.
				speedX := body.GetLinearVelocity().X
				speedY := body.GetLinearVelocity().Y
				if math.Abs(speedX) < 1.0 {
					speedX = 1.0
				}
				if math.Abs(speedY) < 1.0 {
					speedY = 1.0
				}
				currentSpeed = math.Sqrt(math.Pow(speedX, 2) + math.Pow(speedY, 2))
				expectedSpeed = linearSpeedPenalty * float64(g.speed)

				offsetX := speedX / 20
				offsetY := speedY / 20

				if currentSpeed < expectedSpeed {
					body.SetLinearVelocity(box2d.MakeB2Vec2(speedX+offsetX, speedY+offsetY))
				} else {
					body.SetLinearVelocity(box2d.MakeB2Vec2(speedX-offsetX, speedY-offsetY))
				}

			}
			return true
		},
		*allMap,
	)

}

func (g *Game) viewPort(v *viewport) map[int]*box2d.B2Body {

	cookies := make(map[int]*box2d.B2Body, 0)

	g.worldMutex.RLock()

	g.world.QueryAABB(
		func(fixture *box2d.B2Fixture) bool {
			info := fixture.M_body.GetUserData()
			switch info.(type) {
			case *Cookie:
				cookies[info.(*Cookie).Id] = fixture.M_body
			}
			return true
		},
		box2d.B2AABB{LowerBound: box2d.MakeB2Vec2(float64(v.x), float64(v.y)), UpperBound: box2d.MakeB2Vec2(float64(v.xx), float64(v.yy))},
		//Return all box2d.B2AABB{LowerBound: box2d.MakeB2Vec2(0, 0), UpperBound: box2d.MakeB2Vec2(g.widthX, g.widthY)},
	)

	g.worldMutex.RUnlock()

	return cookies
}
