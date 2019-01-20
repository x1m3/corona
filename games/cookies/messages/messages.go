package messages

import (
	"encoding/json"
)

type msgType int8

const (
	ViewPortRequestType      = 1
	ViewPortResponseType     = 2
	UserJoinRequestType      = 3
	UserJoinResponseType     = 4
	CreateCookieRequestType  = 5
	CreateCookieResponseType = 6
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
	X     float32 `json:"X"`
	Y     float32 `json:"Y"`
	XX    float32 `json:"XX"`
	YY    float32 `json:"YY"`
	Angle float32 `json:"R"`
	Turbo bool    `json:"T"`
}

type ViewportResponse struct {
	BaseMessage
	Cookies []*CookieInfo `json:"C"`
}

type CookieInfo struct {
	ID              uint64  `json:"ID"`
	Score           uint64  `json:"SC"`
	X               float32 `json:"X"`
	Y               float32 `json:"Y"`
	AngularVelocity float32 `json:"AV"`
}

type CreateCookieResponse struct {
	BaseMessage
	Data CookieInfo `json:"d"`
}

func NewCreateCookieResponse(ID uint64, sc uint64, X float32, Y float32, AngularVelocity float32) *CreateCookieResponse {
	resp := &CreateCookieResponse{
		Data: CookieInfo{
			ID:              ID,
			Score:           sc,
			X:               X,
			Y:               Y,
			AngularVelocity: AngularVelocity,
		},
	}
	resp.SetType(CreateCookieResponseType)
	return resp
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

type CreateCookieRequest struct {
	BaseMessage
}
