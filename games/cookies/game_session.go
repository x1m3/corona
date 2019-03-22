package cookies

import (
	"sync"
	"github.com/pkg/errors"
	"math/rand"
	"github.com/ByteArena/box2d"
	"sync/atomic"
)

type gameSession struct {
	sync.RWMutex
	ID       uint64
	userName string
	score    uint64
	state    state
	viewport viewport
	box2dbody *box2d.B2Body
}

type viewport struct {
	x     float32
	y     float32
	xx    float32
	yy    float32
	angle float32
	turbo bool
}

var errUserWasLogged = errors.New("user already logged")
var errCannotSendScreenUpdates = errors.New("cannot send screen updates")

func newGameSession(id uint64) *gameSession {
	return &gameSession{ID: id, state: &notLoggedState{}, score: 100}
}

func (s *gameSession) getViewport() (*viewport, error) {
	s.RLock()
	defer s.RUnlock()

	if !s.state.canSendScreenUpdates() {
		return nil, errCannotSendScreenUpdates
	}

	return &s.viewport, nil
}

func (s *gameSession) updateViewPort(x float32, y float32, xx float32, yy float32, a float32, t bool) {
	s.Lock()
	if s.state.canSendScreenUpdates() {
		s.viewport.x = x
		s.viewport.y = y
		s.viewport.xx = xx
		s.viewport.yy = yy
		s.viewport.angle = a
		s.viewport.turbo = t
	}
	s.Unlock()
}

func (s *gameSession) login(username string) error {

	if s.inLoggedState() {
		return errUserWasLogged
	}
	s.Lock()
	s.state = &loggedState{}
	s.userName = username
	s.Unlock()

	return nil
}

func (s *gameSession) inLoggedState() bool {
	s.RLock()
	_,  logged := s.state.(*loggedState)
	s.RUnlock()
	return logged
}

func (s *gameSession) startPlaying() error {

	if _, ok := s.state.(*loggedState); !ok {
		return errors.New("not logged user wants to play")
	}

	s.Lock()
	s.state = &playingState{}
	s.Unlock()

	return nil
}

func (s *gameSession) stopPlaying() error {

	if _, ok := s.state.(*playingState); !ok {
		return errors.New("not playing user wants to stop playing")
	}

	s.Lock()
	s.state = &loggedState{}
	s.Unlock()

	return nil
}

func (s *gameSession) getScore() uint64 {
	return atomic.LoadUint64(&s.score)
}

func (s *gameSession) setScore(i uint64) {
	atomic.StoreUint64(&s.score, i)
}

func (s *gameSession) incScore(i uint64) {
	atomic.AddUint64(&s.score, i)
}


func (s *gameSession) setBox2DBody(b *box2d.B2Body) {
	s.Lock()
	s.box2dbody = b
	s.Unlock()
}

func (s *gameSession) getBox2DBody() *box2d.B2Body {
	s.RLock()
	defer s.RUnlock()
	return s.box2dbody
}




type gameSessions struct {
	sync.RWMutex
	sessions map[uint64]*gameSession
}

func newGameSessions() *gameSessions {
	return &gameSessions{
		sessions: make(map[uint64]*gameSession),
	}
}

func (s *gameSessions) add() uint64 {
	ID := rand.Uint64() << 8// Javascript does not support number larger than 57 bits. Let's avoid problems.
	s.Lock()
	s.sessions[ID] = newGameSession(ID)
	s.Unlock()
	return ID
}

func (s *gameSessions) close(ID uint64)  {
	s.Lock()
	_ = s.sessions[ID].stopPlaying()
	delete(s.sessions, ID)
	s.Unlock()
	return
}

func (s *gameSessions) session(id uint64) *gameSession {
	s.RLock()
	defer s.RUnlock()
	return s.sessions[id]
}

func (s *gameSessions) each(fn func(s *gameSession) bool) {
	s.Lock()
	defer s.Unlock()
	for _, session := range s.sessions {
		if !fn(session) {
			return
		}
	}
}
