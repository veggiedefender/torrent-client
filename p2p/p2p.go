package p2p

import (
	"fmt"
	"net"
	"sync"
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

type pieceWork struct {
	index int
	hash  [20]byte
}

type swarm struct {
	clients []*client
	queue   chan *pieceWork
	mux     sync.Mutex
}

// Download downloads a torrent
func (d *Download) Download() error {
	clients := d.initClients()
	if len(clients) == 0 {
		return fmt.Errorf("Could not connect to any of %d clients", len(d.Peers))
	}

	queue := make(chan *pieceWork, len(d.PieceHashes))
	for index, hash := range d.PieceHashes {
		queue <- &pieceWork{index, hash}
	}
	processQueue(clients, queue)

	return nil
}

func (d *Download) initClients() []*client {
	// Create clients in parallel
	c := make(chan *client)
	for _, p := range d.Peers {
		go func(p Peer) {
			client, err := newClient(p, d.PeerID, d.InfoHash)
			if err != nil {
				c <- nil
			} else {
				c <- client
			}
		}(p)
	}

	clients := make([]*client, 0)
	for range d.Peers {
		client := <-c
		if client != nil {
			clients = append(clients, client)
		}
	}
	return clients
}

func (s *swarm) selectClient(index int) *client {
	return s.clients[0]
}

func processQueue(clients []*client, queue chan *pieceWork) {
	s := swarm{clients, queue, sync.Mutex{}}
	for pw := range s.queue {
		client := s.selectClient(pw.index)
		fmt.Println(client.conn.RemoteAddr())
		break
	}
}
