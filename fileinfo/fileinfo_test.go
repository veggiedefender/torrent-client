package fileinfo

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreatePathSingleFileTorrent(t *testing.T) {
	files := []FileInfo{
		{
			Path:   []string{"archlinux-2019.12.01-x86_64.iso"},
			Length: 1,
		},
	}

	want := "test/archlinux-2019.12.01-x86_64.iso"
	got := createPath("test", files[0].Path[0], "archlinux-2019.12.01-x86_64.iso", files)

	assert.Equal(t, want, got)
}

func TestCreatePathMultiFileTorrent(t *testing.T) {
	files := []FileInfo{
		{
			Path:   []string{"archlinux-2019.12.01-x86_64.iso"},
			Length: 1,
		},
		{
			Path:   []string{"sali.txt"},
			Length: 1,
		},
	}

	givenPath := "dir1"
	torrentName := "dir2"

	for _, file := range files {
		want := fmt.Sprintf("%s/%s/%s", givenPath, torrentName, file.Path[0])
		got := createPath(givenPath, torrentName, file.Path[0], files)
		assert.Equal(t, want, got)
	}
}

// good idea to test writes to the system?
// func TestWriteToDisk(t *testing.T) {
// 	buf := []byte("helloworld")
// 	files := []FileInfo{
// 		{
// 			Path:   []string{"file1.txt"},
// 			Length: 5,
// 		},
// 		{
// 			Path:   []string{"file2.txt"},
// 			Length: 5,
// 		},
// 	}
// 	path := "dir1"
// 	torrentName := "alexandria"

// 	for _, file := range files {
// 		err := file.WriteToDisk(buf, files, path, torrentName)
// 		require.Nil(t, err)
// 	}
// }
