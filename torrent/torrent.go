package torrent

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"io"

	"github.com/jackpal/bencode-go"
)

const port = 6881

// Torrent encodes the metadata from a .torrent file
type Torrent struct {
	Announce    string
	InfoHash    []byte
	PieceHashes [][]byte
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
func (to *Torrent) Download() error {
	peerID := make([]byte, 20)
	_, err := rand.Read(peerID)
	if err != nil {
		return err
	}

	tracker := Tracker{
		PeerID:  peerID,
		Torrent: to,
		Port:    port,
	}
	peers, err := tracker.getPeers()
	fmt.Println(peers)
	return nil
}

// Open parses a torrent file
func Open(r io.Reader) (*Torrent, error) {
	bto := bencodeTorrent{}
	err := bencode.Unmarshal(r, &bto)
	if err != nil {
		return nil, err
	}
	to, err := bto.toTorrent()
	if err != nil {
		return nil, err
	}
	return to, nil
}

func (i *bencodeInfo) hash() ([]byte, error) {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, *i)
	if err != nil {
		return nil, err
	}
	h := sha1.Sum(buf.Bytes())
	return h[:], nil
}

func (i *bencodeInfo) splitPieceHashes() ([][]byte, error) {
	hashLen := 20 // Length of SHA-1 hash
	buf := []byte(i.Pieces)
	if len(buf)%hashLen != 0 {
		err := fmt.Errorf("Received malformed pieces of length %d", len(buf))
		return nil, err
	}
	numHashes := len(buf) / hashLen
	hashes := make([][]byte, numHashes)

	for i := 0; i < numHashes; i++ {
		hashes[i] = buf[i*hashLen : (i+1)*hashLen]
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
	to := Torrent{
		Announce:    bto.Announce,
		InfoHash:    infoHash,
		PieceHashes: pieceHashes,
		PieceLength: bto.Info.PieceLength,
		Length:      bto.Info.Length,
		Name:        bto.Info.Name,
	}
	return &to, nil
}
