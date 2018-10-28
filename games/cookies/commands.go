package cookies

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
	SC int64 // Score
	X  float64
	Y  float64
	AV float64 // Angular velocity

}

type ViewportResponse struct {
	Ants []antResponseDTO
}

func (r *ViewportResponse) Response() {}

func (r *ViewportResponse) Type() int {
	return 2
}
