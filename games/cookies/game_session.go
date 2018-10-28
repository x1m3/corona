package cookies

import (
	"github.com/nu7hatch/gouuid"
	"sync"
)

type gameSession struct {
	ID         uuid.UUID
	viewportX  float64
	viewportY  float64
	viewportXX float64
	viewportYY float64
}

func (s *gameSession) viewPortRequest() (float64, float64, float64, float64) {
	return s.viewportX, s.viewportY, s.viewportXX, s.viewportYY
}

func (s *gameSession) updateViewPort(x float64, y float64, xx float64, yy float64) {
	s.viewportX, s.viewportY, s.viewportXX, s.viewportYY = x, y, xx, yy
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
	s.sessions[*ID] = &gameSession{ID: *ID}
	s.Unlock()
	return *ID
}

func (s *gameSessions) viewPortRequest(ID uuid.UUID) (float64, float64, float64, float64) {
	s.RLock()
	defer s.RUnlock()
	return s.sessions[ID].viewPortRequest()
}

func (s *gameSessions) UpdateViewPort(ID uuid.UUID, x float64, y float64, xx float64, yy float64) {
	s.Lock()
	s.sessions[ID].updateViewPort(x, y, xx, yy)
	s.Unlock()
}
