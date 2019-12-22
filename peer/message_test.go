package peer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSerialize(t *testing.T) {
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

func TestRead(t *testing.T) {
	tests := map[string]struct {
		input  []byte
		output *Message
		fails  bool
	}{
		"parse normal message intro struct": {
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
		m, err := Read(reader)
		if test.fails {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, test.output, m)
	}
}

func TestString(t *testing.T) {
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
