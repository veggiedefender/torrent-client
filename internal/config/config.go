package config

import (
	"os"
)

type Environment struct {
	DFDir string `short:"d" long:"dir-files" env:"CTC_DOWNLOAD_FILES_DIR" description:"Download files dir" default:"./"`
	TFDir string `short:"t" long:"torrent-files" env:"CTC_TORRENT_FILES_DIR" description:".torrent files storage dir" default:"./"`
}

func (e Environment) Init() (err error) {
	if e.DFDir == "./" {
		if e.DFDir, err = os.Getwd(); err != nil {
			return err
		}
	}

	if e.TFDir == "./" {
		if e.TFDir, err = os.Getwd(); err != nil {
			return err
		}
	}

	return nil
}
