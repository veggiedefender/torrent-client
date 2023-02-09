package config

import (
	"github.com/kirsle/configdir"
	"os"
	"path/filepath"
)

const (
	constSQLiteFileName = "ctc_sqlite.db"
)

type Environment struct {
	SQLiteFile string `short:"s" long:"sqlite-file" env:"CTC_SQLITE_FILE" description:"SQLite db file" default:""`
	DFDir      string `short:"d" long:"dir-files" env:"CTC_DOWNLOAD_FILES_DIR" description:"Download files dir" default:"./"`
	TFDir      string `short:"t" long:"torrent-files" env:"CTC_TORRENT_FILES_DIR" description:".torrent files storage dir" default:"./"`
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

func (e Environment) GetSQLiteFIle() string {
	if e.SQLiteFile != "" {
		return e.SQLiteFile
	}
	cfgDir := configdir.LocalConfig()
	return filepath.Join(cfgDir, constSQLiteFileName)
}
