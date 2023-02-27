package torrentfile

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"os"

	"github.com/jackpal/bencode-go"
	"github.com/veggiedefender/torrent-client/fileinfo"
	"github.com/veggiedefender/torrent-client/p2p"
)

// Port to listen on
const Port uint16 = 6881

// TorrentFile encodes the metadata from a .torrent file
type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Name        string
	Length      int
	Files       []fileinfo.FileInfo
}

type bencodeTorrent struct {
	Announce string      `bencode:"announce"`
	Info     bencodeInfo `bencode:"info"`
}

type bencodeInfo struct {
	Pieces      string              `bencode:"pieces"`
	PieceLength int                 `bencode:"piece length"`
	Name        string              `bencode:"name"`
	Length      int                 `bencode:"length,omitempty"`
	Files       []fileinfo.FileInfo `bencode:"files,omitempty"`
}

// DownloadToFile downloads a torrent and writes it to a file
func (t *TorrentFile) DownloadToFile(path string) error {
	var peerID [20]byte
	_, err := rand.Read(peerID[:])
	if err != nil {
		return err
	}

	peers, err := t.requestPeers(peerID, Port)
	if err != nil {
		return err
	}

	torrent := p2p.Torrent{
		Peers:       peers,
		PeerID:      peerID,
		InfoHash:    t.InfoHash,
		PieceHashes: t.PieceHashes,
		PieceLength: t.PieceLength,
		Name:        t.Name,
		Length:      t.Length,
	}
	buf, err := torrent.Download()
	if err != nil {
		return err
	}

	return t.createFiles(path, buf)
}

func (t *TorrentFile) createFiles(path string, buf []byte) error {
	fmt.Println("Saving files...")
	for _, file := range t.Files {
		err := file.WriteToDisk(buf, t.Files, path, t.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

// Open parses a torrent file
func Open(path string) (TorrentFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return TorrentFile{}, err
	}
	defer file.Close()

	bto := bencodeTorrent{}
	err = bencode.Unmarshal(file, &bto)
	if err != nil {
		return TorrentFile{}, err
	}
	return bto.toTorrentFile()
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
		err := fmt.Errorf("received malformed pieces of length %d", len(buf))
		return nil, err
	}
	numHashes := len(buf) / hashLen
	hashes := make([][20]byte, numHashes)

	for i := 0; i < numHashes; i++ {
		copy(hashes[i][:], buf[i*hashLen:(i+1)*hashLen])
	}
	return hashes, nil
}

func (i *bencodeInfo) setTotalLength() int {
	var length int

	if len(i.Files) != 0 && i.Length == 0 {

		for _, file := range i.Files {
			length += file.Length
		}

	} else {
		length = i.Length
	}

	return length
}

func (i *bencodeInfo) setFileInfos() []fileinfo.FileInfo {
	if len(i.Files) != 0 && i.Length == 0 {
		return i.Files
	}

	singleFile := fileinfo.FileInfo{
		Length: i.Length,
		Path:   []string{i.Name},
	}

	return []fileinfo.FileInfo{
		singleFile,
	}
}

func (bto *bencodeTorrent) toTorrentFile() (TorrentFile, error) {
	infoHash, err := bto.Info.hash()
	if err != nil {
		return TorrentFile{}, err
	}
	pieceHashes, err := bto.Info.splitPieceHashes()
	if err != nil {
		return TorrentFile{}, err
	}
	t := TorrentFile{
		Announce:    bto.Announce,
		InfoHash:    infoHash,
		PieceHashes: pieceHashes,
		PieceLength: bto.Info.PieceLength,
		Name:        bto.Info.Name,
		Length:      bto.Info.setTotalLength(),
		Files:       bto.Info.setFileInfos(),
	}
	return t, nil
}
