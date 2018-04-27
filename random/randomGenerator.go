package random

import (
	"time"
	"math/rand"
)

var simpleGenerator *SimpleRandomGenerator

func init() {
	simpleGenerator = NewSimpleRandomGenerator()
	simpleGenerator.Init(500 * time.Millisecond)
}

type Generator interface {
	Next(int64) int64
}

func GetSimpleRandomGenerator() *SimpleRandomGenerator{
	return simpleGenerator
}

type SimpleRandomGenerator struct {
	source rand.Source
	r      chan int64
	ticker *time.Ticker
}

func NewSimpleRandomGenerator() *SimpleRandomGenerator {
	return &SimpleRandomGenerator{
		source: rand.NewSource(time.Now().UnixNano()),
		r:      make(chan int64, 1024),
	}
}

func (g *SimpleRandomGenerator) Init(discardTime time.Duration) {
	go g.next() // Start generating random numbers and fill the channel
	g.ticker = time.NewTicker(discardTime)
	go g.discardSome(100) // Discard some numbers from time to time
}

func (g *SimpleRandomGenerator) Next(max int64) int64 {
	next := <-g.r
	return next % max
}

func (g *SimpleRandomGenerator) next() {
	for {
		g.r <- rand.Int63()
	}
}

func (g *SimpleRandomGenerator) discardSome(max int64 ) {
	for range g.ticker.C {
		for i := 0; i < int(g.Next(max)); i++ {
			<-g.r
		}
	}
}
