package cookies

import (
	"fmt"
	"github.com/ByteArena/box2d"
	"github.com/x1m3/elixir/games/cookies/sessionmanager"
	"github.com/x1m3/elixir/pkg/list"
	"log"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type world struct {
	worldMutex sync.RWMutex

	// TODO: Ideally, world shouldn't know nothing about sessions. We should break this dependency with events or other mechanism.
	gSessions *sessionmanager.Sessions

	box2d.B2World
	width  float64
	height float64

	minFPS     float64
	maxFPS     float64
	currentFPS float64

	col2Cookies   chan *collision2CookiesDTO
	colCookieFood chan *collissionCookieFoodDTO

	speed          int
	turboSpeed     int
	minFoodCount   uint64
	foodCount      uint64
	bodies2Destroy sync.Map // TODO: Refactor with something more performance. or not...
	foodQueue      *list.LIFO
}

func NewWorld(gs *sessionmanager.Sessions, w, h, minFPS, maxFPS float64, speed, turboSpeed int, minFoodCount uint64) *world {

	chColl2Cookies := make(chan *collision2CookiesDTO, 1024)
	chCollCookieFood := make(chan *collissionCookieFoodDTO, 1024)

	world := &world{
		B2World:       box2d.MakeB2World(box2d.MakeB2Vec2(0, 0)),
		gSessions:     gs,
		width:         w,
		height:        h,
		minFPS:        minFPS,
		maxFPS:        maxFPS,
		currentFPS:    (maxFPS - minFPS) / 2,
		col2Cookies:   chColl2Cookies,
		colCookieFood: chCollCookieFood,
		speed:         speed,
		turboSpeed:    turboSpeed,
		minFoodCount:  minFoodCount,
		foodQueue:     list.NewLIFO(),
	}
	world.B2World.SetContactListener(newContactListener(chColl2Cookies, chCollCookieFood))
	return world
}

func (w *world) createWorld() {

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

	w.B2World.SetGravity(box2d.MakeB2Vec2(0, 0))

	createWorldBoundary(&w.B2World, w.width, 0, w.width, 0.1, true)
	createWorldBoundary(&w.B2World, w.width/2, w.height, w.width, 0.1, true)
	createWorldBoundary(&w.B2World, 0, w.height/2, 0.1, w.height, true)
	createWorldBoundary(&w.B2World, w.width, w.height/2, 0.1, w.height, true)
}

func (w *world) runSimulation(velocityIterations int, positionIterations int) {

	timeStep := time.Duration(time.Second / time.Duration(w.currentFPS))
	timeStepBox2D := float64(timeStep) / float64(time.Second) // Seconds as a float
	var notime int

	foodTicker := time.NewTicker(2 * time.Second)
	go func() {
		for {
			<-foodTicker.C
			w.worldMutex.Lock()
			w.adjustFood()
			w.worldMutex.Unlock()
		}
	}()

	go w.listenContactBetweenCookies()
	go w.listenContactBetweenCookiesAndFood()

	for {
		t1 := time.Now()

		w.worldMutex.Lock()

		w.B2World.Step(timeStepBox2D, velocityIterations, positionIterations)

		w.removeBodies()
		w.runFoodTasks()
		w.adjustSpeedsAndSizes()

		w.worldMutex.Unlock()

		elapsed := time.Since(t1)
		if elapsed < timeStep {
			notime--
			time.Sleep(timeStep - elapsed)
		} else {
			notime++
			log.Printf("WARNING: Cannot sustain frame rate. Expected time <%v>. Real time <%v>. FPS <%f>", timeStep, elapsed, w.currentFPS)
		}
		if notime < -10 && w.currentFPS < w.maxFPS {
			w.currentFPS++
			timeStep = time.Duration(time.Second / time.Duration(w.currentFPS))
			timeStepBox2D = float64(timeStep) / float64(time.Second) // Seconds as a float
			notime = 0
		}
		if notime > 4 && w.currentFPS > w.minFPS {
			w.currentFPS--
			timeStep = time.Duration(time.Second / time.Duration(w.currentFPS))
			timeStepBox2D = float64(timeStep) / float64(time.Second) // Seconds as a float
			notime = 0
		}
	}
}

func (w *world) adjustSpeedsAndSizes() {

	const penaltyDuration = float64(2 * time.Second)

	w.gSessions.Each(func(sessionID uint64) bool {

		if ok, _ := w.gSessions.IsPlaying(sessionID); !ok {
			return true
		}
		body, _ := w.gSessions.GetCookieBody(sessionID)

		data := body.GetUserData().(*Cookie)

		contactPenalty := math.Min(float64(time.Since(data.lastCookieContact)), penaltyDuration) / penaltyDuration

		inertia := body.GetInertia()

		body.SetAngularVelocity(0)

		speedX := body.GetLinearVelocity().X
		speedY := body.GetLinearVelocity().Y
		currentSpeed := math.Sqrt(math.Pow(speedX, 2) + math.Pow(speedY, 2))
		expectedSpeed := float64(w.speed)

		magnitude := 2 * (expectedSpeed - currentSpeed) * inertia * contactPenalty

		viewport, _ := w.gSessions.GetViewport(sessionID)

		vector := box2d.MakeB2Vec2(math.Cos(float64(viewport.Angle)), math.Sin(float64(viewport.Angle)))

		if magnitude < 0 {
			magnitude *= 0.005
		}
		vector.OperatorScalarMulInplace(magnitude)
		body.ApplyForce(vector, body.GetPosition(), true)

		// size and score
		score, _ := w.gSessions.GetScore(sessionID)
		if score != data.Score {
			data.Score = score

			body.DestroyFixture(body.GetFixtureList())
			body.CreateFixtureFromDef(w.getCookieFixtureDefByScore(score))
		}
		return true
	})
}

func (w *world) removeBodies() {
	w.bodies2Destroy.Range(func(id interface{}, body interface{}) bool {
		body.(*box2d.B2Body).SetActive(false)
		w.B2World.DestroyBody(body.(*box2d.B2Body))
		w.bodies2Destroy.Delete(id)
		return true
	})
}

func (w *world) removeCookie(body *box2d.B2Body) {
	w.worldMutex.Lock()
	body.SetActive(false)
	w.B2World.DestroyBody(body)
	w.worldMutex.Unlock()
}

func (w *world) runFoodTasks() {

	for i := 0; i < 5; i++ {
		o := w.foodQueue.Pop()

		if o == nil {
			return
		}

		task := o.(*throwFoodTask)

		for i := 0; i < task.count; i ++ {
			w.addFoodToWorld(task.x, task.y, 1, rand.Intn(100000))
		}
		atomic.AddUint64(&w.foodCount, uint64(task.count))
	}

}

func (w *world) adjustFood() {
	const N = 500

	foodCount := atomic.LoadUint64(&w.foodCount)

	if foodCount < w.minFoodCount {
		log.Println("ajustando", foodCount, w.minFoodCount)
		for i := 0; i < N; i++ {
			w.addFoodToWorld(float64(30+rand.Intn(int(w.width-30))), float64(30+rand.Intn(int(w.width-30))), uint64(1+rand.Intn(3)), 1000)
		}
		atomic.AddUint64(&w.foodCount, N)
	}
}

func (w *world) addFoodToWorld(x, y float64, score uint64, dispersion int) {
	if dispersion <= 0 {
		dispersion = 1
	}

	def := box2d.MakeB2BodyDef()
	def.Position.Set(x, y)
	def.Type = box2d.B2BodyType.B2_dynamicBody
	def.LinearDamping = 1
	def.FixedRotation = false
	def.AllowSleep = true

	// Shape
	shape := box2d.MakeB2CircleShape()
	shape.M_radius = 1

	// fixture
	fd := box2d.MakeB2FixtureDef()
	fd.Shape = &shape
	fd.Density = 1
	fd.Restitution = 0
	fd.Friction = 1

	// Create body
	body := w.B2World.CreateBody(&def)

	body.CreateFixtureFromDef(&fd)

	// Save link to session
	body.SetUserData(&Food{ID: rand.Uint64() << 8, Score: score, body: body, createdOn: time.Now()})

	body.ApplyForce(box2d.MakeB2Vec2(float64(2*rand.Intn(dispersion)-dispersion), float64(2*rand.Intn(dispersion)-dispersion)), body.GetPosition(), true)
}

func (w *world) addCookieToWorld(x float64, y float64, sessionID uint64, score uint64) *box2d.B2Body {

	w.worldMutex.Lock()
	defer w.worldMutex.Unlock()

	// Body definition
	def := box2d.MakeB2BodyDef()
	def.Position.Set(x, y)
	def.Type = box2d.B2BodyType.B2_dynamicBody
	def.FixedRotation = false
	def.AllowSleep = true
	def.LinearVelocity = box2d.MakeB2Vec2(float64(rand.Intn(w.speed)-10), float64(rand.Intn(w.speed)-10))
	def.LinearDamping = 1.0
	def.AngularDamping = 0.0
	def.Angle = rand.Float64() * 2 * math.Pi
	def.AngularVelocity = 10

	// Create body

	body := w.B2World.CreateBody(&def)

	body.CreateFixtureFromDef(w.getCookieFixtureDefByScore(score))

	// Save link to session
	body.SetUserData(&Cookie{ID: sessionID, Score: score, body: body, lastCookieContact: time.Now().Add(-5 * time.Second)})

	return body
}

func (w *world) listenContactBetweenCookies() {
	for {
		collision := <-w.col2Cookies

		cookie1, cookie2 := collision.cookie1, collision.cookie2
		playing1, err := w.gSessions.IsPlaying(cookie1.ID)
		if err != nil {
			fmt.Printf("Error on contact, <%s>", err)
			continue
		}

		playing2, err := w.gSessions.IsPlaying(cookie2.ID)
		if err != nil {
			fmt.Printf("Error on contact, <%s>", err)
			continue
		}

		if !playing1 || !playing2 {
			fmt.Println("######################## Colision con cookie que no está jugando ya ################")
			continue
		}

		score1, score2 := float64(cookie1.Score), float64(cookie2.Score)

		diff := math.Abs(score1 - score2)

		newScore1 := score1 - 0.1*score1 - diff
		newScore2 := score2 - 0.1*score2 - diff

		_ = w.gSessions.SetScore(cookie1.ID, uint64(math.Floor(newScore1)))
		_ = w.gSessions.SetScore(cookie2.ID, uint64(math.Floor(newScore2)))

		// Throw some food
		w.foodQueue.Push(newThrowFoodTask(int(math.Floor(diff)/2), (cookie1.body.GetPosition().X+cookie2.body.GetPosition().X)/2, (cookie1.body.GetPosition().Y+cookie2.body.GetPosition().Y)/2))

		if newScore1 < 50 {
			if err := w.gSessions.StopPlaying(cookie1.ID); err != nil {
				log.Println(err)
			}
			w.bodies2Destroy.Store(cookie1.ID, cookie1.body)

			// TODO: Notify explotion
			continue
		}
		if newScore2 < 50 {
			if err := w.gSessions.StopPlaying(cookie2.ID); err != nil {
				log.Println(err)
			}
			w.bodies2Destroy.Store(cookie2.ID, cookie2.body)

			// TODO: Notify explotion
			continue
		}

		data := cookie1.body.GetUserData().(*Cookie)
		data.lastCookieContact = time.Now()
		cookie1.body.SetUserData(data)

		data = cookie2.body.GetUserData().(*Cookie)
		data.lastCookieContact = time.Now()
		cookie2.body.SetUserData(data)

	}
}

func (w *world) listenContactBetweenCookiesAndFood() {
	for {
		collision := <-w.colCookieFood

		cookie := collision.cookie
		food := collision.food

		playing, err := w.gSessions.IsPlaying(cookie.ID)
		if err != nil {
			fmt.Printf("Error on contact, <%s>", err)
			continue
		}

		if !playing {
			fmt.Println("######################## Colision con cookie que no está jugando ya y comida ################")
			continue
		}

		err = w.gSessions.IncScore(cookie.ID, food.Score)
		if err!=nil {
			log.Printf("Error updating score, <%s>", err)
		}

		atomic.AddUint64(&w.foodCount, ^uint64(0)) // Decrement 1 :-)

		// adding body to the to be destroyed list.
		w.bodies2Destroy.Store(food.ID, food.body)
	}
}

func (w *world) getCookieFixtureDefByScore(score uint64) *box2d.B2FixtureDef {

	// Shape
	shape := box2d.MakeB2CircleShape()
	sc := float64(score)
	shape.M_radius = (math.Log2(sc) + math.Sqrt(sc)) / 2

	// fixture
	fd := box2d.MakeB2FixtureDef()
	fd.Shape = &shape
	fd.Density = 100 * math.Sqrt(sc)
	fd.Restitution = 1
	fd.Friction = 0.1
	return &fd
}

func (w *world) viewPort(v *sessionmanager.Viewport) ([]*box2d.B2Body, []*box2d.B2Body) {

	cookies := make([]*box2d.B2Body, 0)
	food := make([]*box2d.B2Body, 0)

	w.worldMutex.RLock()

	w.QueryAABB(
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
		box2d.B2AABB{LowerBound: box2d.MakeB2Vec2(float64(v.X), float64(v.Y)), UpperBound: box2d.MakeB2Vec2(float64(v.XX), float64(v.YY))},
	)

	w.worldMutex.RUnlock()

	return cookies, food
}
