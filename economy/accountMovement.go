package economy

import "time"

type AccountMovement struct {
	Money *Money
	Concept string
	Timestamp time.Time
}

func NewAccountMovement(m *Money, concept string, t time.Time) *AccountMovement {
	return &AccountMovement{Money:m, Concept:concept, Timestamp:t}
}


