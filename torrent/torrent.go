package torrent

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"io"
	"net"

	"github.com/jackpal/bencode-go"
	"github.com/veggiedefender/torrent-client/p2p"
)

// Port to listen on
const Port uint16 = 6881

// Torrent encodes the metadata from a .torrent file
type Torrent struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type bencodeInfo struct {
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
}

type bencodeTorrent struct {
	Announce string      `bencode:"announce"`
	Info     bencodeInfo `bencode:"info"`
}

// Download downloads a torrent
func (t *Torrent) Download() error {
	var peerID [20]byte
	_, err := rand.Read(peerID[:])
	if err != nil {
		return err
	}

	// peers, err := t.getPeers(peerID, Port)
	peers := []p2p.Peer{{IP: net.IP{127, 0, 0, 1}, Port: 51413}}
	downloader := p2p.Download{
		Peers:       peers,
		PeerID:      peerID,
		InfoHash:    t.InfoHash,
		PieceHashes: t.PieceHashes,
		Length:      t.Length,
	}
	err = downloader.Download()
	return err
}

// Open parses a torrent file
func Open(r io.Reader) (*Torrent, error) {
	bto := bencodeTorrent{}
	err := bencode.Unmarshal(r, &bto)
	if err != nil {
		return nil, err
	}
	t, err := bto.toTorrent()
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (i *bencodeInfo) hash() ([20]byte, error) {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, *i)
	if err != nil {
		return [20]byte{}, err
	}
	h := sha1.Sum(buf.Bytes())
	return h, nil
}

func (i *bencodeInfo) splitPieceHashes() ([][20]byte, error) {
	hashLen := 20 // Length of SHA-1 hash
	buf := []byte(i.Pieces)
	if len(buf)%hashLen != 0 {
		err := fmt.Errorf("Received malformed pieces of length %d", len(buf))
		return nil, err
	}
	numHashes := len(buf) / hashLen
	hashes := make([][20]byte, numHashes)

	for i := 0; i < numHashes; i++ {
		copy(hashes[i][:], buf[i*hashLen:(i+1)*hashLen])
	}
	return hashes, nil
}

func (bto *bencodeTorrent) toTorrent() (*Torrent, error) {
	infoHash, err := bto.Info.hash()
	if err != nil {
		return nil, err
	}
	pieceHashes, err := bto.Info.splitPieceHashes()
	if err != nil {
		return nil, err
	}
	t := Torrent{
		Announce:    bto.Announce,
		InfoHash:    infoHash,
		PieceHashes: pieceHashes,
		PieceLength: bto.Info.PieceLength,
		Length:      bto.Info.Length,
		Name:        bto.Info.Name,
	}
	return &t, nil
}
