package othello

import (
	"errors"
)

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

type tuple struct {
	X int
	Y int
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
	if b[moveToX][moveToY] != EMPTY { // Cannot move to an empty position
		return nil
	}
	eats := make([]*tuple, 0)
	eats = append(eats, b.evalMoveHoriz(player, moveToX, moveToY)...)
	eats = append(eats, b.evalMoveVert(player, moveToX, moveToY)...)
	eats = append(eats, b.evalMoveDiagonal(player, moveToX, moveToY)...)

	return eats
}

func (b *board) evalMoveHoriz(player int8, moveToX int, moveToY int) []*tuple {
	eats := make([]*tuple, 0)
	for _, deltaX := range []int{-1, 1} {
		acum := make([]*tuple, 0)
		for x := moveToX + deltaX; x >= 0 && x < WIDTH; x += deltaX {
			if b[x][moveToY] != player && b[x][moveToY] != EMPTY {
				acum = append(acum, &tuple{X: x, Y: moveToY})
			} else {
				if b[x][moveToY] == EMPTY {
					acum = nil
				}
				break
			}
		}
		eats = append(eats, acum...)
	}
	return eats
}

func (b *board) evalMoveVert(player int8, moveToX int, moveToY int) []*tuple {
	eats := make([]*tuple, 0)
	for _, deltaY := range []int{-1, 1} {
		acum := make([]*tuple, 0)
		for y := moveToY + deltaY; y >= 0 && y < HEIGHT; y += deltaY {
			if b[moveToX][y] != player && b[moveToX][y] != EMPTY {
				acum = append(acum, &tuple{X: moveToX, Y: y})
			} else {
				if b[moveToX][y] == EMPTY {
					acum = nil
				}
				break
			}
		}
		eats = append(eats, acum...)
	}
	return eats
}

func (b *board) evalMoveDiagonal(player int8, moveToX int, moveToY int) []*tuple {
	eats := make([]*tuple, 0)
	for _, deltaX := range []int{-1, 1} {
		for _, deltaY := range []int{-1, 1} {
			acum := make([]*tuple, 0)
			for x, y := moveToX+deltaX, moveToY+deltaY; x >= 0 && x < WIDTH && y >= 0 && y < HEIGHT; x, y = x+deltaX, y+deltaY {
				if b[x][y] != player && b[x][y] != EMPTY {
					acum = append(acum, &tuple{X: x, Y: y})
				} else {
					if b[x][y] == EMPTY {
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
