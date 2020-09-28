package client

import (
	"net"
	"testing"

	"github.com/veggiedefender/torrent-client/bitfield"
	"github.com/veggiedefender/torrent-client/handshake"

	"github.com/veggiedefender/torrent-client/message"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createClientAndServer(t *testing.T) (clientConn, serverConn net.Conn) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.Nil(t, err)

	// net.Dial does not block, so we need this signalling channel to make sure
	// we don't return before serverConn is ready
	done := make(chan struct{})
	go func() {
		defer ln.Close()
		serverConn, err = ln.Accept()
		require.Nil(t, err)
		done <- struct{}{}
	}()
	clientConn, err = net.Dial("tcp", ln.Addr().String())
	<-done

	return clientConn, serverConn
}

func TestRecvBitfield(t *testing.T) {
	tests := map[string]struct {
		msg    []byte
		output bitfield.Bitfield
		fails  bool
	}{
		"successful bitfield": {
			msg:    []byte{0x00, 0x00, 0x00, 0x06, 5, 1, 2, 3, 4, 5},
			output: bitfield.Bitfield{1, 2, 3, 4, 5},
			fails:  false,
		},
		"message is not a bitfield": {
			msg:    []byte{0x00, 0x00, 0x00, 0x06, 99, 1, 2, 3, 4, 5},
			output: nil,
			fails:  true,
		},
		"message is keep-alive": {
			msg:    []byte{0x00, 0x00, 0x00, 0x00},
			output: nil,
			fails:  true,
		},
	}

	for _, test := range tests {
		clientConn, serverConn := createClientAndServer(t)
		serverConn.Write(test.msg)

		bf, err := recvBitfield(clientConn)

		if test.fails {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
			assert.Equal(t, bf, test.output)
		}
	}
}

func TestCompleteHandshake(t *testing.T) {
	tests := map[string]struct {
		clientInfohash  [20]byte
		clientPeerID    [20]byte
		serverHandshake []byte
		output          *handshake.Handshake
		fails           bool
	}{
		"successful handshake": {
			clientInfohash:  [20]byte{134, 212, 200, 0, 36, 164, 105, 190, 76, 80, 188, 90, 16, 44, 247, 23, 128, 49, 0, 116},
			clientPeerID:    [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			serverHandshake: []byte{19, 66, 105, 116, 84, 111, 114, 114, 101, 110, 116, 32, 112, 114, 111, 116, 111, 99, 111, 108, 0, 0, 0, 0, 0, 0, 0, 0, 134, 212, 200, 0, 36, 164, 105, 190, 76, 80, 188, 90, 16, 44, 247, 23, 128, 49, 0, 116, 45, 83, 89, 48, 48, 49, 48, 45, 192, 125, 147, 203, 136, 32, 59, 180, 253, 168, 193, 19},
			output: &handshake.Handshake{
				Pstr:     "BitTorrent protocol",
				InfoHash: [20]byte{134, 212, 200, 0, 36, 164, 105, 190, 76, 80, 188, 90, 16, 44, 247, 23, 128, 49, 0, 116},
				PeerID:   [20]byte{45, 83, 89, 48, 48, 49, 48, 45, 192, 125, 147, 203, 136, 32, 59, 180, 253, 168, 193, 19},
			},
			fails: false,
		},
		"wrong infohash": {
			clientInfohash:  [20]byte{134, 212, 200, 0, 36, 164, 105, 190, 76, 80, 188, 90, 16, 44, 247, 23, 128, 49, 0, 116},
			clientPeerID:    [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			serverHandshake: []byte{19, 66, 105, 116, 84, 111, 114, 114, 101, 110, 116, 32, 112, 114, 111, 116, 111, 99, 111, 108, 0, 0, 0, 0, 0, 0, 0, 0, 0xde, 0xe8, 0x6a, 0x7f, 0xa6, 0xf2, 0x86, 0xa9, 0xd7, 0x4c, 0x36, 0x20, 0x14, 0x61, 0x6a, 0x0f, 0xf5, 0xe4, 0x84, 0x3d, 45, 83, 89, 48, 48, 49, 48, 45, 192, 125, 147, 203, 136, 32, 59, 180, 253, 168, 193, 19},
			output:          nil,
			fails:           true,
		},
	}

	for _, test := range tests {
		clientConn, serverConn := createClientAndServer(t)
		serverConn.Write(test.serverHandshake)

		h, err := completeHandshake(clientConn, test.clientInfohash, test.clientPeerID)

		if test.fails {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
			assert.Equal(t, h, test.output)
		}
	}
}

func TestRead(t *testing.T) {
	clientConn, serverConn := createClientAndServer(t)
	client := Client{Conn: clientConn}

	msgBytes := []byte{
		0x00, 0x00, 0x00, 0x05,
		4,
		0x00, 0x00, 0x05, 0x3c,
	}
	expected := &message.Message{
		ID:      message.MsgHave,
		Payload: []byte{0x00, 0x00, 0x05, 0x3c},
	}
	_, err := serverConn.Write(msgBytes)
	require.Nil(t, err)

	msg, err := client.Read()
	assert.Equal(t, expected, msg)
}

func TestSendRequest(t *testing.T) {
	clientConn, serverConn := createClientAndServer(t)
	client := Client{Conn: clientConn}
	err := client.SendRequest(1, 2, 3)
	assert.Nil(t, err)
	expected := []byte{
		0x00, 0x00, 0x00, 0x0d,
		6,
		0x00, 0x00, 0x00, 0x01,
		0x00, 0x00, 0x00, 0x02,
		0x00, 0x00, 0x00, 0x03,
	}
	buf := make([]byte, len(expected))
	_, err = serverConn.Read(buf)
	assert.Nil(t, err)
	assert.Equal(t, expected, buf)
}

func TestSendInterested(t *testing.T) {
	clientConn, serverConn := createClientAndServer(t)
	client := Client{Conn: clientConn}
	err := client.SendInterested()
	assert.Nil(t, err)
	expected := []byte{
		0x00, 0x00, 0x00, 0x01,
		2,
	}
	buf := make([]byte, len(expected))
	_, err = serverConn.Read(buf)
	assert.Nil(t, err)
	assert.Equal(t, expected, buf)
}

func TestSendNotInterested(t *testing.T) {
	clientConn, serverConn := createClientAndServer(t)
	client := Client{Conn: clientConn}
	err := client.SendNotInterested()
	assert.Nil(t, err)
	expected := []byte{
		0x00, 0x00, 0x00, 0x01,
		3,
	}
	buf := make([]byte, len(expected))
	_, err = serverConn.Read(buf)
	assert.Nil(t, err)
	assert.Equal(t, expected, buf)
}

func TestSendUnchoke(t *testing.T) {
	clientConn, serverConn := createClientAndServer(t)
	client := Client{Conn: clientConn}
	err := client.SendUnchoke()
	assert.Nil(t, err)
	expected := []byte{
		0x00, 0x00, 0x00, 0x01,
		1,
	}
	buf := make([]byte, len(expected))
	_, err = serverConn.Read(buf)
	assert.Nil(t, err)
	assert.Equal(t, expected, buf)
}

func TestSendHave(t *testing.T) {
	clientConn, serverConn := createClientAndServer(t)
	client := Client{Conn: clientConn}
	err := client.SendHave(1340)
	assert.Nil(t, err)
	expected := []byte{
		0x00, 0x00, 0x00, 0x05,
		4,
		0x00, 0x00, 0x05, 0x3c,
	}
	buf := make([]byte, len(expected))
	_, err = serverConn.Read(buf)
	assert.Nil(t, err)
	assert.Equal(t, expected, buf)
}
