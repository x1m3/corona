package corona_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/davecgh/go-spew/spew"

	"github.com/x1m3/corona/internal/codec"
	"github.com/x1m3/corona/internal/codec/json"
	"github.com/x1m3/corona/internal/codec/msgpack"
	"github.com/x1m3/corona/internal/corona"
	"github.com/x1m3/corona/internal/messages"
)

type dummyConnection struct {
	msgs [][]byte
}

func (c *dummyConnection) Close() error {
	panic("implement me")
}

func (c *dummyConnection) WriteMessage(data []byte) error {
	c.msgs = append(c.msgs, data)
	return nil
}

func (c *dummyConnection) ReadMessage() (p []byte, err error) {
	item := c.msgs[len(c.msgs)-1]
	c.msgs = c.msgs[0 : len(c.msgs)-1]
	return item, nil
}

func TestTransport_MarshalUnmarshall(t *testing.T) {

	msg := &messages.ViewPortRequest{X: 1.0 / 3, Y: 2.0 / 3, XX: 3.0 / 3, YY: 4.0 / 3}
	msg.SetType(messages.ViewPortRequestType)

	for _, codec := range []codec.MarshalUnmarshaler{json.Codec, msgpack.Codec} {
		transport := corona.NewTransport(codec, &dummyConnection{})

		err := transport.Send(msg)
		assert.NoError(t, err, codec.Name())

		recData, err := transport.Receive()
		assert.NoError(t, err, codec.Name())
		recMsg, ok := recData.(*messages.ViewPortRequest)
		assert.True(t, ok, codec.Name())

		spew.Dump(recData)

		assert.Equal(t, msg, recMsg, codec.Name())
	}
}

func BenchmarkTransport_MarshalJson(b *testing.B) {

	msg := &messages.ViewPortRequest{X: 1.0 / 3, Y: 2.0 / 3, XX: 3.0 / 3, YY: 4.0 / 3}
	msg.SetType(messages.ViewPortRequestType)

	transport := corona.NewTransport(json.Codec, &dummyConnection{})
	for n := 0; n < b.N; n++ {
		transport.Send(msg)
	}
}

func BenchmarkTransport_MarshalMsgPack(b *testing.B) {

	msg := &messages.ViewPortRequest{X: 1.0 / 3, Y: 2.0 / 3, XX: 3.0 / 3, YY: 4.0 / 3}
	msg.SetType(messages.ViewPortRequestType)

	transport := corona.NewTransport(msgpack.Codec, &dummyConnection{})
	for n := 0; n < b.N; n++ {
		transport.Send(msg)
	}
}
