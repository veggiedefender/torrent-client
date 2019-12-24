package p2p

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"math"
	"net"
	"strconv"

	"github.com/veggiedefender/torrent-client/handshake"
	"github.com/veggiedefender/torrent-client/message"
)

// Peer encodes connection information for a peer
type Peer struct {
	IP   net.IP
	Port uint16
}

// Downloader holds data required to download a torrent from a list of peers
type Downloader struct {
	Peers       []Peer
	PeerID      [20]byte
	InfoHash    [20]byte
	PieceHashes [][20]byte
	Length      int
}

// Download downloads a torrent
func (d *Downloader) Download() error {
	conn, err := d.Peers[0].connect(d.PeerID, d.InfoHash)
	if err != nil {
		return err
	}
	defer conn.Close()
	h, err := d.handshake(conn)
	if err != nil {
		return err
	}
	fmt.Println(h)

	choked := false
	pieceSize := d.Length / len(d.PieceHashes)
	buf := make([]byte, pieceSize)
	i := 0
	for i < pieceSize {
		msg, err := message.Read(conn)
		if err != nil {
			return err
		}

		if msg.ID != message.MsgPiece {
			fmt.Println(msg.String())
		} else {
			fmt.Println("Received", len(msg.Payload), "bytes")
		}

		switch msg.ID {
		case message.MsgChoke:
			choked = true
		case message.MsgUnchoke:
			choked = false
		case message.MsgPiece:
			n, err := message.ParsePiece(0, buf, msg)
			if err != nil {
				return err
			}
			i += n
		}

		if !choked {
			index := 0 // Piece number
			begin := i // Offset
			remain := pieceSize - i
			length := int(math.Min(float64(16384), float64(pieceSize)))
			length = int(math.Min(float64(remain), float64(length)))
			_, err := conn.Write(message.FormatRequest(index, begin, length).Serialize())
			if err != nil {
				return err
			}
		}
	}

	s := sha1.Sum(buf)
	fmt.Printf("Downloaded %d bytes.\n", len(buf))
	fmt.Printf("Got SHA1\t%s\n", hex.EncodeToString(s[:]))
	fmt.Printf("Expected\t%s\n", hex.EncodeToString(d.PieceHashes[0][:]))

	return nil
}

func (p *Peer) connect(peerID [20]byte, infoHash [20]byte) (net.Conn, error) {
	hostPort := net.JoinHostPort(p.IP.String(), strconv.Itoa(int(p.Port)))
	conn, err := net.Dial("tcp", hostPort)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (d *Downloader) handshake(conn net.Conn) (*handshake.Handshake, error) {
	req := handshake.New(d.InfoHash, d.PeerID)
	_, err := conn.Write(req.Serialize())
	if err != nil {
		return nil, err
	}

	res, err := handshake.Read(conn)
	if err != nil {
		return nil, err
	}
	return res, nil
}
