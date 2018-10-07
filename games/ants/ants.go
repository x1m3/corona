package ants

import (
	"github.com/x1m3/elixir/pubsub"
	"github.com/x1m3/elixir/games/command"
	"github.com/ByteArena/box2d"
	"math/rand"
	"time"
	"math"
	"sync"
	"log"
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
	speed      int
	turboSpeed int
}

type Cookie struct {
	Id         int
	PlayerName string
	Score      int
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
		world:      world,
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
	g.ants = g.initCookies(g.nAnts, g.widthX, g.widthY)
	go g.runSimulation(time.Duration(time.Second/time.Duration(g.fpsSimul)), 8, 2)
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
		for _, ant := range ants {
			pos := ant.GetPosition()
			response.Ants = append(response.Ants,
				antResponseDTO{
					ID: ant.GetUserData().(*Cookie).Id,
					SC: int64(ant.GetUserData().(*Cookie).Score),
					X:  pos.X,
					Y:  pos.Y,
					AV: ant.GetAngularVelocity(),
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
		cookie := g.addCookieToWorld(maxX*rand.Float64(), maxY*rand.Float64(), &Cookie{Id: i, PlayerName: "manolo", Score: rand.Intn(200) + 20})
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

		g.adjustSpeeds(&allMap)
		g.world.Step(timeStep64, velocityIterations, positionIterations)

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

func (g *Game) viewPort(x, y, xx, yy float64) map[int]*box2d.B2Body {
	box := box2d.MakeB2AABB()
	box.LowerBound = box2d.MakeB2Vec2(x, y)
	box.UpperBound = box2d.MakeB2Vec2(xx, yy)
	ants := make(map[int]*box2d.B2Body, 0)

	callback := func(fixture *box2d.B2Fixture) bool {
		info := fixture.M_body.GetUserData()
		switch info.(type) {
		case *Cookie:
			ants[info.(*Cookie).Id] = fixture.M_body
		}

		return true
	}

	g.worldMutex.RLock()
	g.world.QueryAABB(callback, box)
	g.worldMutex.RUnlock()

	return ants
}
