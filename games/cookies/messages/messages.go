package messages

import (
	"encoding/json"
)

type msgType int8

const (
	ViewPortRequestType     = 1
	ViewPortResponseType    = 2
	UserJoinRequestType     = 3
	UserJoinResponseType    = 4
	CreateCookieRequestType = 5
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
	Ants []*CookieInfoResponse
}

type cookieInfo struct {
	ID              int     `json:"ID"`
	Score           int     `json:"SC"`
	X               float32 `json:"X"`
	Y               float32 `json:"Y"`
	AngularVelocity float32 `json:"AV"`
}

type CookieInfoResponse struct {
	BaseMessage
	Data cookieInfo `json:"d"`
}

func NewCookieInfoResponse(ID int, sc int, X float32, Y float32, AngularVelocity float32) *CookieInfoResponse {
	resp := &CookieInfoResponse{
		Data : cookieInfo{
			ID:              ID,
			Score:           sc,
			X:               X,
			Y:               Y,
			AngularVelocity: AngularVelocity,
		},
	}
	resp.SetType(ViewPortResponseType)
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
