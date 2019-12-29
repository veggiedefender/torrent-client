package p2p

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"
	"net"
	"runtime"

	"github.com/veggiedefender/torrent-client/message"
)

const maxBlockSize = 16384
const maxBacklog = 5

// Peer encodes connection information for a peer
type Peer struct {
	IP   net.IP
	Port uint16
}

// Torrent holds data required to download a torrent from a list of peers
type Torrent struct {
	Peers       []Peer
	PeerID      [20]byte
	InfoHash    [20]byte
	PieceHashes [][20]byte
	Length      int
}

type pieceWork struct {
	index  int
	hash   [20]byte
	length int
}

type pieceResult struct {
	index int
	buf   []byte
}

type downloadState struct {
	index      int
	client     *client
	buf        []byte
	downloaded int
	backlog    int
}

func readMessage(state *downloadState) error {
	msg, err := state.client.read() // this call blocks
	if err != nil {
		return err
	}

	if msg == nil { // keep-alive
		return nil
	}

	switch msg.ID {
	case message.MsgUnchoke:
		state.client.choked = false
	case message.MsgChoke:
		state.client.choked = true
	case message.MsgHave:
		index, err := message.ParseHave(msg)
		if err != nil {
			return err
		}
		state.client.bitfield.SetPiece(index)
	case message.MsgPiece:
		n, err := message.ParsePiece(state.index, state.buf, msg)
		if err != nil {
			return err
		}
		state.downloaded += n
		state.backlog--
	}
	return nil
}

func attemptDownloadPiece(c *client, pw *pieceWork) ([]byte, error) {
	state := downloadState{
		index:  pw.index,
		client: c,
		buf:    make([]byte, pw.length),
	}

	requested := 0
	for state.downloaded < pw.length {
		for c.hasNext() {
			err := readMessage(&state)
			if err != nil {
				return nil, err
			}
		}

		if !c.choked && requested < pw.length && requested-state.downloaded <= maxBacklog+1 {
			for i := 0; i < maxBacklog; i++ {
				blockSize := maxBlockSize
				if pw.length-requested < blockSize {
					// Last block might be shorter than the typical block
					blockSize = pw.length - requested
				}
				c.request(pw.index, requested, blockSize)
				requested += blockSize
			}
		}

		err := readMessage(&state)
		if err != nil {
			return nil, err
		}
	}
	return state.buf, nil
}

func checkIntegrity(pw *pieceWork, buf []byte) error {
	hash := sha1.Sum(buf)
	if !bytes.Equal(hash[:], pw.hash[:]) {
		return fmt.Errorf("Index %d failed integrity check", pw.index)
	}
	return nil
}

func (t *Torrent) startDownloadWorker(peer Peer, workQueue chan *pieceWork, results chan *pieceResult) {
	c, err := newClient(peer, t.PeerID, t.InfoHash)
	if err != nil {
		log.Printf("Could not handshake with %s. Disconnecting\n", peer.IP)
		return
	}
	defer c.conn.Close()
	log.Printf("Completed handshake with %s\n", peer.IP)

	c.unchoke()
	c.interested()

	for pw := range workQueue {
		if !c.hasPiece(pw.index) {
			workQueue <- pw // Put piece back on the queue
			continue
		}

		// Download the piece
		buf, err := attemptDownloadPiece(c, pw)
		if err != nil {
			log.Println("Exiting", err)
			workQueue <- pw // Put piece back on the queue
			return
		}

		err = checkIntegrity(pw, buf)
		if err != nil {
			log.Printf("Piece #%d failed integrity check\n", pw.index)
			workQueue <- pw // Put piece back on the queue
			continue
		}

		results <- &pieceResult{pw.index, buf}
		c.have(pw.index)
	}
}

func calculateBoundsForPiece(index, numPieces, length int) (begin int, end int) {
	pieceLength := length / numPieces
	begin = index * pieceLength
	end = begin + pieceLength
	return begin, end
}

// Download downloads the torrent
func (t *Torrent) Download() ([]byte, error) {
	// Init queues for workers to retrieve work and send results
	workQueue := make(chan *pieceWork, len(t.PieceHashes))
	results := make(chan *pieceResult, len(t.PieceHashes))
	for index, hash := range t.PieceHashes {
		length := t.Length / len(t.PieceHashes)
		workQueue <- &pieceWork{index, hash, length}
	}

	// Start workers
	for _, peer := range t.Peers {
		go t.startDownloadWorker(peer, workQueue, results)
	}

	// Collect results into a buffer until full
	buf := make([]byte, t.Length)
	donePieces := 0
	for donePieces < len(t.PieceHashes) {
		res := <-results
		begin, end := calculateBoundsForPiece(res.index, len(t.PieceHashes), t.Length)
		copy(buf[begin:end], res.buf)
		donePieces++

		percent := float64(donePieces) / float64(len(t.PieceHashes)) * 100
		log.Printf("(%0.2f%%) Downloaded piece #%d with %d goroutines\n", percent, res.index, runtime.NumGoroutine())
	}
	close(workQueue)

	return buf, nil
}
