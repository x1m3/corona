package cookies

type msgType int8

const (
	ViewPortRequestType  = 1
	ViewPortResponseType = 2
	UserJoinRequestType  = 3
)

type Message interface {
	GetType() msgType
	SetType(msgType)
}

type BaseMessage struct {
	Type msgType `json:"t"`
}

func (m *BaseMessage) GetType() msgType {
	return m.Type
}

func (m *BaseMessage) SetType(t msgType) {
	m.Type = t
}

type ViewPortRequest struct {
	BaseMessage
	X  float32
	Y  float32
	XX float32
	YY float32
}

type ViewportResponse struct {
	BaseMessage
	Ants []CookieInfoResponse
}

type CookieInfoResponse struct {
	BaseMessage
	ID              int
	Score           int     `json:"SC"`
	X               float64 `json:"X"`
	Y               float64 `json:"Y"`
	AngularVelocity float64 `json:"AV"`
}

type UserJoinRequest struct {
	BaseMessage
	Username string `json:"UN"`
}
