package ants

type ViewPortRequest struct {
	X  float64
	Y  float64
	XX float64
	YY float64
}

func (r *ViewPortRequest) Request() {}

func (r *ViewPortRequest) Type() int {
	return 1
}

type antResponseDTO struct {
	ID int
	X  float64
	Y  float64
/*
	Vx float64
	Vy float64
*/
	R  float64
}

type ViewportResponse struct {
	Ants []antResponseDTO
}

func (r *ViewportResponse) Response() {}

func (r *ViewportResponse) Type() int {
	return 2
}
