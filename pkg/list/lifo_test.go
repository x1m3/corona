package list

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLIFO_PushPop(t *testing.T) {
	const N = 1000

	l := NewLIFO()
	for i := 0; i < N; i++ {
		l.Push(i)
	}

	for i := N - 1; i >= 0; i-- {
		v := l.Pop().(int)
		assert.Equal(t, i, v)
	}

	// List is empty
	assert.Equal(t, nil, l.Pop())
}
