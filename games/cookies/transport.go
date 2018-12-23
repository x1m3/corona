package cookies

import (
	"fmt"
	"github.com/x1m3/elixir/games/cookies/codec"
	"github.com/gorilla/websocket"
)

type Transport struct {
	conn *websocket.Conn
	e codec.MarshalUnmarshaler
}

func NewTransport(e codec.MarshalUnmarshaler, c *websocket.Conn) *Transport {

	return &Transport{e: e, conn: c}
}

func(t *Transport) Send(data Message) error {
	panic("implement me")
}

func(t *Transport) Receive() (Message, error) {
	panic("implement me")
}

func (t *Transport) Marshal(data Message) ([]byte, error) {
	return t.e.Marshal(data)
}

func (t *Transport) Unmarshal(data []byte) (interface{}, error) {
	var msg Message
	var baseMsg BaseMessage

	if err := t.e.Unmarshal(data, &baseMsg); err != nil {
		return nil, err
	}
	switch baseMsg.GetType() {
	case ViewPortRequestType:
		msg = &ViewPortRequest{}
	case ViewPortResponseType:
		msg = &ViewportResponse{}
	case UserJoinRequestType:
		msg = &UserJoinRequest{}
	default:
		return nil, fmt.Errorf("unknown message type <%s>", baseMsg.GetType())
	}

	if err := t.e.Unmarshal(data, msg); err != nil {
		return nil, err
	}
	msg.SetType(baseMsg.GetType())
	return msg, nil
}
