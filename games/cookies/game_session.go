package cookies

import (
	"github.com/nu7hatch/gouuid"
	"sync"
	"github.com/pkg/errors"
	"fmt"
	"github.com/x1m3/elixir/games/cookies/messages"
)

type gameSession struct {
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

func (s *gameSession) updateViewPort(x float32, y float32, xx float32, yy float32, a float32, t bool) {
	s.viewport = &viewport{x: x, y: y, xx: xx, yy: xx, angle: a, turbo: t}
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

func (s *gameSessions) viewPortRequest(ID uuid.UUID) (*viewport, error) {
	s.RLock()
	defer s.RUnlock()

	session := s.sessions[ID] // Always should exist.
	viewport := session.viewport

	if !session.state.canSendScreenUpdates() || viewport == nil {
		return nil, errCannotSendScreenUpdates
	}

	return session.viewport, nil
}

func (s *gameSessions) UpdateViewPort(ID uuid.UUID, req *messages.ViewPortRequest) error {
	s.Lock()
	defer s.Unlock()

	session := s.sessions[ID] // Always should exist.

	if !session.state.canSendScreenUpdates() {
		return errCannotSendScreenUpdates
	}

	session.updateViewPort(req.X, req.Y, req.XX, req.YY, req.Angle, req.Turbo)

	return nil
}

func (s *gameSessions) Login(ID uuid.UUID, username string) error {
	s.Lock()
	defer s.Unlock()

	session := s.sessions[ID]
	if session.logged {
		return errUserWasLogged
	}

	session.state = &loggedState{}
	session.userName = username
	session.logged = true

	return nil
}

func (s *gameSessions) StartPlaying(ID uuid.UUID) error {
	s.Lock()
	defer s.Unlock()

	session := s.sessions[ID]
	if _, ok := session.state.(*loggedState); !ok {
		return fmt.Errorf("not logged user wants to play")
	}

	session.state = &playingState{}

	return nil
}
