package cookies

type throwFoodTask struct {
	count int
	x     float64
	y     float64
}

func newThrowFoodTask(count int, x float64, y float64) *throwFoodTask {
	return &throwFoodTask{count: count, x: x, y: y}
}
