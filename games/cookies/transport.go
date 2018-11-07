package cookies

import (
	"fmt"
	"encoding/json"
)

type Encoder interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

type JSONEncoder struct{}

func (e *JSONEncoder) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (e *JSONEncoder) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

type Transport struct {
	e Encoder
}

func NewTransport(e Encoder) *Transport {

	return &Transport{e: e}
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
