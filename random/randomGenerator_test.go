package random

import (
	"testing"

	"math"
	"time"
)

func TestSimpleRandomGenerator(t *testing.T) {

	const TRIES = 1000000
	const SAMPLE = 1000
	const MARGIN = 0.001

	gen := NewSimpleRandomGenerator()
	gen.Init(5 * time.Millisecond)

	mem := make(map[int]int, SAMPLE)

	for i := 0; i < TRIES; i++ {
		r := int(gen.Next(SAMPLE))
		if _, found := mem[r]; found {
			mem[r]++
		} else {
			mem[r] = 0
		}
	}

	for number, times := range mem {
		if MARGIN < math.Abs(float64(times)/TRIES-1/float64(SAMPLE)) {
			t.Errorf("Probability for number %d is out of margin. Expecting %f, got %f", number, MARGIN, math.Abs(float64(times)/TRIES-1/float64(SAMPLE)))
		}
	}
}
