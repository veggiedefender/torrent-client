package p2p

import (
	"encoding/hex"
	"fmt"
	"net"
	"strconv"

	"github.com/veggiedefender/torrent-client/message"
)

// Peer encodes information for connecting to a peer
type Peer struct {
	IP   net.IP
	Port uint16
}

// Connect connects to a peer
func Connect(p *Peer, peerID [20]byte, infoHash [20]byte) error {
	hostPort := net.JoinHostPort(p.IP.String(), strconv.Itoa(int(p.Port)))
	conn, err := net.Dial("tcp", hostPort)
	if err != nil {
		return err
	}

	h := message.Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: infoHash,
		PeerID:   infoHash,
	}
	_, err = conn.Write(h.Serialize())
	if err != nil {
		return err
	}

	reply, err := message.ReadHandshake(conn)
	if err != nil {
		return err
	}

	fmt.Println("Handshake received:")
	fmt.Println(reply.Pstr)
	fmt.Println("Peer ID", string(reply.PeerID[:]))
	fmt.Println("Hash", hex.EncodeToString(reply.InfoHash[:]))

	return nil
}
