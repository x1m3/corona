package othello

import (
	"testing"
	"fmt"
	"math/rand"
)

func TestCanMove(t *testing.T) {

	board := NewBoard(8, 8)
	board.Init()

	// horizontal
	if got, expected := len(board.EvalMove(BLACK, 3, 3)), 0; got != expected {
		t.Error("Cannot move into a played position.")
	}

	if got, expected := len(board.EvalMove(BLACK, 0, 0)), 0; got != expected {
		t.Error("Cannot move into a isolated position.")
	}

	if got, expected := len(board.EvalMove(BLACK, 0, 3)), 0; got != expected {
		t.Error("Cannot move into a isolated position.")
	}

	if got, expected := len(board.EvalMove(BLACK, 0, 4)), 0; got != expected {
		t.Error("Cannot move into a isolated position.")
	}

	if got, expected := len(board.EvalMove(BLACK, 7, 3)), 0; got != expected {
		t.Error("Cannot move into a isolated position.")
	}

	if got, expected := len(board.EvalMove(BLACK, 7, 4)), 0; got != expected {
		t.Error("Cannot move into a isolated position.")
	}

	if got, expected := len(board.EvalMove(BLACK, 2, 4)), 0; got != expected {
		t.Error("Cannot move into a this position.")
	}

	if got, expected := len(board.EvalMove(BLACK, 2, 3)), 1; got != expected {
		t.Error("This position should be valid.")
	}
	if got, expected := len(board.EvalMove(BLACK, 5, 4)), 1; got != expected {
		t.Error("This position should be valid.")
	}

	// Vertical
	if got, expected := len(board.EvalMove(BLACK, 0, 3)), 0; got != expected {
		t.Error("Cannot move into a isolated position.")
	}

	if got, expected := len(board.EvalMove(BLACK, 4, 2)), 0; got != expected {
		t.Error("Cannot move into a this position.")
	}

	if got, expected := len(board.EvalMove(BLACK, 3, 2)), 1; got != expected {
		t.Error("This position should be valid.")
	}
	if got, expected := len(board.EvalMove(BLACK, 4, 5)), 1; got != expected {
		t.Error("This position should be valid.")
	}
}

func TestCanMoveDiagonal(t *testing.T) {
	board := NewBoard(8, 8)
	board.Init()

	board.board[3][2] = BLACK
	board.board[3][3] = BLACK

	if got, expected := len(board.EvalMove(WHITE, 2, 2)), 1; got != expected {
		t.Error("This position is valid.")
	}
	if got, expected := len(board.EvalMove(BLACK, 5, 5)), 1; got != expected {
		t.Error("This position is valid.")
	}
}

func TestBoard_Count(t *testing.T) {
	board := NewBoard(8, 8)
	board.Init()

	w, b, e := board.Count()
	if w != 2 {
		t.Error("Expecting 2 whites")
	}
	if b != 2 {
		t.Error("Expecting 2 blacks")
	}
	if e != 60 {
		t.Error("Expecting 60 empties")
	}
}

func TestValidMovementsForPlayer(t *testing.T) {
	board := NewBoard(8, 8)
	board.Init()

	inArray := func(p tuple, l []tuple) bool {
		for _, c := range l {
			if c.X == p.X && c.Y == p.Y {
				return true
			}
		}
		return false
	}

	expectedWhites := []tuple{
		tuple{2, 4},
		tuple{3, 5},
		tuple{4, 2},
		tuple{5, 3},
	}
	expectedBlacks := []tuple{
		tuple{2, 3},
		tuple{3, 2},
		tuple{4, 5},
		tuple{5, 4},
	}
	whiteMovs := board.ValidMovementsForPlayer(WHITE)
	for _, mov := range whiteMovs {
		if !inArray(mov, expectedWhites) {
			t.Errorf("[%d,%d] is not a valid movement", mov.X, mov.Y)
		}
	}

	blackMovs := board.ValidMovementsForPlayer(BLACK)
	for _, mov := range blackMovs {
		if !inArray(mov, expectedBlacks) {
			t.Errorf("[%d,%d] is not a valid movement", mov.X, mov.Y)
		}
	}
}

func TestValidMovementsForPlayerDiagonal(t *testing.T) {
	board := NewBoard(8, 8)
	board.Init()

	board.board[0][2] = WHITE
	board.board[2][2] = WHITE

	board.board[1][3] = WHITE
	board.board[2][3] = BLACK
	board.board[3][3] = BLACK
	board.board[4][3] = BLACK

	board.board[2][4] = WHITE
	board.board[3][4] = WHITE
	board.board[4][4] = WHITE

	inArray := func(p tuple, l []tuple) bool {
		for _, c := range l {
			if c.X == p.X && c.Y == p.Y {
				return true
			}
		}
		return false
	}

	expected := []tuple{
		tuple{3, 4},
		tuple{3, 5},
	}

	changes, err := board.Move(BLACK, 3, 5)
	if err != nil {
		t.Error(err)
	}

	for _, mov := range changes {
		if !inArray(mov, expected) {
			t.Errorf("[%d,%d] shouldn't change", mov.X, mov.Y)
		}
	}

}

func TestBoard_Clone(t *testing.T) {
	var i, j int8
	origin := NewBoard(8, 8)
	origin.Init()

	dest := origin.Clone()
	for i = 0; i < origin.Width; i++ {
		for j = 0; j < origin.Height; j++ {
			if origin.board[i][j] != dest.board[i][j] {
				t.Error("Boards differ.")
			}
		}
	}
}

func TestBoard_IsEdge(t *testing.T) {
	var i, j, WIDTH, HEIGHT int8

	HEIGHT = 8
	WIDTH = 8

	b := NewBoard(WIDTH, HEIGHT)
	b.Init()

	var edges = []tuple{{0, 0}, {0, HEIGHT - 1}, {WIDTH - 1, HEIGHT - 1}, {WIDTH - 1, 0}}

	for _, p := range edges {
		if !b.isEdge(p.X, p.Y) {
			t.Errorf("[%d,%d] should be an edge.", p.X, p.Y)
		}
	}

	for i = 0; i < WIDTH; i++ {
		for j = 0; j < HEIGHT; j++ {
			if !inArray(tuple{i, j}, edges) {
				if b.isEdge(i, j) {
					t.Errorf("[%d,%d] shouldn't be an edge.", i, j)

				}
			}
		}
	}
}

func TestBoard_IsSide(t *testing.T) {
	var i, j, WIDTH, HEIGHT int8
	var sides = []tuple{
		{0, 0},
		{1, 0},
		{2, 0},
		{3, 0},
		{4, 0},
		{5, 0},
		{6, 0},
		{7, 0},

		{0, 7},
		{1, 7},
		{2, 7},
		{3, 7},
		{4, 7},
		{5, 7},
		{6, 7},
		{7, 7},

		{0, 0},
		{0, 1},
		{0, 2},
		{0, 3},
		{0, 4},
		{0, 5},
		{0, 6},
		{0, 7},

		{7, 0},
		{7, 1},
		{7, 2},
		{7, 3},
		{7, 4},
		{7, 5},
		{7, 6},
		{7, 7},
	}
	HEIGHT = 8
	WIDTH = 8

	b := NewBoard(WIDTH, HEIGHT)
	b.Init()

	for _, p := range sides {
		if !b.isSide(p.X, p.Y) {
			t.Errorf("[%d,%d] should be an side.", p.X, p.Y)
		}
	}

	for i = 0; i < WIDTH; i++ {
		for j = 0; j < HEIGHT; j++ {
			if !inArray(tuple{i, j}, sides) {
				if b.isEdge(i, j) {
					t.Errorf("[%d,%d] shouldn't be an side.", i, j)

				}
			}
		}
	}
}

func TestBoard_IsNearEdge(t *testing.T) {
	var i, j, WIDTH, HEIGHT int8
	var neardEdge = []tuple{
		{1, 0},
		{1, 1},
		{0, 1},

		{0, 6},
		{1, 6},
		{1, 7},

		{6, 0},
		{6, 1},
		{7, 1},

		{6, 6},
		{6, 7},
		{7, 6},
	}
	HEIGHT = 8
	WIDTH = 8

	b := NewBoard(WIDTH, HEIGHT)
	b.Init()

	for _, p := range neardEdge {
		if !b.isNearEdge(p.X, p.Y) {
			t.Errorf("[%d,%d] should be a near edge.", p.X, p.Y)
		}
	}

	for i = 0; i < WIDTH; i++ {
		for j = 0; j < HEIGHT; j++ {
			if !inArray(tuple{i, j}, neardEdge) {
				if b.isNearEdge(i, j) {
					t.Errorf("[%d,%d] shouldn't be a near edge.", i, j)

				}
			}
		}
	}
}

func Test_ID_And_RestoreFromID(t *testing.T) {
	board := NewBoard(8, 8)
	otherBoard := NewBoard(8, 8)
	board.Init()

	for n := 0; n < 1000000; n++ {
		for x:=0; x<8; x++ {
			for y:=0; y<8; y++ {
				board.board[x][y] = int8(rand.Intn(3))
			}
		}
		otherBoard.RestoreFromID(board.ID())

		for x:=0; x<8; x++ {
			for y:=0; y<8; y++ {
				if board.board[x][y] != otherBoard.board[x][y] {
					t.Fatal("Error recovering board.")
				}
			}
		}
	}
}

func TestGame_Ends(t *testing.T) {
	board := NewBoard(8, 8)
	board.Init()

	for {
		whiteCannotPlay := false
		fmt.Println("WHITE PLAYS:")
		_, err := board.ComputerMove(WHITE)
		if err != nil {
			whiteCannotPlay = true
			fmt.Println(err)

		}
		fmt.Print(string(board.Dump()))
		b, w, e := board.Count()
		fmt.Printf("WHITES[X]:%d, BLACKS[0]:%d, EMPTIES:%d\n\n", w, b, e)

		blackCannotPlay := false
		fmt.Println("BLACK PLAYS:")
		_, err = board.ComputerMoveMinMax(BLACK)
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

func inArray(val tuple, tuples []tuple) bool {
	for _, t := range tuples {
		if val.X == t.X && val.Y == t.Y {
			return true
		}
	}
	return false
}
