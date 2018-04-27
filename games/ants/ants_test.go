package ants

import (
	"testing"
	"time"
	"sync"
)

func TestAnts_Init(t *testing.T) {
	game := New(200, 200, 100)
	game.Init()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for i := 0; i < 100; i++ {
			ants := game.ProcessCommand(&ViewPortRequest{X: 0, Y: 0, XX: 200, YY: 200}).(*ViewportResponse).Ants
			_ = ants
/*
			for _, ant := range ants {
				fmt.Printf("(ant:%d)[%f, %f]  ", ant.ID, ant.X, ant.Y)
			}
			fmt.Println()
*/
			time.Sleep(1 * time.Millisecond)
		}
		wg.Done()
	}()
	wg.Wait()
}