package sessionmanager

type state interface {
	mustBeLogged() bool
	canSendScreenUpdates() bool
}


type notLoggedState struct {}

func (s *notLoggedState) mustBeLogged() bool {
	return false
}

func (s *notLoggedState) canSendScreenUpdates() bool {
	return false
}

type loggedState struct {}

func (s *loggedState) mustBeLogged() bool {
	return true
}

func (s *loggedState) canSendScreenUpdates() bool {
	return false
}

type playingState struct {}


func (s *playingState) mustBeLogged() bool {
	return true
}

func (s *playingState) canSendScreenUpdates() bool {
	return true
}

type endGameState struct {}


func (s *endGameState) mustBeLogged() bool {
	return true
}

func (s *endGameState) canSendScreenUpdates() bool {
	return false
}


