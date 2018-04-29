package othello

import (
	"testing"
)

func TestCanMove(t *testing.T) {
	var board board
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
	var board board
	board.Init()
	board[3][2] = BLACK
	board[3][3] = BLACK

	if got, expected := len(board.EvalMove(WHITE, 2, 2)), 1; got != expected {
		t.Error("This position is valid.")
	}
	if got, expected := len(board.EvalMove(BLACK, 5, 5)), 1; got != expected {
		t.Error("This position is valid.")
	}
}

func TestBoard_Count(t *testing.T) {
	var board board
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
	var board board
	board.Init()

	inArray := func(p *tuple, l []*tuple) bool {
		for _, c := range l {
			if c.X == p.X && c.Y == p.Y {
				return true
			}
		}
		return false
	}

	expectedWhites := []*tuple{
		&tuple{2, 4},
		&tuple{3, 5},
		&tuple{4, 2},
		&tuple{5, 3},
	}
	expectedBlacks := []*tuple{
		&tuple{2, 3},
		&tuple{3, 2},
		&tuple{4, 5},
		&tuple{5, 4},
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
