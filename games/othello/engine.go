package othello

import (
	"errors"
	"math"
	"bytes"
	"fmt"
)

type tuple struct {
	X int
	Y int
}

type board [WIDTH][HEIGHT]int8

func (b *board) Init() {
	for i := 0; i < WIDTH; i++ {
		for j := 0; j < HEIGHT; j++ {
			b[i][j] = EMPTY
		}
	}
	b[3][3], b[4][4] = WHITE, WHITE
	b[4][3], b[3][4] = BLACK, BLACK
}

func (b *board) Clone() *board {
	var new board
	for i := 0; i < WIDTH; i++ {
		for j := 0; j < HEIGHT; j++ {
			new[i][j] = b[i][j]
		}
	}
	return &new
}

func (b *board) Count() (blacks int, whites int, empties int) {
	for i := 0; i < WIDTH; i++ {
		for j := 0; j < HEIGHT; j++ {
			switch b[i][j] {
			case BLACK:
				blacks++
			case WHITE:
				whites++
			case EMPTY:
				empties++
			}
		}
	}
	return
}

func (b *board) Move(player int8, moveToX int, moveToY int) ([]*tuple, error) {
	changes := b.EvalMove(player, moveToX, moveToY)
	if len(changes) == 0 {
		return nil, errors.New("invalid movement")
	}
	changes = append(changes, &tuple{moveToX, moveToY})
	for _, position := range changes {
		b[position.X][position.Y] = player
	}
	return changes, nil
}

// A very simple and stupid implementation
func (b *board) ComputerMove(player int8) ([]*tuple, error) {
	var candidate *tuple

	validMovements := b.ValidMovementsForPlayer(player)
	if len(validMovements) == 0 {
		return nil, errors.New("player cannot move")
	}
	best := math.MinInt64

	for _, movement := range validMovements {
		board := b.Clone()
		board.Move(player, movement.X, movement.Y)
		heuristic := board.Heuristic(player) //+ rand.Intn(30) - 15
		if best < heuristic {
			best = heuristic
			candidate = movement
		}
	}
	return b.Move(player, candidate.X, candidate.Y)
}

func (b *board) ComputerMoveMinMax(player int8) ([]*tuple, error) {
	_, movement := b.minimax(player, 6, true)
	if movement == nil {
		return nil, errors.New("player cannot move")
	}
	return b.Move(player, movement.X, movement.Y)
}

func (b *board) ValidMovementsForPlayer(player int8) []*tuple {
	movements := make([]*tuple, 0)
	for i := 0; i < WIDTH; i++ {
		for j := 0; j < HEIGHT; j++ {
			if b[i][j] == EMPTY {
				if len(b.EvalMove(player, i, j)) > 0 {
					movements = append(movements, &tuple{X: i, Y: j})
				}
			}
		}
	}
	return movements
}

func (b *board) EvalMove(player int8, moveToX int, moveToY int) []*tuple {
	/*
	if b[moveToX][moveToY] != EMPTY { // Cannot move to an empty position
		return nil
	}
	*/
	eats := make([]*tuple, 0)
	eats = b.evalMoveHoriz(player, eats, moveToX, moveToY)
	eats = b.evalMoveVert(player, eats, moveToX, moveToY)
	eats = b.evalMoveDiagonal(player, eats, moveToX, moveToY)

	return eats
}

func (b *board) evalMoveHoriz(player int8, eats []*tuple, moveToX int, moveToY int) []*tuple {
	var toX int

	for _, deltaX := range []int{-1, 1} {
		toX = moveToX
		for x := moveToX + deltaX; x >= 0 && x < WIDTH; x += deltaX {

			if b[x][moveToY] != player && b[x][moveToY] != EMPTY && !b.outOfLimits(x+deltaX, moveToY) {
				toX = x
			} else {
				if b[x][moveToY] != player {
					toX = moveToX
				}
				break
			}
		}
		if toX != moveToX {
			for x := moveToX + deltaX; x != toX+deltaX; x += deltaX {
				eats = append(eats, &tuple{x, moveToY})
			}
		}
	}
	return eats
}

func (b *board) evalMoveVert(player int8, eats []*tuple, moveToX int, moveToY int) []*tuple {
	var toY int
	for _, deltaY := range []int{-1, 1} {
		toY = moveToY
		for y := moveToY + deltaY; y >= 0 && y < HEIGHT; y += deltaY {
			if b[moveToX][y] != player && b[moveToX][y] != EMPTY && !b.outOfLimits(moveToX, y+deltaY) {
				toY = y
			} else {
				if b[moveToX][y] != player {
					toY = moveToY
				}
				break
			}
		}
		if toY != moveToY {
			for y := moveToY + deltaY; y != toY+deltaY; y += deltaY {
				eats = append(eats, &tuple{moveToX, y})
			}
		}
	}
	return eats
}

func (b *board) evalMoveDiagonal(player int8, eats []*tuple, moveToX int, moveToY int) []*tuple {
	acum := make([]*tuple, 0, 8)
	for _, deltaX := range []int{-1, 1} {
		for _, deltaY := range []int{-1, 1} {
			for x, y := moveToX+deltaX, moveToY+deltaY; x >= 0 && x < WIDTH && y >= 0 && y < HEIGHT; x, y = x+deltaX, y+deltaY {
				if b[x][y] != player && b[x][y] != EMPTY && !b.outOfLimits(x+deltaX, y+deltaY) {
					acum = append(acum, &tuple{X: x, Y: y})
				} else {
					if b[x][y] != player {
						acum = nil
					}
					break
				}
			}
			eats = append(eats, acum...)
		}
	}
	return eats
}

func (b *board) Heuristic(player int8) int {
	heuristic := 0
	value := 0

	for x := 0; x < WIDTH; x++ {
		for y := 0; y < HEIGHT; y++ {
			chip := b[x][y]
			if  chip != EMPTY {
				value = 1
				if b.isSide(x, y) {
					if b.isEdge(x, y) {
						value = 100
					} else if b.isNearEdge(x, y) {
						value = 5
					} else {
						value = 10
					}
				}
			}
			if chip == player {
				heuristic += value
			} else {
				heuristic -= value
			}
		}
	}
	return heuristic
}

func (b *board) HeuristicOnMovements(player int8) int {
	if player == BLACK {
		return len(b.ValidMovementsForPlayer(BLACK)) - len(b.ValidMovementsForPlayer(WHITE))
	} else {
		return len(b.ValidMovementsForPlayer(WHITE)) - len(b.ValidMovementsForPlayer(BLACK))
	}
}

func (b *board) outOfLimits(x, y int) bool {
	return x < 0 || x > WIDTH-1 || y < 0 || y > HEIGHT-1
}

func (b *board) isEdge(x, y int) bool {
	return x == 0 && y == 0 || x == WIDTH-1 && y == 0 || x == WIDTH-1 && y == HEIGHT-1 || x == 0 && y == HEIGHT-1
}

func (b *board) isNearEdge(x, y int) bool {
	for _, deltaX := range []int{-1, 0, 1} {
		for _, deltaY := range []int{-1, 0, 1} {
			if deltaX == 0 && deltaY == 0 {
				continue
			}
			if b.isEdge(x+deltaX, y+deltaY) {
				return true
			}
		}
	}
	return false
}

func (b *board) isSide(x, y int) bool {
	return x == 0 || x == WIDTH-1 || y == 0 || y == HEIGHT-1
}

func (b *board) minimax(player int8, depth int, max bool) (int, *tuple) {
	var bestVal int
	var bestMovement *tuple
	var opositePlayer int8

	if depth == 0 {
		return b.Heuristic(player), nil
	}
	if player == WHITE {
		opositePlayer = BLACK
	} else {
		opositePlayer = WHITE
	}
	if max {
		bestVal = math.MinInt64
		for _, movement := range b.ValidMovementsForPlayer(player) {
			newBoard := b.Clone()
			newBoard.Move(player, movement.X, movement.Y)
			minmax, _ := newBoard.minimax(opositePlayer, depth-1, !max)
			if bestVal < minmax {
				bestVal = minmax
				bestMovement = movement
			}
		}
	} else {
		bestVal = math.MaxInt64
		for _, movement := range b.ValidMovementsForPlayer(player) {
			newBoard := b.Clone()
			newBoard.Move(player, movement.X, movement.Y)
			minmax, _ := newBoard.minimax(opositePlayer, depth-1, !max)
			if bestVal > minmax {
				bestVal = minmax
				bestMovement = movement
			}
		}
	}
	return bestVal, bestMovement
}

func (b *board) Dump() []byte {
	var buf bytes.Buffer
	var c string

	for j := 0; j < HEIGHT; j++ {
		for i := 0; i < WIDTH; i++ {

			switch b[i][j] {
			case BLACK:
				c = "0"
			case WHITE:
				c = "X"
			default:
				c = "."
			}

			buf.WriteString(fmt.Sprintf(" %s ", c))
		}
		buf.WriteString("\n")
	}
	return buf.Bytes()
}
