package cookies

import (
	"github.com/ByteArena/box2d"
	"github.com/x1m3/elixir/pubsub"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/x1m3/elixir/games/cookies/messages"
	"sync/atomic"
)

type Game struct {
	worldMutex sync.RWMutex
	gSessions  *gameSessions

	world        box2d.B2World
	fpsSimul     float64
	fps          float64
	widthX       float64
	widthY       float64
	nAnts        int
	events       chan pubsub.Event
	speed        int
	turboSpeed   int
	maxFoodCount uint64
	foodCount    uint64
}

type Cookie struct {
	ID    int
	Score int
}

type Food struct {
	ID    uint64
	Score int
}

// New returns a new cookies game.
func New(widthX, widthY float64, nAnts int) *Game {

	return &Game{
		gSessions:    newGameSessions(),
		world:        box2d.MakeB2World(box2d.MakeB2Vec2(0, 0)),
		fpsSimul:     45,
		fps:          10,
		nAnts:        nAnts,
		widthX:       widthX,
		widthY:       widthY,
		events:       make(chan pubsub.Event, 10000),
		speed:        40,
		turboSpeed:   65,
		maxFoodCount: 5000,
	}
}

func (g *Game) Init() {
	g.createWorld()
	go g.runSimulation(time.Duration(time.Second/time.Duration(g.fpsSimul)), 8, 2)
}

func (g *Game) NewSession() uint64 {
	return g.gSessions.add()
}

func (g *Game) UserJoin(sessionID uint64, req *messages.UserJoinRequest) (*messages.UserJoinResponse, error) {

	if err := g.gSessions.session(sessionID).login(req.Username); err != nil {
		return nil, err
	}

	// TODO: Change type name and inform if login was succesful
	return messages.NewUserJoinResponse(true, nil), nil
}

func (g *Game) CreateCookie(sessionID uint64, req *messages.CreateCookieRequest) (*messages.CreateCookieResponse, error) {

	session := g.gSessions.session(sessionID)

	if err := session.startPlaying(); err != nil {
		return nil, err
	}

	x := float64(300 + rand.Intn(int(g.widthX-300)))
	y := float64(300 + rand.Intn(int(g.widthY-300)))

	log.Printf("New cookie at position <%f, %f>\n", x, y)
	session.setBox2DBody(g.addCookieToWorld(x, y, session))
	return messages.NewCreateCookieResponse(sessionID, session.getScore(), float32(x), float32(y), 10), nil
}

func (g *Game) ViewPortRequest(sessionID uint64) (*messages.ViewportResponse, error) {

	v, err := g.gSessions.session(sessionID).getViewport()
	if err != nil {
		return nil, err
	}

	cookies, food := g.viewPort(v)

	response := messages.ViewportResponse{}
	response.Type = messages.ViewPortResponseType

	response.Cookies = make([]*messages.CookieInfo, 0, len(cookies))
	response.Food = make([]*messages.FoodInfo, 0, len(food))

	g.worldMutex.RLock()
	for _, ant := range cookies {
		pos := ant.GetPosition()
		response.Cookies = append(
			response.Cookies,
			&messages.CookieInfo{
				ID:              ant.GetUserData().(*gameSession).ID,
				Score:           ant.GetUserData().(*gameSession).getScore(),
				X:               float32(pos.X),
				Y:               float32(pos.Y),
				AngularVelocity: float32(ant.GetAngularVelocity()),
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
	g.worldMutex.RUnlock()

	return &response, nil
}

func (g *Game) UpdateViewPortRequest(sessionID uint64, req *messages.ViewPortRequest) {
	g.gSessions.session(sessionID).updateViewPort(req.X, req.Y, req.XX, req.YY, req.Angle, req.Turbo)
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

func (g *Game) addCookieToWorld(x float64, y float64, session *gameSession) *box2d.B2Body {

	score := 100

	// Body definition
	def := box2d.MakeB2BodyDef()
	def.Position.Set(x, y)
	def.Type = box2d.B2BodyType.B2_dynamicBody
	def.FixedRotation = false
	def.AllowSleep = true
	def.LinearVelocity = box2d.MakeB2Vec2(float64(rand.Intn(g.speed)-10), float64(rand.Intn(g.speed)-10))
	def.LinearDamping = 1.0
	def.AngularDamping = 0.0
	def.Angle = rand.Float64() * 2 * math.Pi
	def.AngularVelocity = 10

	// Shape
	shape := box2d.MakeB2PolygonShape()
	shape.SetAsBox(math.Sqrt(float64(score)), math.Sqrt(float64(score)))

	// fixture
	fd := box2d.MakeB2FixtureDef()
	fd.Shape = &shape
	fd.Density = 10 * math.Sqrt(float64(score))
	fd.Restitution = 0.7
	fd.Friction = 0.1

	// Create body
	g.worldMutex.Lock()
	antBody := g.world.CreateBody(&def)
	g.worldMutex.Unlock()

	antBody.CreateFixtureFromDef(&fd)

	// Save link to session
	antBody.SetUserData(session)

	return antBody
}

func (g *Game) addFoodToWorld(x, y float64, score int) {
	def := box2d.MakeB2BodyDef()
	def.Position.Set(x, y)
	def.Type = box2d.B2BodyType.B2_staticBody
	def.FixedRotation = true
	def.AllowSleep = true

	// Shape
	shape := box2d.MakeB2CircleShape()
	shape.M_radius = 1

	// fixture
	fd := box2d.MakeB2FixtureDef()
	fd.Shape = &shape

	// Create body
	g.worldMutex.Lock()
	antBody := g.world.CreateBody(&def)
	g.worldMutex.Unlock()
	antBody.CreateFixtureFromDef(&fd)

	// Save link to session
	antBody.SetUserData(&Food{ID: rand.Uint64() << 8, Score: score})
}

func (g *Game) initCookies(number int, maxX float64, maxY float64) {

	for i := 0; i < number; i++ {
		sessionID := g.NewSession()
		_, err := g.UserJoin(sessionID, &messages.UserJoinRequest{Username: "manolo"})
		spew.Dump(err)
		cookie := g.addCookieToWorld(maxX*rand.Float64(), maxY*rand.Float64(), g.gSessions.session(sessionID))
		// TODO: Refactor.. use a function!!!!
		g.gSessions.session(sessionID).setBox2DBody(cookie)
		g.gSessions.session(sessionID).state = &playingState{}
	}
}

func (g *Game) runSimulation(timeStep time.Duration, velocityIterations int, positionIterations int) {
	timeStep64 := float64(timeStep) / float64(time.Second)

	foodTicker := time.NewTicker(2 * time.Second)
	go func() {
		for {
			<-foodTicker.C
			g.adjustFood()
		}
	}()

	/*
	allMap := box2d.MakeB2AABB()
	allMap.LowerBound = box2d.MakeB2Vec2(0, 0)
	allMap.UpperBound = box2d.MakeB2Vec2(g.widthX, g.widthY)
	*/
	for {
		t1 := time.Now()
		g.worldMutex.Lock()

		g.world.Step(timeStep64, velocityIterations, positionIterations)
		g.adjustSpeeds()

		g.worldMutex.Unlock()

		elapsed := time.Since(t1)
		if elapsed < timeStep {
			time.Sleep(timeStep - elapsed)
		} else {
			log.Printf("WARNING: Cannot sustain frame rate. Expected time <%v>. Real time <%v>", timeStep, elapsed)
		}
	}
}

func (g *Game) adjustSpeeds() {

	g.gSessions.each(func(session *gameSession) bool {

		if _, ok := session.state.(*playingState); !ok {
			return true
		}
		body := session.box2dbody
		inertia := body.GetInertia()

		// Angular speed
		currentSpeed := body.M_angularVelocity
		expectedSpeed := float64(session.getScore() / 10)
		body.ApplyTorque(inertia*(expectedSpeed-currentSpeed)/2, true)

		if session.viewport == nil {
			return true
		}

		// Linear speed, based on configuration, but also on spinning angular speed.
		speedX := body.GetLinearVelocity().X
		speedY := body.GetLinearVelocity().Y
		currentSpeed = math.Sqrt(math.Pow(speedX, 2) + math.Pow(speedY, 2))
		expectedSpeed = float64(g.speed)

		// Consider adding a moving average here to smooth movements
		magnitude := 1 * (expectedSpeed - currentSpeed) * inertia

		vector := box2d.MakeB2Vec2(math.Cos(float64(session.viewport.angle)), math.Sin(float64(session.viewport.angle)))
		if magnitude < 0 {
			magnitude *= 0.005
		}
		vector.OperatorScalarMulInplace(magnitude)
		body.ApplyForce(vector, body.GetPosition(), true)

		return true
	})
}

func (g *Game) adjustFood() {
	const N = 100

	foodCount := atomic.LoadUint64(&g.foodCount)

	if foodCount < g.maxFoodCount {
		for i := 0; i < N; i++ {
			g.addFoodToWorld(float64(300+rand.Intn(int(g.widthX-300))), float64(300+rand.Intn(int(g.widthX-300))), 5)
		}
	}
	atomic.AddUint64(&g.foodCount, N)
}

func (g *Game) viewPort(v *viewport) ([]*box2d.B2Body, []*box2d.B2Body) {

	cookies := make([]*box2d.B2Body, 0)
	food := make([]*box2d.B2Body, 0)

	g.worldMutex.RLock()

	g.world.QueryAABB(
		func(fixture *box2d.B2Fixture) bool {
			info := fixture.M_body.GetUserData()
			switch info.(type) {
			case *gameSession:
				cookies = append(cookies, fixture.M_body)
			case *Food:
				food = append(food, fixture.M_body)
			}
			return true
		},
		box2d.B2AABB{LowerBound: box2d.MakeB2Vec2(float64(v.x), float64(v.y)), UpperBound: box2d.MakeB2Vec2(float64(v.xx), float64(v.yy))},
		//Return all box2d.B2AABB{LowerBound: box2d.MakeB2Vec2(0, 0), UpperBound: box2d.MakeB2Vec2(g.widthX, g.widthY)},
	)

	g.worldMutex.RUnlock()

	return cookies, food
}
