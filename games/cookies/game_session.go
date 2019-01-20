package cookies

import (
	"github.com/nu7hatch/gouuid"
	"sync"
	"github.com/pkg/errors"
)

type gameSession struct {
	sync.RWMutex
	ID       uuid.UUID
	userName string
	logged   bool
	state    state
	viewport *viewport
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
var errCannotSendScreenUpdates = errors.New("session not found")

func newGameSession(id uuid.UUID) *gameSession {
	return &gameSession{ID: id, state: &notLoggedState{}}
}

func (s *gameSession) getViewport() (*viewport, error) {
	s.RLock()
	defer s.RUnlock()

	if !s.state.canSendScreenUpdates() || s.viewport == nil {
		return nil, errCannotSendScreenUpdates
	}

	return s.viewport, nil
}

func (s *gameSession) updateViewPort(x float32, y float32, xx float32, yy float32, a float32, t bool) {
	s.Lock()
	if s.state.canSendScreenUpdates() {
		s.viewport = &viewport{x: x, y: y, xx: xx, yy: yy, angle: a, turbo: t}
	}
	s.Unlock()
}


func (s *gameSession) login(username string) error{
	s.Lock()
	defer s.Unlock()

	if s.logged {
		return errUserWasLogged
	}
	s.state = &loggedState{}
	s.userName = username
	s.logged = true

	return nil
}

func (s *gameSession) startPlaying() error {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.state.(*loggedState); !ok {
		return errors.New("not logged user wants to play")
	}
	s.state = &playingState{}

	return nil
}

type gameSessions struct {
	sync.RWMutex
	sessions map[uuid.UUID]*gameSession
}

func newGameSessions() *gameSessions {
	return &gameSessions{
		sessions: make(map[uuid.UUID]*gameSession),
	}
}

func (s *gameSessions) add() uuid.UUID {
	ID, _ := uuid.NewV4()
	s.Lock()
	s.sessions[*ID] = newGameSession(*ID)
	s.Unlock()
	return *ID
}

func (s *gameSessions) session(id uuid.UUID) *gameSession {
	s.RLock()
	defer s.RUnlock()
	return s.sessions[id]
}

func (s *gameSessions) StartPlaying(ID uuid.UUID) error {
	return s.session(ID).startPlaying()
}
