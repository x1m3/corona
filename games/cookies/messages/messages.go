package messages

import (
	"encoding/json"
)

type msgType int8

const (
	ViewPortRequestType  = 1
	ViewPortResponseType = 2
	UserJoinRequestType  = 3
	UserJoinResponseType = 4
)

type Message interface {
	GetType() msgType
	SetType(msgType)
}

type BaseMessage struct {
	Type msgType         `json:"t"`
	Data json.RawMessage `json:"d,omitempty"`
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
	ID              int     `json:"ID"`
	Score           int     `json:"SC"`
	X               float64 `json:"X"`
	Y               float64 `json:"Y"`
	AngularVelocity float64 `json:"AV"`
}

type UserJoinRequest struct {
	BaseMessage
	Username string `json:"UN"`
}

type userJoinResponseData struct {
	Ok       bool     `json:"OK"`
	AltNames []string `json:"AN"`
}

type UserJoinResponse struct {
	BaseMessage
	Data userJoinResponseData `json:"d"`
}

func NewUserJoinResponse(ok bool, altNames []string) *UserJoinResponse {
	resp := &UserJoinResponse{Data: userJoinResponseData{Ok: ok, AltNames: altNames}}
	resp.SetType(UserJoinResponseType)
	return resp
}
