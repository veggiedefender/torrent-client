package message

import (
	"encoding/binary"
	"fmt"
	"io"
)

type messageID uint8

// Message ID
const (
	MsgChoke         messageID = 0
	MsgUnchoke       messageID = 1
	MsgInterested    messageID = 2
	MsgNotInterested messageID = 3
	MsgHave          messageID = 4
	MsgBitfield      messageID = 5
	MsgRequest       messageID = 6
	MsgPiece         messageID = 7
	MsgCancel        messageID = 8
	MsgPort          messageID = 9
)

// Message m
type Message struct {
	ID      messageID
	Payload []byte
}

// Serialize serializes a message into a buffer of the form
// <length prefix><message ID><payload>
func (m *Message) Serialize() []byte {
	if m == nil {
		return make([]byte, 4)
	}
	length := uint32(len(m.Payload) + 1) // +1 for id
	buf := make([]byte, 4+length)
	binary.BigEndian.PutUint32(buf[0:4], length)
	buf[4] = byte(m.ID)
	copy(buf[5:], m.Payload)
	return buf
}

// Read parses a message from a stream. Returns `nil` on keep-alive message
func Read(r io.Reader) (*Message, error) {
	lengthBuf := make([]byte, 4)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lengthBuf)

	// keep-alive message
	if length == 0 {
		return nil, nil
	}

	messageBuf := make([]byte, length)
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		return nil, err
	}

	m := Message{
		ID:      messageID(messageBuf[0]),
		Payload: messageBuf[1:],
	}

	return &m, nil
}

func (m *Message) String() string {
	if m == nil {
		return "KeepAlive"
	}

	var idName string
	switch m.ID {
	case MsgChoke:
		idName = "Choke"
	case MsgUnchoke:
		idName = "Unchoke"
	case MsgInterested:
		idName = "Interested"
	case MsgNotInterested:
		idName = "NotInterested"
	case MsgHave:
		idName = "Have"
	case MsgBitfield:
		idName = "Bitfield"
	case MsgRequest:
		idName = "Request"
	case MsgPiece:
		idName = "Piece"
	case MsgCancel:
		idName = "Cancel"
	case MsgPort:
		idName = "Port"
	default:
		idName = fmt.Sprintf("Unknown#%d", m.ID)
	}

	return fmt.Sprintf("%s\t[% x]", idName, m.Payload)
}
