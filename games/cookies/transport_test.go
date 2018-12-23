package cookies

import (
	"testing"
	"github.com/stretchr/testify/assert"

	"github.com/x1m3/elixir/games/cookies/codec/json"
	"github.com/x1m3/elixir/games/cookies/codec"
	"github.com/x1m3/elixir/games/cookies/codec/msgpack"
	"github.com/davecgh/go-spew/spew"
)

func TestTransport_MarshalUnmarshall(t *testing.T) {

	msg := &ViewPortRequest{X: 1.0/3, Y: 2.0/3, XX: 3.0/3, YY: 4.0/3}
	msg.SetType(ViewPortRequestType)

	for _, codec := range []codec.MarshalUnmarshaler{json.Codec, msgpack.Codec} {
		transport := NewTransport(codec)

		data, err := transport.Marshal(msg)
		assert.NoError(t, err, codec.Name())

		recData, err := transport.Unmarshal(data)
		assert.NoError(t, err, codec.Name())
		recMsg, ok := recData.(*ViewPortRequest)
		assert.True(t, ok, codec.Name())

		spew.Dump(data)

		assert.Equal(t, msg, recMsg, codec.Name())
	}
}

func BenchmarkTransport_MarshalJson(b *testing.B) {

	msg := &ViewPortRequest{X: 1.0/3, Y: 2.0/3, XX: 3.0/3, YY: 4.0/3}
	msg.SetType(ViewPortRequestType)

	transport := NewTransport(json.Codec)
	for n:=0;n<b.N; n++ {
		transport.Marshal(msg)
	}
}

func BenchmarkTransport_MarshalMsgPack(b *testing.B) {

	msg := &ViewPortRequest{X: 1.0/3, Y: 2.0/3, XX: 3.0/3, YY: 4.0/3}
	msg.SetType(ViewPortRequestType)

	transport := NewTransport(msgpack.Codec)
	for n:=0;n<b.N; n++ {
		transport.Marshal(msg)
	}
}

func BenchmarkTransport_UnMarshalJson(b *testing.B) {

	msg := &ViewPortRequest{X: 1.0/3, Y: 2.0/3, XX: 3.0/3, YY: 4.0/3}
	msg.SetType(ViewPortRequestType)

	transport := NewTransport(json.Codec)

	data, _ := transport.Marshal(msg)

	for n:=0;n<b.N; n++ {
		transport.Unmarshal(data)
	}
}

func BenchmarkTransport_UnMarshalMsgPack(b *testing.B) {

	msg := &ViewPortRequest{X: 1.0/3, Y: 2.0/3, XX: 3.0/3, YY: 4.0/3}
	msg.SetType(ViewPortRequestType)

	transport := NewTransport(msgpack.Codec)

	data, _ := transport.Marshal(msg)

	for n:=0;n<b.N; n++ {
		transport.Unmarshal(data)
	}
}
