package torrentfile

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"errors"
	"fmt"
	"github.com/edelars/console-torrent-client/p2p"
	"github.com/jackpal/bencode-go"
	"net/url"
	"os"
)

const (
	TrackerProtoHttp = iota
	TrackerProtoUdp
)

var (
	ErrUnknownProto = errors.New("unknown proto")
)

type Proto int

// Port to listen on
const Port uint16 = 6881

// TorrentFile encodes the metadata from a .torrent file
type TorrentFile struct {
	UrlAnnounce url.URL
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
	Proto       Proto
}

func (t *TorrentFile) VerifyFiles(ctx context.Context, path string) error {
	//TODO implement me
	panic("implement me")
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

// DownloadToFile downloads a torrent and writes it to a file
func (t *TorrentFile) DownloadToFile(ctx context.Context, path string) error {
	var peerID [20]byte
	_, err := rand.Read(peerID[:])
	if err != nil {
		return err
	}

	peers, err := t.requestPeers(peerID, Port)
	if err != nil {
		return err
	}

	outFile, err := os.Create(path)
	if err != nil {
		return err
	}

	defer outFile.Close()

	torrent := p2p.Torrent{
		Peers:       peers,
		PeerID:      peerID,
		InfoHash:    t.InfoHash,
		PieceHashes: t.PieceHashes,
		PieceLength: t.PieceLength,
		Length:      t.Length,
		Name:        t.Name,
	}
	ch, err := torrent.Download(ctx)
	if err != nil {
		return err
	}
	for d := range ch {
		if err := writePieceToFile(outFile, d); err != nil {
			return err
		}
	}

	return nil
}

func writePieceToFile(outFile *os.File, data p2p.PieceFile) error {
	if _, err := outFile.Seek(data.Begin, 0); err != nil {
		return err
	}
	if _, err := outFile.Write(data.Data); err != nil {
		return err
	}
	return nil
}

// NewTorrentFileFromFile parses a torrent file
func NewTorrentFileFromFile(path string) (TorrentFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return TorrentFile{}, err
	}
	defer file.Close()

	bto := bencodeTorrent{}
	err = bencode.Unmarshal(file, &bto)

	//bto.Announce = "udp://tracker.opentrackr.org:1337/announce"

	if err != nil {
		return TorrentFile{}, err
	}

	return newFromBto(bto)
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

func newFromBto(bto bencodeTorrent) (t TorrentFile, err error) {

	u, err := url.Parse(bto.Announce)
	if err != nil {
		return t, err
	}

	switch u.Scheme {
	case "https":
		fallthrough
	case "http":
		t.Proto = TrackerProtoHttp
	case "udp":
		t.Proto = TrackerProtoUdp
	default:
		return t, ErrUnknownProto
	}

	infoHash, err := bto.Info.hash()
	if err != nil {
		return t, err
	}
	pieceHashes, err := bto.Info.splitPieceHashes()
	if err != nil {
		return t, err
	}

	t.UrlAnnounce = *u
	t.InfoHash = infoHash
	t.PieceHashes = pieceHashes
	t.PieceLength = bto.Info.PieceLength
	t.Length = bto.Info.Length
	t.Name = bto.Info.Name

	return t, nil
}
