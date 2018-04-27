package command

const SLOT3_INIT_REQUEST = 1
const SLOT3_INIT_RESPONSE = 2
const SLOT3_SPIN_REQUEST = 3
const SLOT3_SPIN_RESPONSE = 4

type Slot3InitRequest struct {
	request
}

func (r *Slot3InitRequest) Type() int {
	return SLOT3_INIT_REQUEST
}

type Slot3InitResponse struct {
	response
	Wheel1 [24]int8
	Wheel2 [24]int8
	Wheel3 [24]int8
	P1     int8
	P2     int8
	P3     int8
}

func (r *Slot3InitResponse) Type() int {
	return SLOT3_INIT_RESPONSE
}

type Slot3SpinRequest struct {
	request
	Bet int64
}

func (r *Slot3SpinRequest) Type() int {
	return SLOT3_SPIN_REQUEST
}

type Slot3SpinResponse struct {
	response
	Win int64
	P1  int8
	P2  int8
	P3  int8
}

func (r *Slot3SpinResponse) Type() int {
	return SLOT3_SPIN_RESPONSE
}
