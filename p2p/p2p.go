package p2p

import (
	"net"
	"strconv"
	"time"

	"github.com/veggiedefender/torrent-client/handshake"
)

// Peer encodes connection information for a peer
type Peer struct {
	IP   net.IP
	Port uint16
}

// Download holds data required to download a torrent from a list of peers
type Download struct {
	Peers       []Peer
	PeerID      [20]byte
	InfoHash    [20]byte
	PieceHashes [][20]byte
	Length      int
}

type peerState struct {
	peer *Peer
	conn net.Conn
}

type swarm struct {
	peerStates []*peerState
}

func (p *Peer) connect(peerID [20]byte, infoHash [20]byte) (net.Conn, error) {
	hostPort := net.JoinHostPort(p.IP.String(), strconv.Itoa(int(p.Port)))
	conn, err := net.DialTimeout("tcp", hostPort, 3*time.Second)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (d *Download) handshake(conn net.Conn) (*handshake.Handshake, error) {
	conn.SetDeadline(time.Now().Local().Add(3 * time.Second))
	req := handshake.New(d.InfoHash, d.PeerID)
	_, err := conn.Write(req.Serialize())
	if err != nil {
		return nil, err
	}

	res, err := handshake.Read(conn)
	if err != nil {
		return nil, err
	}
	conn.SetDeadline(time.Time{}) // Disable the deadline
	return res, nil
}

func (d *Download) initPeer(p *Peer, c chan *peerState) {
	conn, err := p.connect(d.PeerID, d.InfoHash)
	if err != nil {
		c <- nil
		return
	}
	_, err = d.handshake(conn)
	if err != nil {
		c <- nil
		return
	}
	c <- &peerState{p, conn}
}

func (d *Download) startSwarm() *swarm {
	c := make(chan *peerState)
	for i := range d.Peers {
		go d.initPeer(&d.Peers[i], c)
	}

	peerStates := make([]*peerState, 0)
	for range d.Peers {
		ps := <-c
		if ps != nil {
			peerStates = append(peerStates, ps)
		}
	}

	return &swarm{peerStates}
}

// Download downloads a torrent
func (d *Download) Download() error {
	d.startSwarm()
	return nil
}
