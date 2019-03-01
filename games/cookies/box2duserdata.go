package cookies

import "github.com/ByteArena/box2d"

type Food struct {
	ID    uint64
	Score uint64
	body *box2d.B2Body
}

type Cookie struct {
	ID uint64
	Score uint64
	body *box2d.B2Body
}
