package tcp_test

import (
	"github.com/stretchr/testify/assert"
	"net"
	"tcp"
	"testing"
)

func TestEncodeDecodeMessage_String(t *testing.T) {
	msgs := []tcp.Message{
		{
			tcp.Header{
				Version:       tcp.V1,
				Flags:         tcp.FlagError | tcp.FlagHuffman,
				TransactionID: tcp.TransactionID(make([]byte, 32)),
			},
			"Hello",
		},
		{
			tcp.Header{
				Version:       tcp.V1,
				Flags:         tcp.FlagError | tcp.FlagHuffman,
				TransactionID: tcp.TransactionID(make([]byte, 32)),
			},
			"World",
		},
	}
	testEncodeDecodeMessages(t, msgs)
}

func testEncodeDecodeMessages(t *testing.T, msgs []tcp.Message) {
	for _, msg := range msgs {
		t.Run(msg.Body.(string), func(t *testing.T) {
			testEncodeMessage(t, msg)
		})
	}
}

func testEncodeMessage(t *testing.T, msg tcp.Message) {
	client, server := net.Pipe()
	// Write to server
	go func() {
		encoder := tcp.NewEncoder(server)
		err := encoder.Encode(&msg)
		assert.NoError(t, err)
		server.Close()
	}()

	// Read from client
	decoder := tcp.NewDecoder(client)
	var res tcp.Message
	err := decoder.Decode(&res)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, msg, res)
	client.Close()
}
