package cookies

import (
	"github.com/ByteArena/box2d"
	"time"
)

type Food struct {
	ID    uint64
	Score uint64
	body *box2d.B2Body
	createdOn time.Time
}

type Cookie struct {
	ID uint64
	Score uint64
	body *box2d.B2Body
	lastCookieContact time.Time
}
