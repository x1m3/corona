package corona

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ByteArena/box2d"

	"github.com/x1m3/corona/internal/corona/mybox2d"
	"github.com/x1m3/corona/internal/corona/sessionmanager"
	"github.com/x1m3/corona/internal/messages"
	"github.com/x1m3/corona/pkg/list"
)

type world struct {
	worldMutex sync.RWMutex

	gSessions *sessionmanager.Sessions

	box2d.B2World
	width  float64
	height float64

	updateClientPeriod time.Duration

	minFPS     float64
	maxFPS     float64
	currentFPS float64

	col2Cookies   chan *collision2CookiesDTO
	colCookieFood chan *collissionCookieFoodDTO

	speed          int
	turboSpeed     int
	minFoodCount   uint64
	foodCount      uint64
	bodies2Destroy list.LIFO
	foodQueue      list.LIFO
}

func NewWorld(gs *sessionmanager.Sessions, w, h, minFPS, maxFPS float64, speed, turboSpeed int, minFoodCount uint64, updateClientPeriod time.Duration) *world {

	chColl2Cookies := make(chan *collision2CookiesDTO, 1024)
	chCollCookieFood := make(chan *collissionCookieFoodDTO, 1024)

	world := &world{
		B2World:            box2d.MakeB2World(box2d.MakeB2Vec2(0, 0)),
		gSessions:          gs,
		width:              w,
		height:             h,
		updateClientPeriod: updateClientPeriod,
		minFPS:             minFPS,
		maxFPS:             maxFPS,
		currentFPS:         (maxFPS + minFPS) / 2,
		col2Cookies:        chColl2Cookies,
		colCookieFood:      chCollCookieFood,
		speed:              speed,
		turboSpeed:         turboSpeed,
		minFoodCount:       minFoodCount,
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

	go w.adjustFood(2 * time.Second)
	go w.broadcastStats(5 * time.Second)
	go w.listenContactBetweenCookies()
	go w.listenContactBetweenCookiesAndFood()

	i := 0
	for {
		i++
		t1 := time.Now()

		w.worldMutex.Lock()

		w.B2World.Step(timeStepBox2D, velocityIterations, positionIterations)
		w.B2World.ClearForces()

		w.removeBodies()
		if i%7 == 0 {
			w.runFoodTasks()
		}
		if i%5 == 0 {
			w.adjustSpeedsAndSizes()
		}

		w.updateViewportResponses()

		w.worldMutex.Unlock()

		elapsed := time.Since(t1)
		if elapsed < timeStep {
			notime--
			time.Sleep(timeStep - elapsed)
		} else {
			notime++
			log.Printf("WARNING: Cannot sustain frame rate. Expected time <%v>. Real time <%v>. FPS <%f>", timeStep, elapsed, w.currentFPS)
		}
		if notime < -60 && w.currentFPS < w.maxFPS {
			log.Println("FPS up")
			w.currentFPS++
			timeStep = time.Duration(time.Second / time.Duration(w.currentFPS))
			timeStepBox2D = float64(timeStep) / float64(time.Second) // Seconds as a float
			notime = 0
		}
		if notime > 0 && w.currentFPS > w.minFPS {
			log.Println("FPS down")
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

		viewport, _ := w.gSessions.GetViewportRequest(sessionID)

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
			body.CreateFixtureFromDef(mybox2d.GetCookieFixtureDefByScore(score))
		}
		return true
	})
}

func (w *world) removeBodies() {
	for {
		o := w.bodies2Destroy.Pop()
		if o == nil {
			return
		}
		body := o.(*box2d.B2Body)
		body.SetActive(false)
		w.B2World.DestroyBody(body)
	}
}

func (w *world) removeCookie(body *box2d.B2Body) {
	w.bodies2Destroy.Push(body)
}

func (w *world) runFoodTasks() {

	for {
		o := w.foodQueue.Pop()
		if o == nil {
			return
		}

		task := o.(throwFoodTask)

		for i := 0; i < task.count; i++ {
			w.addFoodToWorld(task.x, task.y, 1, rand.Intn(100000))
		}
		atomic.AddUint64(&w.foodCount, uint64(task.count))
	}

}

func (w *world) broadcastStats(d time.Duration) {
	ticker := time.NewTicker(d)
	for {
		<-ticker.C
		stats := messages.NewStatsResponse(w.foodCount, w.gSessions.Count())
		w.broadcast(stats)
	}
}

func (w *world) adjustFood(d time.Duration) {
	const N = 500

	ticker := time.NewTicker(d)
	for {
		<-ticker.C
		foodCount := atomic.LoadUint64(&w.foodCount)

		if foodCount < w.minFoodCount {
			log.Println("ajustando", foodCount, w.minFoodCount)
			for i := 0; i < N; i++ {
				w.foodQueue.Push(throwFoodTask{count: 1, x: float64(30 + rand.Intn(int(w.width-30))), y: float64(30 + rand.Intn(int(w.width-30)))})
			}
		}
	}
}

func (w *world) addFoodToWorld(x, y float64, score uint64, dispersion int) {
	if dispersion <= 0 {
		dispersion = 1
	}

	def := mybox2d.GetFoodBodyDef()
	fd := mybox2d.GetFoodFixtureDef()

	// Create body
	body := w.B2World.CreateBody(def)

	body.CreateFixtureFromDef(fd)

	body.SetTransform(box2d.MakeB2Vec2(x, y), 0)

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

	body.CreateFixtureFromDef(mybox2d.GetCookieFixtureDefByScore(score))

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

		var diff, newScore1, newScore2, ratio1, ratio2 float64

		if score1 > score2 {
			ratio1, ratio2 = score2/score1, 1-(score2/score1)
			diff = math.Min(score1-score2, score2)

		} else {
			ratio1, ratio2 = 1-(score1/score2), score1/score2
			diff = math.Min(score2-score1, score1)
		}

		newScore1 = math.Max(0, score1-0.1*score1-diff*ratio1)
		newScore2 = math.Max(0, score2-0.1*score2-diff*ratio2)

		_ = w.gSessions.SetScore(cookie1.ID, uint64(math.Floor(newScore1)))
		_ = w.gSessions.SetScore(cookie2.ID, uint64(math.Floor(newScore2)))

		// Throw some food
		w.foodQueue.Push(throwFoodTask{count: int(math.Floor(diff)), x: (cookie1.body.GetPosition().X + cookie2.body.GetPosition().X) / 2, y: (cookie1.body.GetPosition().Y + cookie2.body.GetPosition().Y) / 2})

		if newScore1 < 50 {
			if err := w.gSessions.StopPlaying(cookie1.ID); err != nil {
				log.Println(err)
			}

			w.bodies2Destroy.Push(cookie1.body)

			// TODO: Notify explotion
			continue
		}
		if newScore2 < 50 {
			if err := w.gSessions.StopPlaying(cookie2.ID); err != nil {
				log.Println(err)
			}
			w.bodies2Destroy.Push(cookie2.body)

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
		if err != nil {
			log.Printf("Error updating score, <%s>", err)
		}

		atomic.AddUint64(&w.foodCount, ^uint64(0)) // Decrement 1 :-)

		// adding body to the to be destroyed list.
		w.bodies2Destroy.Push(food.body)
	}
}

func (w *world) updateViewportResponses() {
	w.gSessions.EachParallel(
		func(sessionID uint64) {
			needsUpdate, v, respCh, err := w.gSessions.GetViewportRequestEnhanced(sessionID, w.updateClientPeriod)
			if !needsUpdate || err != nil {
				return
			}
			respCh <- w.viewPort(v)
		})
}

func (w *world) viewPort(v *sessionmanager.Viewport) *messages.ViewportResponse {

	response := &messages.ViewportResponse{}
	response.Type = messages.ViewPortResponseType

	response.Cookies = make([]*messages.CookieInfo, 0)
	response.Food = make([]*messages.FoodInfo, 0)

	w.QueryAABB(
		func(fixture *box2d.B2Fixture) bool {
			info := fixture.M_body.GetUserData()
			pos := fixture.M_body.GetPosition()
			switch info.(type) {
			case *Cookie:
				response.Cookies = append(
					response.Cookies,
					&messages.CookieInfo{
						ID:    info.(*Cookie).ID,
						Score: info.(*Cookie).Score,
						X:     float32(pos.X),
						Y:     float32(pos.Y),
					})
			case *Food:
				response.Food = append(
					response.Food,
					&messages.FoodInfo{
						ID:    info.(*Food).ID,
						Score: info.(*Food).Score,
						X:     float32(pos.X),
						Y:     float32(pos.Y),
					})
			}
			return true
		},
		box2d.B2AABB{LowerBound: box2d.MakeB2Vec2(float64(v.X), float64(v.Y)), UpperBound: box2d.MakeB2Vec2(float64(v.XX), float64(v.YY))},
	)

	return response
}

func (w *world) broadcast(message interface{}) {
	w.gSessions.EachParallel(func(id uint64) {
		ch, err := w.gSessions.GetResponseChannel(id)
		if err != nil {
			return
		}
		ch <- message
	})
}
