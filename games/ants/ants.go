package ants

import (
	"github.com/x1m3/elixir/pubsub"
	"github.com/x1m3/elixir/games/command"
	"github.com/ByteArena/box2d"
	"math/rand"
	"time"
	"math"
	"sync"
)

type Game struct {
	worldMutex sync.RWMutex
	world      box2d.B2World
	ants       []*box2d.B2Body
	fpsSimul   float64
	fps        float64
	widthX     float64
	widthY     float64
	nAnts      int
	events     chan pubsub.Event
}

func New(widthX, widthY float64, nAnts int) *Game {
	world := box2d.MakeB2World(box2d.MakeB2Vec2(0, 0))
	world.SetGravity(box2d.MakeB2Vec2(0, 0))

	//
	// Muro superior, a lo bruto!!!
	//Body definition
	def := box2d.MakeB2BodyDef()
	def.Position.Set(widthX/2, 0)
	def.Type = box2d.B2BodyType.B2_staticBody
	def.FixedRotation = true
	def.AllowSleep = false

	// Create body
	body := world.CreateBody(&def)
	body.SetUserData(-1)

	//shape
	shape := box2d.MakeB2PolygonShape()
	shape.SetAsBox(widthX, 1)

	//fixture
	fd := box2d.MakeB2FixtureDef()
	fd.Shape = &shape
	fd.Restitution = 4
	body.CreateFixtureFromDef(&fd)

	// Muro inferior, a lo bruto
	//Body definition
	def = box2d.MakeB2BodyDef()
	def.Position.Set(widthX/2, widthY)
	def.Type = box2d.B2BodyType.B2_staticBody
	def.FixedRotation = true
	def.AllowSleep = false

	// Create body
	body = world.CreateBody(&def)
	body.SetUserData(-2)

	//shape
	shape = box2d.MakeB2PolygonShape()
	shape.SetAsBox(widthX, 1)

	//fixture
	fd = box2d.MakeB2FixtureDef()
	fd.Shape = &shape
	fd.Restitution = 4
	body.CreateFixtureFromDef(&fd)

	// Muro izquierdo, a lo bruto
	//Body definition
	def = box2d.MakeB2BodyDef()
	def.Position.Set(0, widthY/2)
	def.Type = box2d.B2BodyType.B2_staticBody
	def.FixedRotation = true
	def.AllowSleep = false

	// Create body
	body = world.CreateBody(&def)
	body.SetUserData(-3)

	//shape
	shape = box2d.MakeB2PolygonShape()
	shape.SetAsBox(1, widthY)

	//fixture
	fd = box2d.MakeB2FixtureDef()
	fd.Shape = &shape
	fd.Restitution = 4
	body.CreateFixtureFromDef(&fd)

	// Muro derecho, a lo bruto
	//Body definition
	def = box2d.MakeB2BodyDef()
	def.Position.Set(widthX, widthY/2)
	def.Type = box2d.B2BodyType.B2_staticBody
	def.FixedRotation = true
	def.AllowSleep = false

	// Create body
	body = world.CreateBody(&def)
	body.SetUserData(-4)

	//shape
	shape = box2d.MakeB2PolygonShape()
	shape.SetAsBox(1, widthY)

	//fixture
	fd = box2d.MakeB2FixtureDef()
	fd.Shape = &shape
	fd.Restitution = 4
	body.CreateFixtureFromDef(&fd)

	return &Game{
		world:    world,
		fpsSimul: 60,
		fps:      10,
		nAnts:    nAnts,
		widthX:   widthX,
		widthY:   widthY,
		events:   make(chan pubsub.Event, 10000),
	}
}

func (g *Game) Init() {
	g.ants = g.initAnts(g.nAnts, g.widthX, g.widthY)
	go g.runSimulation(1.0/g.fpsSimul, 8, 3)
}

func (g *Game) StartGame() command.Response {
	panic("implement me")
}

func (g *Game) ProcessCommand(c command.Request) command.Response {
	switch c.(type) {

	case *ViewPortRequest:
		v := c.(*ViewPortRequest)
		ants := g.viewPort(v.X, v.Y, v.XX, v.YY)
		response := ViewportResponse{}
		response.Ants = make([]antResponseDTO, 0, len(ants))
		g.worldMutex.RLock()
		for id, ant := range ants {
			pos := ant.GetPosition()
			response.Ants = append(response.Ants,
				antResponseDTO{
					ID: id,
					X:  pos.X,
					Y:  pos.Y,
				/*	Vx: ant.GetLinearVelocity().X,
					Vy: ant.GetLinearVelocity().X,*/
					R:  ant.GetAngle(),
				})
		}
		g.worldMutex.RUnlock()
		return &response

	default:
		return nil
	}
}

func (g *Game) EventListener() <-chan pubsub.Event {
	return g.events
}

func (g *Game) Stop() {
	panic("implement me")
}

func (g *Game) initAnts(number int, maxX float64, maxY float64) []*box2d.B2Body {
	bodies := make([]*box2d.B2Body, 0, number)

	for i := 0; i < number; i++ {
		// Body definition
		def := box2d.MakeB2BodyDef()
		def.Position.Set(maxX*rand.Float64(), maxY*rand.Float64())
		def.Type = box2d.B2BodyType.B2_dynamicBody
		def.FixedRotation = false
		def.AllowSleep = true
		def.LinearVelocity = box2d.MakeB2Vec2(float64(rand.Intn(20)-10), float64(rand.Intn(20)-10))
		def.LinearDamping = 0.01
		def.AngularDamping = 0.3
		def.Angle = rand.Float64() * 2 * math.Pi
		def.AngularVelocity = 0.1

		// Create body
		antBody := g.world.CreateBody(&def)

		antBody.SetUserData(i)

		// Shape
		//shape := box2d.MakeB2CircleShape()
		//shape.M_radius = 1.9
		shape := box2d.MakeB2PolygonShape()
		shape.SetAsBox(2, 2)

		// fixture
		fd := box2d.MakeB2FixtureDef()
		fd.Shape = &shape
		fd.Density = 10.0
		fd.Restitution = 0.5
		fd.Friction = 1

		antBody.CreateFixtureFromDef(&fd)
		bodies = append(bodies, antBody)
	}
	return bodies
}

func (g *Game) runSimulation(timeStep float64, velocityIterations int, positionIterations int) {
	for {
		g.worldMutex.Lock()
		g.world.Step(timeStep, velocityIterations, positionIterations)
		g.worldMutex.Unlock()

		// TODO: Avoid this timer using a ticker
		time.Sleep(time.Duration(timeStep * float64(time.Second)))

	}
}

func (g *Game) viewPort(x, y, xx, yy float64) map[int]*box2d.B2Body {
	box := box2d.MakeB2AABB()
	box.LowerBound = box2d.MakeB2Vec2(x, y)
	box.UpperBound = box2d.MakeB2Vec2(xx, yy)
	ants := make(map[int]*box2d.B2Body, 0)

	callback := func(fixture *box2d.B2Fixture) bool {
		ants[fixture.M_body.GetUserData().(int)] = fixture.M_body
		return true
	}

	g.worldMutex.RLock()
	g.world.QueryAABB(callback, box)
	g.worldMutex.RUnlock()

	return ants
}
