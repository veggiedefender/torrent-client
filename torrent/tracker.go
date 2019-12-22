package torrent

import (
	"encoding/binary"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/jackpal/bencode-go"
)

// Tracker tracker
type Tracker struct {
	PeerID  []byte
	Torrent *Torrent
	Port    uint16
}

// TrackerResponse t
type TrackerResponse struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"port"`
}

// Peer p
type Peer struct {
	IP   net.IP
	Port uint16
}

func parsePeers(peersBin string) ([]Peer, error) {
	peerSize := 6 // 4 for IP, 2 for port
	numPeers := len(peersBin) / peerSize
	if len(peersBin)%peerSize != 0 {
		err := errors.New("Received malformed peers")
		return nil, err
	}
	peers := make([]Peer, numPeers)
	for i := 0; i < numPeers; i++ {
		offset := i * peerSize
		peers[i].IP = net.IP(peersBin[offset : offset+4])
		peers[i].Port = binary.BigEndian.Uint16([]byte(peersBin[offset+4 : offset+6]))
	}
	return peers, nil
}

func (tr *Tracker) buildTrackerURL() (string, error) {
	base, err := url.Parse(tr.Torrent.Announce)
	if err != nil {
		return "", err
	}
	infoHash, err := tr.Torrent.Info.hash()
	if err != nil {
		return "", err
	}
	params := url.Values{
		"info_hash":  []string{string(infoHash)},
		"peer_id":    []string{string(tr.PeerID)},
		"port":       []string{strconv.Itoa(int(tr.Port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(tr.Torrent.Info.Length)},
	}
	base.RawQuery = params.Encode()
	return base.String(), nil
}

func (tr *Tracker) getPeers() ([]Peer, error) {
	url, err := tr.buildTrackerURL()
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	trackerResp := TrackerResponse{}
	err = bencode.Unmarshal(resp.Body, &trackerResp)
	if err != nil {
		return nil, err
	}

	peers, err := parsePeers(trackerResp.Peers)
	if err != nil {
		return nil, err
	}

	return peers, nil
}
