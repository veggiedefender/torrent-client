package message

import (
	"encoding/binary"
	"errors"
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

// Message stores ID and payload of a message
type Message struct {
	ID      messageID
	Payload []byte
}

// FormatRequest formats the ID and payload for a request message
func FormatRequest(index, begin, length int) *Message {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))
	return &Message{
		ID:      MsgRequest,
		Payload: payload,
	}
}

// ParsePiece parses a piece message and copies its payload into a buffer
func ParsePiece(index int, buf []byte, msg *Message) (int, error) {
	if msg.ID != MsgPiece {
		return 0, fmt.Errorf("Expected ID %d, got ID %d", MsgPiece, msg.ID)
	}
	if len(msg.Payload) < 8 {
		return 0, errors.New("Payload too short")
	}
	parsedIndex := int(binary.BigEndian.Uint32(msg.Payload[0:4]))
	if parsedIndex != index {
		return 0, fmt.Errorf("Expected index %d, got %d", index, parsedIndex)
	}
	begin := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
	if begin >= len(buf) {
		return 0, fmt.Errorf("Begin offset too high. %d >= %d", begin, len(buf))
	}
	data := msg.Payload[8:]
	if begin+len(data) > len(buf) {
		return 0, fmt.Errorf("Data too long [%d] for offset %d with length %d", len(data), begin, len(buf))
	}
	copy(buf[begin:], data)
	return len(data), nil
}

// Serialize serializes a message into a buffer of the form
// <length prefix><message ID><payload>
// Interprets `nil` as a keep-alive message
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

// ReadMessage parses a message from a stream. Returns `nil` on keep-alive message
func ReadMessage(r io.Reader) (*Message, error) {
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
