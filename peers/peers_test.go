package peers

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshal(t *testing.T) {
	tests := map[string]struct {
		input  string
		output []Peer
		fails  bool
	}{
		"correctly parses peers": {
			input: string([]byte{127, 0, 0, 1, 0x00, 0x50, 1, 1, 1, 1, 0x01, 0xbb}),
			output: []Peer{
				{IP: net.IP{127, 0, 0, 1}, Port: 80},
				{IP: net.IP{1, 1, 1, 1}, Port: 443},
			},
		},
		"not enough bytes in peers": {
			input:  string([]byte{127, 0, 0, 1, 0x00}),
			output: nil,
			fails:  true,
		},
	}

	for _, test := range tests {
		peers, err := Unmarshal([]byte(test.input))
		if test.fails {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, test.output, peers)
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		input  Peer
		output string
	}{
		{
			input:  Peer{IP: net.IP{127, 0, 0, 1}, Port: 8080},
			output: "127.0.0.1:8080",
		},
	}
	for _, test := range tests {
		s := test.input.String()
		assert.Equal(t, test.output, s)
	}
}
