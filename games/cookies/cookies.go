package cookies

import (
	"errors"
	"github.com/ByteArena/box2d"
	"github.com/x1m3/elixir/pubsub"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/x1m3/elixir/games/cookies/messages"
	"sync/atomic"
)

type Game struct {
	worldMutex sync.RWMutex
	gSessions  *gameSessions

	world           box2d.B2World
	contactListener box2d.B2ContactListenerInterface
	fpsSimul        float64
	fps             float64
	widthX          float64
	widthY          float64
	nAnts           int
	events          chan pubsub.Event
	col2Cookies     chan *collision2CookiesDTO
	colCookieFood   chan *collissionCookieFoodDTO
	speed           int
	turboSpeed      int
	maxFoodCount    uint64
	foodCount       uint64
	bodys2Destroy   sync.Map // TODO: Refactor with something more performance. or not...
}

// New returns a new cookies game.
func New(widthX, widthY float64, nAnts int) *Game {

	chColl2Cookies := make(chan *collision2CookiesDTO, 1024)
	chCollCookieFood := make(chan *collissionCookieFoodDTO, 1024)

	return &Game{
		gSessions:       newGameSessions(),
		world:           box2d.MakeB2World(box2d.MakeB2Vec2(0, 0)),
		contactListener: newContactListener(chColl2Cookies, chCollCookieFood),
		fpsSimul:        45,
		fps:             10,
		nAnts:           nAnts,
		widthX:          widthX,
		widthY:          widthY,
		events:          make(chan pubsub.Event, 10000),
		col2Cookies:     chColl2Cookies,
		colCookieFood:   chCollCookieFood,
		speed:           45,
		turboSpeed:      70,
		maxFoodCount:    5000,
	}
}

func (g *Game) Init() {
	g.createWorld()
	g.initCollissionListeners()
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

	if !session.inLoggedState() {
		return nil, errors.New("not logged user wants to play")
	}

	x := float64(300 + rand.Intn(int(g.widthX-300)))
	y := float64(300 + rand.Intn(int(g.widthY-300)))

	session.setBox2DBody(g.addCookieToWorld(x, y, session))

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
				ID:              ant.GetUserData().(*Cookie).ID,
				Score:           ant.GetUserData().(*Cookie).Score,
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
}

func (g *Game) addCookieToWorld(x float64, y float64, session *gameSession) *box2d.B2Body {

	var score uint64 = 100

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

	// Create body
	g.worldMutex.Lock()
	body := g.world.CreateBody(&def)

	body.CreateFixtureFromDef(g.getCookieFixtureDefByScore(score))

	// Save link to session
	body.SetUserData(&Cookie{ID: session.ID, Score: session.getScore(), body: body})
	g.worldMutex.Unlock()

	return body
}

func (g *Game) addFoodToWorld(x, y float64, score uint64) {
	def := box2d.MakeB2BodyDef()
	def.Position.Set(x, y)
	def.Type = box2d.B2BodyType.B2_dynamicBody
	def.LinearDamping = 0
	def.FixedRotation = true
	def.AllowSleep = true

	// Shape
	shape := box2d.MakeB2CircleShape()
	shape.M_radius = 1

	// fixture
	fd := box2d.MakeB2FixtureDef()
	fd.Shape = &shape
	fd.Density = 0.000001
	fd.Restitution = 0
	fd.Friction = 0

	// Create body
	g.worldMutex.Lock()
	body := g.world.CreateBody(&def)

	g.worldMutex.Unlock()
	body.CreateFixtureFromDef(&fd)

	// Save link to session
	body.SetUserData(&Food{ID: rand.Uint64() << 8, Score: score, body: body})
}

func (g *Game) initCollissionListeners() {
	g.world.SetContactListener(g.contactListener)
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

	go g.contactBetweenCookies()
	go g.contactBetweenCookiesAndFood()

	for {
		t1 := time.Now()
		g.worldMutex.Lock()

		g.world.Step(timeStep64, velocityIterations, positionIterations)
		g.removeBodies()

		g.adjustSpeedsAndSizes()

		g.worldMutex.Unlock()

		elapsed := time.Since(t1)
		if elapsed < timeStep {
			time.Sleep(timeStep - elapsed)
		} else {
			log.Printf("WARNING: Cannot sustain frame rate. Expected time <%v>. Real time <%v>", timeStep, elapsed)
		}
	}
}

func (g *Game) removeBodies() {
	g.bodys2Destroy.Range(func(id interface{}, body interface{}) bool {
		g.world.DestroyBody(body.(*box2d.B2Body))
		g.bodys2Destroy.Delete(id)
		return true
	})
}

func (g *Game) adjustSpeedsAndSizes() {

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

		// size and score
		data := body.GetUserData().(*Cookie)
		if session.getScore() != data.Score {
			data.Score = session.score

			body.DestroyFixture(body.GetFixtureList())
			body.CreateFixtureFromDef(g.getCookieFixtureDefByScore(data.Score))
		}
		return true
	})
}

func (g *Game) getCookieFixtureDefByScore(score uint64) *box2d.B2FixtureDef {

	// Shape
	shape := box2d.MakeB2CircleShape()
	sc := float64(score)
	shape.M_radius = (math.Log2(sc) + math.Sqrt(sc)) / 2

	// fixture
	fd := box2d.MakeB2FixtureDef()
	fd.Shape = &shape
	fd.Density = 100 * math.Sqrt(sc)
	fd.Restitution = 0.7
	fd.Friction = 0.1
	return &fd
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
			case *Cookie:
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

func (g *Game) contactBetweenCookies() {
	for {
		collision := <-g.col2Cookies

		cookie1 := collision.cookie1
		cookie2 := collision.cookie2
		_ = cookie1
		_ = cookie2
	}
}

func (g *Game) contactBetweenCookiesAndFood() {
	for {
		collision := <-g.colCookieFood

		cookie := collision.cookie
		food := collision.food

		g.gSessions.session(cookie.ID).score += food.Score

		atomic.AddUint64(&g.foodCount, ^uint64(0)) // Decrement 1 :-)

		// TODO: Remove food
		g.bodys2Destroy.Store(food.ID, food.body)
	}
}

func (g *Game) markBodyToBeDestroyed(body *box2d.B2Body) {

}
