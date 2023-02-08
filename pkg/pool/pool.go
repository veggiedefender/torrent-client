package pool

import (
	"context"
	"github.com/edelars/console-torrent-client/pkg/console_print"
	"github.com/edelars/console-torrent-client/torrentfile"
	"sync"
)

type TorrentPool struct {
	logger                          console_print.Logger
	pool                            map[[20]byte]itemPool
	torrentFileDir, downloadFileDir string
	mutex                           sync.Mutex
}

type itemPool struct {
	tf     torrentfile.TFInt
	status int
}

func NewTorrentPool(logger console_print.Logger, torrentFileDir, downloadFileDir string) TorrentPool {

	return TorrentPool{
		logger:          logger,
		torrentFileDir:  torrentFileDir,
		downloadFileDir: downloadFileDir,
		pool:            make(map[[20]byte]itemPool),
	}
}

// Start perform start our pool of torrents
func (p *TorrentPool) Start(ctx context.Context) error {

	return nil
}

// Stop perform stop our pool of torrents
func (p *TorrentPool) Stop() error {

	return nil
}

func (p *TorrentPool) AddFileToPool(filename string) error {

	tf, err := torrentfile.NewTorrentFileFromFile(filename)
	if err != nil {
		return err
	}

	itemPool := itemPool{
		tf:     &tf,
		status: 0,
	}
	p.addPool(tf.InfoHash, itemPool)

	err = tf.DownloadToFile(context.Background(), p.downloadFileDir)
	if err != nil {
		return err
	}
	return nil
}

func (p *TorrentPool) addPool(id [20]byte, i itemPool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.pool[id] = i
}
