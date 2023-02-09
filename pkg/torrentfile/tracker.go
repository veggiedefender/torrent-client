package torrentfile

import (
	"crypto/tls"
	"fmt"
	"github.com/edelars/console-torrent-client/peers"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/jackpal/bencode-go"
)

const (
	httpTimeout = 15 * time.Second
)

type bencodeTrackerResp struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

func (t *TorrentFile) buildTrackerURL(peerID [20]byte, port uint16) (string, error) {
	base := t.UrlAnnounce
	params := url.Values{
		"info_hash":  []string{string(t.InfoHash[:])},
		"peer_id":    []string{string(peerID[:])},
		"port":       []string{strconv.Itoa(int(Port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(t.Length)},
	}
	base.RawQuery = params.Encode()
	return base.String(), nil
}

func (t *TorrentFile) requestPeers(peerID [20]byte, port uint16) ([]peers.Peer, error) {
	switch t.Proto {
	case TrackerProtoUdp:
		return t.requestPeersHttp(peerID, port)
	case TrackerProtoHttp:
		return t.requestPeersUdp(peerID, port)
	default:
		return []peers.Peer{}, ErrUnknownProto
	}
}

func (t *TorrentFile) requestPeersUdp(peerID [20]byte, port uint16) ([]peers.Peer, error) {
	url, err := t.buildTrackerURL(peerID, port)
	if err != nil {
		return nil, err
	}
	//TODO
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}

	c := &http.Client{
		Timeout:   httpTimeout,
		Transport: tr,
	}

	resp, err := c.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	trackerResp := bencodeTrackerResp{}
	err = bencode.Unmarshal(resp.Body, &trackerResp)
	if err != nil {
		return nil, err
	}
	fmt.Println(resp.Body)
	return peers.Unmarshal([]byte(trackerResp.Peers))
}

func (t *TorrentFile) requestPeersHttp(peerID [20]byte, port uint16) ([]peers.Peer, error) {
	url, err := t.buildTrackerURL(peerID, port)
	if err != nil {
		return nil, err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}

	c := &http.Client{
		Timeout:   httpTimeout,
		Transport: tr,
	}

	resp, err := c.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	trackerResp := bencodeTrackerResp{}
	err = bencode.Unmarshal(resp.Body, &trackerResp)
	if err != nil {
		return nil, err
	}
	fmt.Println(resp.Body)
	return peers.Unmarshal([]byte(trackerResp.Peers))
}
