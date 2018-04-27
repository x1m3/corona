package engine

import "github.com/x1m3/elixir/random"

const ANY_VALUE = -1

type WinRule3Wheels struct {
	v1         int8
	v2         int8
	v3         int8
	multiplier int64
}

func NewWinRule3Wheels(v1, v2, v3 int8, multiplier int64) *WinRule3Wheels {
	rule := &WinRule3Wheels{v1: v1, v2: v2, v3: v3, multiplier: multiplier}

	return rule
}

type WinTable3Wheels struct {
	rules []WinRule3Wheels
}

func (r *WinRule3Wheels) Evaluate(v1, v2, v3 int8) int64 {
	if r.v1 != ANY_VALUE && r.v1 != v1 {
		return 0
	}
	if r.v2 != ANY_VALUE && r.v2 != v2 {
		return 0
	}
	if r.v3 != ANY_VALUE && r.v3 != v3 {
		return 0
	}
	return r.multiplier
}

func NewWinTable3Wheels() *WinTable3Wheels {
	return &WinTable3Wheels{rules: make([]WinRule3Wheels, 0)}
}

func (t *WinTable3Wheels) Add(r *WinRule3Wheels) {
	t.rules = append(t.rules, *r)
}

func (s *WinTable3Wheels) WinMultiplier(v1, v2, v3 int8) int64 {
	for _, rule := range s.rules {
		m := rule.Evaluate(v1, v2, v3)
		if m != 0 {
			return m
		}
	}
	return 0
}

type Slot3X24 struct {
	random random.Generator
	wheel1 [24]int8
	wheel2 [24]int8
	wheel3 [24]int8
	p1     int8
	p2     int8
	p3     int8
	rules  *WinTable3Wheels
}

func New3x24(r random.Generator, w1 [24]int8, w2 [24]int8, w3 [24]int8, rules *WinTable3Wheels) *Slot3X24 {
	return &Slot3X24{random: r, wheel1: w1, wheel2: w2, wheel3: w3, rules: rules}
}

// Returns an init position for each wheel with a no prize combination
func (s *Slot3X24) Init() (w1, w2, w3 [24]int8, p1, p2, p3 int8) {
	var win int64
	for win = 100; win != 0; win, p1, p2, p3 = s.SpinAll(1) {}
	return s.wheel1, s.wheel2, s.wheel3, p1, p2, p3
}

func (s *Slot3X24) SpinAll(bet int64) (win int64, p1, p2, p3 int8) {
	p1 = int8(s.random.Next(24))
	p2 = int8(s.random.Next(24))
	p3 = int8(s.random.Next(24))

	s.p1, s.p2, s.p3 = p1, p2, p3
	return bet * s.rules.WinMultiplier(s.wheel1[p1], s.wheel2[p2], s.wheel3[p3]), p1, p2, p3
}

func (s *Slot3X24) State() (p1, p2, p3 int8) {
	return p1, p2, p3
}
