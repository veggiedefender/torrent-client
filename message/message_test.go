package message

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatRequest(t *testing.T) {
	msg := FormatRequest(4, 567, 4321)
	expected := &Message{
		ID: MsgRequest,
		Payload: []byte{
			0x00, 0x00, 0x00, 0x04, // Index
			0x00, 0x00, 0x02, 0x37, // Begin
			0x00, 0x00, 0x10, 0xe1, // Length
		},
	}
	assert.Equal(t, expected, msg)
}

func TestParsePiece(t *testing.T) {
	tests := map[string]struct {
		inputIndex int
		inputBuf   []byte
		inputMsg   *Message
		outputN    int
		outputBuf  []byte
		fails      bool
	}{
		"parse valid piece": {
			inputIndex: 4,
			inputBuf:   make([]byte, 10),
			inputMsg: &Message{
				ID: MsgPiece,
				Payload: []byte{
					0x00, 0x00, 0x00, 0x04, // Index
					0x00, 0x00, 0x00, 0x02, // Begin
					0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, // Block
				},
			},
			outputBuf: []byte{0x00, 0x00, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x00},
			outputN:   6,
			fails:     false,
		},
		"wrong message type": {
			inputIndex: 4,
			inputBuf:   make([]byte, 10),
			inputMsg: &Message{
				ID:      MsgChoke,
				Payload: []byte{},
			},
			outputBuf: make([]byte, 10),
			outputN:   0,
			fails:     true,
		},
		"payload too short": {
			inputIndex: 4,
			inputBuf:   make([]byte, 10),
			inputMsg: &Message{
				ID: MsgPiece,
				Payload: []byte{
					0x00, 0x00, 0x00, 0x04, // Index
					0x00, 0x00, 0x00, // Malformed offset
				},
			},
			outputBuf: make([]byte, 10),
			outputN:   0,
			fails:     true,
		},
		"wrong index": {
			inputIndex: 4,
			inputBuf:   make([]byte, 10),
			inputMsg: &Message{
				ID: MsgPiece,
				Payload: []byte{
					0x00, 0x00, 0x00, 0x06, // Index is 6, not 4
					0x00, 0x00, 0x00, 0x02, // Begin
					0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, // Block
				},
			},
			outputBuf: make([]byte, 10),
			outputN:   0,
			fails:     true,
		},
		"offset too high": {
			inputIndex: 4,
			inputBuf:   make([]byte, 10),
			inputMsg: &Message{
				ID: MsgPiece,
				Payload: []byte{
					0x00, 0x00, 0x00, 0x04, // Index is 6, not 4
					0x00, 0x00, 0x00, 0x0c, // Begin is 12 > 10
					0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, // Block
				},
			},
			outputBuf: make([]byte, 10),
			outputN:   0,
			fails:     true,
		},
		"offset ok but payload too long": {
			inputIndex: 4,
			inputBuf:   make([]byte, 10),
			inputMsg: &Message{
				ID: MsgPiece,
				Payload: []byte{
					0x00, 0x00, 0x00, 0x04, // Index is 6, not 4
					0x00, 0x00, 0x00, 0x02, // Begin is ok
					// Block is 10 long but begin=2; too long for input buffer
					0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x0a, 0x0b, 0x0c, 0x0d,
				},
			},
			outputBuf: make([]byte, 10),
			outputN:   0,
			fails:     true,
		},
	}

	for _, test := range tests {
		n, err := ParsePiece(test.inputIndex, test.inputBuf, test.inputMsg)
		if test.fails {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, test.outputBuf, test.inputBuf)
		assert.Equal(t, test.outputN, n)
	}
}

func TestMessageSerialize(t *testing.T) {
	tests := map[string]struct {
		input  *Message
		output []byte
	}{
		"serialize message": {
			input:  &Message{ID: MsgHave, Payload: []byte{1, 2, 3, 4}},
			output: []byte{0, 0, 0, 5, 4, 1, 2, 3, 4},
		},
		"serialize keep-alive": {
			input:  nil,
			output: []byte{0, 0, 0, 0},
		},
	}

	for _, test := range tests {
		buf := test.input.Serialize()
		assert.Equal(t, test.output, buf)
	}
}

func TestReadMessage(t *testing.T) {
	tests := map[string]struct {
		input  []byte
		output *Message
		fails  bool
	}{
		"parse normal message into struct": {
			input:  []byte{0, 0, 0, 5, 4, 1, 2, 3, 4},
			output: &Message{ID: MsgHave, Payload: []byte{1, 2, 3, 4}},
			fails:  false,
		},
		"parse keep-alive into nil": {
			input:  []byte{0, 0, 0, 0},
			output: nil,
			fails:  false,
		},
		"length too short": {
			input:  []byte{1, 2, 3},
			output: nil,
			fails:  true,
		},
		"buffer too short for length": {
			input:  []byte{0, 0, 0, 5, 4, 1, 2},
			output: nil,
			fails:  true,
		},
	}

	for _, test := range tests {
		reader := bytes.NewReader(test.input)
		m, err := ReadMessage(reader)
		if test.fails {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, test.output, m)
	}
}

func TestMessageString(t *testing.T) {
	tests := []struct {
		input  *Message
		output string
	}{
		{nil, "KeepAlive"},
		{&Message{MsgChoke, []byte{1, 2, 3}}, "Choke\t[01 02 03]"},
		{&Message{MsgUnchoke, []byte{1, 2, 3}}, "Unchoke\t[01 02 03]"},
		{&Message{MsgInterested, []byte{1, 2, 3}}, "Interested\t[01 02 03]"},
		{&Message{MsgNotInterested, []byte{1, 2, 3}}, "NotInterested\t[01 02 03]"},
		{&Message{MsgHave, []byte{1, 2, 3}}, "Have\t[01 02 03]"},
		{&Message{MsgBitfield, []byte{1, 2, 3}}, "Bitfield\t[01 02 03]"},
		{&Message{MsgRequest, []byte{1, 2, 3}}, "Request\t[01 02 03]"},
		{&Message{MsgPiece, []byte{1, 2, 3}}, "Piece\t[01 02 03]"},
		{&Message{MsgCancel, []byte{1, 2, 3}}, "Cancel\t[01 02 03]"},
		{&Message{MsgPort, []byte{1, 2, 3}}, "Port\t[01 02 03]"},
		{&Message{99, []byte{1, 2, 3}}, "Unknown#99\t[01 02 03]"},
	}

	for _, test := range tests {
		s := test.input.String()
		assert.Equal(t, test.output, s)
	}
}
