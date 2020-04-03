package cookies

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/gorilla/websocket"

	"github.com/x1m3/corona/games/cookies/codec"
	"github.com/x1m3/corona/games/cookies/messages"
)

type connection interface {
	io.Closer
	WriteMessage(data []byte) error
	ReadMessage() (p []byte, err error)
}

type WebsocketConnection struct {
	messageType int
	conn        *websocket.Conn
}

func NewWebsocketConnection(c *websocket.Conn) *WebsocketConnection {
	return &WebsocketConnection{conn: c, messageType: websocket.TextMessage}
}

func (c *WebsocketConnection) Close() error {
	return c.conn.Close()
}

func (c *WebsocketConnection) WriteMessage(data []byte) error {
	_ = c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	return c.conn.WriteMessage(c.messageType, data)
}

func (c *WebsocketConnection) ReadMessage() (p []byte, err error) {
	tMsg, data, err := c.conn.ReadMessage()
	if err != nil {
		return nil, errors.New("bad connection")
	}
	if tMsg != c.messageType {
		return nil, nil
	}
	return data, err
}

type Transport struct {
	conn connection
	e    codec.MarshalUnmarshaler
}

func NewTransport(e codec.MarshalUnmarshaler, c connection) *Transport {
	return &Transport{e: e, conn: c}
}

func (t *Transport) Send(msg messages.Message) error {
	data, err := t.marshal(msg)
	if err != nil {
		return err
	}

	return t.conn.WriteMessage(data)
}

func (t *Transport) Receive() (messages.Message, error) {
	for {

		data, err := t.conn.ReadMessage()
		if err != nil {
			return nil, errors.New("bad connection")
		}

		return t.unmarshal(data)
	}
}

func (t *Transport) Close() error {
	return t.conn.Close()
}

func (t *Transport) marshal(data messages.Message) ([]byte, error) {
	return t.e.Marshal(data)
}

func (t *Transport) unmarshal(data []byte) (messages.Message, error) {
	var msg messages.Message
	var baseMsg messages.BaseMessage

	if err := t.e.Unmarshal(data, &baseMsg); err != nil {
		return nil, err
	}

	msgType := baseMsg.GetType()
	switch msgType {
	case messages.ViewPortRequestType:
		msg = &messages.ViewPortRequest{}
	case messages.ViewPortResponseType:
		msg = &messages.ViewportResponse{}
	case messages.UserJoinRequestType:
		msg = &messages.UserJoinRequest{}
	case messages.CreateCookieRequestType:
		msg = &messages.CreateCookieRequest{}
	default:
		return nil, fmt.Errorf("unknown message type <%v>", baseMsg.GetType())
	}

	if err := t.e.Unmarshal(baseMsg.Data, msg); err != nil {
		return nil, err
	}
	msg.SetType(msgType)
	return msg, nil
}
