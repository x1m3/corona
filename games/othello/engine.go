package othello

import (
	"errors"
	"math"
	"bytes"
	"fmt"
)

const EMPTY = 0
const BLACK = 1
const WHITE = 2

type MovementsCache interface{
	Movements(id []byte, player int8) ([]tuple, bool)
	StoreMovements(player int8, id []byte, movements []tuple)
}

type tuple struct {
	X int8
	Y int8
}

type board struct {
	Width  int8
	Height int8
	board  [][]int8
	movementsCache MovementsCache
}

func NewBoard(width, height int8, cache MovementsCache) *board {
	b := &board{
		Width:  width,
		Height: height,
		movementsCache:cache,
	}
	b.board = make([][]int8, width)
	for i := range b.board {
		b.board[i] = make([]int8, height)
	}
	return b
}

func (b *board) Init() {
	var i, j int8

	for i = 0; i < b.Width; i++ {
		for j = 0; j < b.Height; j++ {
			b.board[i][j] = EMPTY
		}
	}
	b.board[3][3], b.board[4][4] = WHITE, WHITE
	b.board[4][3], b.board[3][4] = BLACK, BLACK
}

func (b *board) Clone() *board {

	newBoard := NewBoard(b.Width, b.Height, b.movementsCache)
	for i := range b.board {
		copy(newBoard.board[i], b.board[i])
	}
	return newBoard
}

func (b *board) ID() []byte {

	id := make([]byte, b.Width*b.Height, b.Width*b.Height)
	for i := range b.board {
		for j, val := range b.board[i] {
			id[i*int(b.Width)+j] = byte(val)
		}
	}
	return id
}

// Please, call Init() first
func (b *board) RestoreFromID(id []byte) {
	for x := range b.board {
		for y := range b.board[x] {
			b.board[x][y] = int8(id[int8(x)*b.Width+int8(y)])
		}
	}
}

func (b *board) Count() (blacks int, whites int, empties int) {
	var i, j int8
	for i = 0; i < b.Width; i++ {
		for j = 0; j < b.Height; j++ {
			switch b.board[i][j] {
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

func (b *board) Move(player int8, moveToX int8, moveToY int8) ([]tuple, error) {
	changes := b.EvalMove(player, moveToX, moveToY)
	if len(changes) == 0 {
		return nil, errors.New("invalid movement")
	}
	changes = append(changes, tuple{moveToX, moveToY})
	for _, position := range changes {
		b.board[position.X][position.Y] = player
	}
	return changes, nil
}

// A very simple and stupid implementation
func (b *board) ComputerMove(player int8) ([]tuple, error) {
	var candidate tuple

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

func (b *board) ComputerMoveMinMax(player int8) ([]tuple, error) {
	var bestMovement tuple
	var oppositePlayer int8
	if player == WHITE {
		oppositePlayer = BLACK
	} else {
		oppositePlayer = WHITE
	}

	bestScore := math.MinInt64

	for _, movement := range b.ValidMovementsForPlayer(player) {
		newBoard := b.Clone()
		newBoard.Move(player, movement.X, movement.Y)
		score := newBoard.minimax(oppositePlayer, 5, false, math.MinInt64, math.MaxInt64)
		if bestScore < score {
			bestScore = score
			bestMovement = movement
		}
	}

	if bestScore == math.MinInt64 {
		return nil, errors.New("player cannot move")
	}
	return b.Move(player, bestMovement.X, bestMovement.Y)
}

func (b *board) ValidMovementsForPlayer(player int8) []tuple {

	boardID := b.ID()

	if movements, found := b.movementsCache.Movements(boardID, player); found {
		return movements
	}

	movements := make([]tuple, 0)
	for i := range b.board {
		for j := range b.board[i] {
			if b.board[i][j] == EMPTY {
				if len(b.EvalMove(player, int8(i), int8(j))) > 0 {
					movements = append(movements, tuple{X: int8(i), Y: int8(j)})
				}
			}
		}
	}

	b.movementsCache.StoreMovements(player, boardID, movements)
	return movements
}

func (b *board) EvalMove(player int8, moveToX int8, moveToY int8) []tuple {
	return b.evalMove(player, moveToX, moveToY)
}

func (b *board) evalMove(player int8, moveToX int8, moveToY int8) []tuple {

	if b.board[moveToX][moveToY] != EMPTY { // Cannot move to an empty position
		return nil
	}

	eats := make([]tuple, 0, 24)
	eats = b.evalMoveHoriz(player, eats, moveToX, moveToY)
	eats = b.evalMoveVert(player, eats, moveToX, moveToY)
	eats = b.evalMoveDiagonal(player, eats, moveToX, moveToY)

	return eats
}


func (b *board) evalMoveHoriz(player int8, eats []tuple, moveToX int8, moveToY int8) []tuple {
	var x, toX, deltaX int8

	for _, deltaX = range []int8{-1, 1} {
		toX = moveToX
		for x = moveToX + deltaX; x >= 0 && x < b.Width; x += deltaX {
			if b.board[x][moveToY] != player && b.board[x][moveToY] != EMPTY && !b.outOfLimits(x+deltaX, moveToY) {
				toX = x
			} else {
				if b.board[x][moveToY] != player {
					toX = moveToX
				}
				break
			}
		}
		if toX != moveToX {
			for x := moveToX + deltaX; x != toX+deltaX; x += deltaX {
				eats = append(eats, tuple{x, moveToY})
			}
		}
	}
	return eats
}

func (b *board) evalMoveVert(player int8, eats []tuple, moveToX int8, moveToY int8) []tuple {
	var y, toY, deltaY int8
	for _, deltaY = range []int8{-1, 1} {
		toY = moveToY
		for y = moveToY + deltaY; y >= 0 && y < b.Height; y += deltaY {
			if b.board[moveToX][y] != player && b.board[moveToX][y] != EMPTY && !b.outOfLimits(moveToX, y+deltaY) {
				toY = y
			} else {
				if b.board[moveToX][y] != player {
					toY = moveToY
				}
				break
			}
		}
		if toY != moveToY {
			for y := moveToY + deltaY; y != toY+deltaY; y += deltaY {
				eats = append(eats, tuple{moveToX, y})
			}
		}
	}
	return eats
}

func (b *board) evalMoveDiagonal(player int8, eats []tuple, moveToX int8, moveToY int8) []tuple {
	var deltaX, deltaY, x, y, toX, toY int8
	for _, deltaX = range []int8{-1, 1} {
		for _, deltaY = range []int8{-1, 1} {
			toX, toY = moveToX, moveToY
			for x, y = moveToX+deltaX, moveToY+deltaY; x >= 0 && x < b.Width && y >= 0 && y < b.Height; x, y = x+deltaX, y+deltaY {
				if b.board[x][y] != player && b.board[x][y] != EMPTY && !b.outOfLimits(x+deltaX, y+deltaY) {
					toX, toY = x, y
				} else {
					if b.board[x][y] != player {
						toX = moveToX
					}
					break
				}
			}
			if toX != moveToX {
				for x, y := moveToX+deltaX, moveToY+deltaY; x != toX+deltaX && y != toY+deltaY; x, y = x+deltaX, y+deltaY {
					eats = append(eats, tuple{x, y})
				}
			}

		}
	}
	return eats
}

func (b *board) Heuristic(player int8) int {
	var x, y int8
	heuristic := 0
	value := 0

	for x = 0; x < b.Width; x++ {
		for y = 0; y < b.Height; y++ {
			chip := b.board[x][y]
			if chip != EMPTY {
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

func (b *board) outOfLimits(x, y int8) bool {
	return x < 0 || x > b.Width-1 || y < 0 || y > b.Height-1
}

func (b *board) isEdge(x, y int8) bool {
	return x == 0 && y == 0 || x == b.Width-1 && y == 0 || x == b.Width-1 && y == b.Height-1 || x == 0 && y == b.Height-1
}

func (b *board) isNearEdge(x, y int8) bool {
	var deltaX, deltaY int8
	for _, deltaX = range []int8{-1, 0, 1} {
		for _, deltaY = range []int8{-1, 0, 1} {
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

func (b *board) isSide(x, y int8) bool {
	return x == 0 || x == b.Width-1 || y == 0 || y == b.Height-1
}

func (b *board) minimax(player int8, depth int8, maxMin bool, alpha int, beta int) int {
	var opositePlayer int8

	if depth == 0 {
		return b.Heuristic(player)
	}
	if player == WHITE {
		opositePlayer = BLACK
	} else {
		opositePlayer = WHITE
	}
	if maxMin {
		for _, movement := range b.ValidMovementsForPlayer(player) {
			newBoard := b.Clone()
			newBoard.Move(player, movement.X, movement.Y)
			minmax := newBoard.minimax(opositePlayer, depth-1, !maxMin, alpha, beta)
			alpha = max(alpha, minmax)
			if beta <= alpha {
				break
			}
		}
		return alpha
	} else {
		for _, movement := range b.ValidMovementsForPlayer(player) {
			newBoard := b.Clone()
			newBoard.Move(player, movement.X, movement.Y)
			minmax := newBoard.minimax(opositePlayer, depth-1, !maxMin, alpha, beta)
			beta = min(beta, minmax)
			if beta <= alpha {
				break
			}
		}
		return beta
	}
}

func (b *board) Dump() []byte {
	var buf bytes.Buffer
	var c string
	var i, j int8

	for j = 0; j < b.Height; j++ {
		for i = 0; i < b.Width; i++ {

			switch b.board[i][j] {
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

func max(x, y int) int {
	if x > y {
		return x
	} else {
		return y
	}
}

func min(x, y int) int {
	if x <= y {
		return x
	} else {
		return y
	}
}
