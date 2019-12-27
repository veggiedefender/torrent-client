package p2p

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/veggiedefender/torrent-client/message"

	"github.com/veggiedefender/torrent-client/handshake"
)

type client struct {
	peer     Peer
	infoHash [20]byte
	peerID   [20]byte
	conn     net.Conn
	reader   *bufio.Reader
	bitfield message.Bitfield
	choked   bool
	engaged  bool
}

func completeHandshake(conn net.Conn, r *bufio.Reader, infohash, peerID [20]byte) (*handshake.Handshake, error) {
	conn.SetDeadline(time.Now().Local().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{}) // Disable the deadline

	req := handshake.New(infohash, peerID)
	_, err := conn.Write(req.Serialize())
	if err != nil {
		return nil, err
	}

	res, err := handshake.Read(r)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func recvBitfield(conn net.Conn, r *bufio.Reader) (message.Bitfield, error) {
	conn.SetDeadline(time.Now().Local().Add(5 * time.Second))
	defer conn.SetDeadline(time.Time{}) // Disable the deadline

	msg, err := message.Read(r)
	if err != nil {
		return nil, err
	}
	if msg.ID != message.MsgBitfield {
		err := fmt.Errorf("Expected bitfield but got ID %d", msg.ID)
		return nil, err
	}

	return msg.Payload, nil
}

func newClient(peer Peer, peerID, infoHash [20]byte) (*client, error) {
	hostPort := net.JoinHostPort(peer.IP.String(), strconv.Itoa(int(peer.Port)))
	conn, err := net.DialTimeout("tcp", hostPort, 3*time.Second)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(conn)

	_, err = completeHandshake(conn, reader, infoHash, peerID)
	if err != nil {
		return nil, err
	}

	bf, err := recvBitfield(conn, reader)
	if err != nil {
		return nil, err
	}

	return &client{
		peer:     peer,
		infoHash: infoHash,
		peerID:   peerID,
		conn:     conn,
		reader:   reader,
		bitfield: bf,
		choked:   true,
		engaged:  false,
	}, nil
}

func (c *client) hasPiece(index int) bool {
	return c.bitfield.HasPiece(index)
}

func (c *client) hasNext() bool {
	fmt.Println(c.reader.Buffered() > 0)
	return c.reader.Buffered() > 0
}

func (c *client) read() (*message.Message, error) {
	msg, err := message.Read(c.reader)
	return msg, err
}

func (c *client) request(index, begin, length int) error {
	req := message.FormatRequest(index, begin, length)
	_, err := c.conn.Write(req.Serialize())
	return err
}

func (c *client) interested() error {
	msg := message.Message{ID: message.MsgInterested}
	_, err := c.conn.Write(msg.Serialize())
	return err
}

func (c *client) notInterested() error {
	msg := message.Message{ID: message.MsgNotInterested}
	_, err := c.conn.Write(msg.Serialize())
	return err
}

func (c *client) unchoke() error {
	msg := message.Message{ID: message.MsgUnchoke}
	_, err := c.conn.Write(msg.Serialize())
	return err
}

func (c *client) have(index int) error {
	pl := make([]byte, 4)
	binary.BigEndian.PutUint32(pl, uint32(index))
	msg := message.Message{ID: message.MsgHave, Payload: pl}
	_, err := c.conn.Write(msg.Serialize())
	return err
}
