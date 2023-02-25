package fileinfo

import (
	"fmt"
	"os"
	"path/filepath"
)

type FileInfo struct {
	Length int      `bencode:"length"`
	Path   []string `bencode:"path"`
}

func (file *FileInfo) WriteToDisk(buf []byte, files []FileInfo, path, torrentName string) error {
	outputPath := createPath(path, torrentName, file.Path[0], files)

	err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm)
	if err != nil {
		return err
	}

	fileBuf := make([]byte, file.Length)
	copy(fileBuf, buf[getOffset(*file, files):getOffset(*file, files)+file.Length])

	err = os.WriteFile(outputPath, fileBuf, os.ModePerm)
	if err != nil {
		return err
	}

	fmt.Printf("saved file %s\n", file.Path[0])

	return nil
}

func getOffset(file FileInfo, files []FileInfo) int {
	offset := 0
	for _, f := range files {
		if f.Length == file.Length && f.Path[0] == file.Path[0] {
			return offset
		}
		offset += f.Length
	}
	return -1
}

func createPath(path, torrentName, filename string, files []FileInfo) (outputPath string) {
	if len(files) < 2 {
		return filepath.Join(path, torrentName)
	}

	return filepath.Join(path, torrentName, filename)
}
