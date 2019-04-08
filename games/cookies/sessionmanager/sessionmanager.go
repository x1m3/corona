package sessionmanager

import (
	"github.com/ByteArena/box2d"
	"github.com/pkg/errors"
	"github.com/x1m3/elixir/games/cookies/messages"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type Viewport struct {
	X     float32
	Y     float32
	XX    float32
	YY    float32
	Angle float32
	Turbo bool
}

type ViewPortResponse struct {
	Cookies []*box2d.B2Body
	Food    []*box2d.B2Body
}

var errSessionNotFound = errors.New("session not found")
var errViewportResponseEmpty = errors.New("viewportresponse is still empty")
var errUserWasLogged = errors.New("user already logged")
var errCannotSendScreenUpdates = errors.New("cannot send screen updates")

const ReadMode = 1
const WriteMode = 2

type gameSessionFunc func(s *gameSession) (interface{}, error)

type gameSession struct {
	ID                          uint64
	userName                    string
	score                       uint64
	state                       state
	viewportRequest             Viewport
	lastViewportResponseRequest time.Time
	viewportResponseCh          chan *messages.ViewportResponse
	box2dbody                   *box2d.B2Body
}

func newGameSession(id uint64) *gameSession {
	return &gameSession{
		ID:                          id,
		state:                       &notLoggedState{},
		score:                       100,
		lastViewportResponseRequest: time.Now(),
		viewportResponseCh:          make(chan *messages.ViewportResponse, 1000),
	}
}

func (s *gameSession) getViewportRequest() (*Viewport, error) {

	if !s.state.canSendScreenUpdates() {
		return nil, errCannotSendScreenUpdates
	}

	return &s.viewportRequest, nil
}

func (s *gameSession) updateViewportRequest(x, y, xx, yy float32, a float32, t bool) {
	if s.state.canSendScreenUpdates() {
		s.viewportRequest.X = x
		s.viewportRequest.Y = y
		s.viewportRequest.XX = xx
		s.viewportRequest.YY = yy
		s.viewportRequest.Angle = a
		s.viewportRequest.Turbo = t
	}
}

func (s *gameSession) login(username string) error {

	if s.inLoggedState() {
		return errUserWasLogged
	}

	s.state = &loggedState{}
	s.userName = username

	return nil
}

func (s *gameSession) inLoggedState() bool {

	_, logged := s.state.(*loggedState)
	return logged
}

func (s *gameSession) inPlayingState() bool {
	_, playing := s.state.(*playingState)
	return playing
}

func (s *gameSession) startPlaying() error {

	if _, ok := s.state.(*loggedState); !ok {
		return errors.New("not logged user wants to play")
	}
	s.state = &playingState{}

	return nil
}

func (s *gameSession) stopPlaying() error {

	if _, ok := s.state.(*playingState); !ok {
		return errors.New("not playing user wants to stop playing")
	}

	s.state = &loggedState{}

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
	s.box2dbody = b
}

func (s *gameSession) getBox2DBody() *box2d.B2Body {
	return s.box2dbody
}

type Sessions struct {
	sync.RWMutex
	sessions map[uint64]*gameSession
}

func New() *Sessions {
	return &Sessions{
		sessions: make(map[uint64]*gameSession),
	}
}

func (s *Sessions) Add() uint64 {
	ID := rand.Uint64() << 8 // Javascript does not support number larger than 57 bits. Let's avoid problems.
	s.Lock()
	s.sessions[ID] = newGameSession(ID)
	s.Unlock()
	return ID
}

func (s *Sessions) Close(id uint64) error {
	_, err := s.ensure(
		id,
		func() gameSessionFunc {
			return func(session *gameSession) (interface{}, error) {
				close(s.sessions[id].viewportResponseCh)
				delete(s.sessions, session.ID)
				return nil, nil
			}
		}(),
		WriteMode)

	return err
}

func (s *Sessions) Login(id uint64, username string) error {
	_, err := s.ensure(
		id,
		func() gameSessionFunc {
			return func(session *gameSession) (interface{}, error) {
				return nil, session.login(username)
			}
		}(),
		WriteMode)
	return err
}

func (s *Sessions) GetCookieBody(id uint64) (*box2d.B2Body, error) {
	body, err := s.ensure(
		id,
		func() gameSessionFunc {
			return func(session *gameSession) (interface{}, error) {
				return session.box2dbody, nil
			}
		}(),
		ReadMode)

	if err != nil {
		return nil, err
	}
	return body.(*box2d.B2Body), err
}

func (s *Sessions) GetScore(id uint64) (uint64, error) {
	score, err := s.ensure(
		id,
		func() gameSessionFunc {
			return func(session *gameSession) (interface{}, error) {
				return session.getScore(), nil
			}
		}(),
		ReadMode)
	return score.(uint64), err
}

func (s *Sessions) GetViewportRequest(id uint64) (*Viewport, error) {
	v, err := s.ensure(
		id,
		func() gameSessionFunc {
			return func(session *gameSession) (interface{}, error) {
				return session.getViewportRequest()
			}
		}(),
		ReadMode)
	if err != nil {
		return nil, err
	}
	return v.(*Viewport), err
}

func (s *Sessions) GetViewportResponseChannel(id uint64) (chan *messages.ViewportResponse, error) {
	v, err := s.ensure(
		id,
		func() gameSessionFunc {
			return func(session *gameSession) (interface{}, error) {
				return session.viewportResponseCh, nil
			}
		}(),
		ReadMode)
	if err != nil {
		return nil, err
	}
	return v.(chan *messages.ViewportResponse), err
}

func (s *Sessions) IsLogged(id uint64) (bool, error) {
	logged, err := s.ensure(
		id,
		func() gameSessionFunc {
			return func(session *gameSession) (interface{}, error) {
				return session.inLoggedState(), nil
			}

		}(),
		ReadMode)
	return logged.(bool), err
}

func (s *Sessions) IsPlaying(id uint64) (bool, error) {
	logged, err := s.ensure(
		id,
		func() gameSessionFunc {
			return func(session *gameSession) (interface{}, error) {
				return session.inPlayingState(), nil
			}

		}(),
		ReadMode)
	if err != nil {
		return false, err
	}
	return logged.(bool), err
}

func (s *Sessions) StartPlaying(id uint64) error {
	_, err := s.ensure(
		id,
		func() gameSessionFunc {
			return func(session *gameSession) (interface{}, error) {
				return nil, session.startPlaying()
			}
		}(),
		WriteMode)
	return err
}

func (s *Sessions) StopPlaying(id uint64) error {
	_, err := s.ensure(
		id,
		func() gameSessionFunc {
			return func(session *gameSession) (interface{}, error) {
				return nil, session.stopPlaying()
			}
		}(),
		WriteMode)
	return err
}

func (s *Sessions) SetCookieBody(id uint64, body *box2d.B2Body) error {
	_, err := s.ensure(
		id,
		func() gameSessionFunc {
			return func(session *gameSession) (interface{}, error) {
				session.box2dbody = body
				return nil, nil
			}
		}(),
		WriteMode)
	return err
}

func (s *Sessions) SetViewportRequest(id uint64, x, y, xx, yy float32, a float32, t bool) error {
	_, err := s.ensure(
		id,
		func() gameSessionFunc {
			return func(session *gameSession) (interface{}, error) {
				session.updateViewportRequest(x, y, xx, yy, a, t)
				return nil, nil
			}
		}(),
		WriteMode)
	return err
}

func (s *Sessions) SetScore(id uint64, score uint64) error {
	_, err := s.ensure(
		id,
		func() gameSessionFunc {
			return func(session *gameSession) (interface{}, error) {
				session.setScore(score)
				return nil, nil
			}
		}(),
		WriteMode)
	return err
}

func (s *Sessions) UpdateLastViewportRequestTime(id uint64) {
	_, _ = s.ensure(
		id,
		func() gameSessionFunc {
			return func(session *gameSession) (interface{}, error) {
				session.lastViewportResponseRequest = time.Now()
				return true, nil
			}
		}(),
		WriteMode)
}

func (s *Sessions) ShouldUpdateViewportResponse(id uint64, updatePeriod time.Duration) bool {
	toUpdate, err := s.ensure(
		id,
		func() gameSessionFunc {
			return func(session *gameSession) (interface{}, error) {
				if time.Since(session.lastViewportResponseRequest) > updatePeriod {
					return true, nil
				}
				return false, nil
			}
		}(),
		ReadMode)
	if err != nil {
		return false
	}
	return toUpdate.(bool)
}

func (s *Sessions) IncScore(id uint64, score uint64) error {
	_, err := s.ensure(
		id,
		func() gameSessionFunc {
			return func(session *gameSession) (interface{}, error) {
				session.incScore(score)
				return nil, nil
			}
		}(),
		WriteMode)
	return err
}

func (s *Sessions) ensure(id uint64, fn gameSessionFunc, lockMode uint8) (interface{}, error) {
	var session *gameSession
	var found bool

	if lockMode == ReadMode {
		s.RLock()
		defer s.RUnlock()
	} else {
		s.Lock()
		defer s.Unlock()
	}

	if session, found = s.sessions[id]; !found {
		return nil, errSessionNotFound
	}
	return fn(session)
}

func (s *Sessions) session(id uint64) *gameSession {
	s.RLock()
	defer s.RUnlock()
	return s.sessions[id]
}

func (s *Sessions) Each(fn func(id uint64) bool) {

	s.Lock()
	sessionIDs := make([]uint64, 0, len(s.sessions))
	for id := range s.sessions {
		sessionIDs = append(sessionIDs, id)
	}
	s.Unlock()

	for _, sessionID := range sessionIDs {
		if !fn(sessionID) {
			return
		}
	}
}


func (s *Sessions) EachParallel(fn func(id uint64) bool) {

	s.Lock()
	sessionIDs := make([]uint64, 0, len(s.sessions))
	for id := range s.sessions {
		sessionIDs = append(sessionIDs, id)
	}
	s.Unlock()

	wg :=sync.WaitGroup{}
	for _, sessionID := range sessionIDs {
		wg.Add(1)
		go func(sessionID uint64) {
			fn(sessionID)
			wg.Done()
		}(sessionID)
	}
	wg.Wait()
}