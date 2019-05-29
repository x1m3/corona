package main

import (
	"fmt"
	"github.com/x1m3/elixir/games/othello"
	"os"
	"log"
	"runtime/pprof"
)

func main() {

	f, err := os.Create("./cpuprofile")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()


	board := othello.NewBoard(8,8)
	board.Init()

	for {
		whiteCannotPlay := false
		fmt.Println("WHITE PLAYS:")
		_, err := board.ComputerMove(othello.WHITE)
		if err != nil {
			whiteCannotPlay = true
			fmt.Println(err)

		}
		fmt.Print(string(board.Dump()))
		b, w, e := board.Count()
		fmt.Printf("WHITES[X]:%d, BLACKS[0]:%d, EMPTIES:%d\n\n", w, b, e)

		blackCannotPlay := false
		fmt.Println("BLACK PLAYS:")
		_, err = board.ComputerMoveMinMax(othello.BLACK)
		if err != nil {
			blackCannotPlay = true
			fmt.Println(err)

		}
		fmt.Print(string(board.Dump()))
		b, w, e = board.Count()
		fmt.Printf("WHITES[X]:%d, BLACKS[0]:%d, EMPTIES:%d\n\n", w, b, e)

		if whiteCannotPlay && blackCannotPlay {
			break
		}
	}
}