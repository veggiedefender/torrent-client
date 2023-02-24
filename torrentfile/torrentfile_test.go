package torrentfile

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/jackpal/bencode-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update .golden.json files")

func TestOpen(t *testing.T) {
	torrent, err := Open("testdata/archlinux-2019.12.01-x86_64.iso.torrent")
	require.Nil(t, err)

	goldenPath := "testdata/archlinux-2019.12.01-x86_64.iso.torrent.golden.json"
	if *update {
		serialized, err := json.MarshalIndent(torrent, "", "  ")
		require.Nil(t, err)
		os.WriteFile(goldenPath, serialized, 0644)
	}

	expected := TorrentFile{}
	golden, err := os.ReadFile(goldenPath)
	require.Nil(t, err)
	err = json.Unmarshal(golden, &expected)

	require.Nil(t, err)

	assert.Equal(t, expected, torrent)
}

func TestCreatePathSingleFileTorrent(t *testing.T) {
	single, err := Open("testdata/archlinux-2019.12.01-x86_64.iso.torrent")
	require.Nil(t, err)

	expect := single.Name
	want := createPath(".", single.Files[0].Path[0], &single)

	assert.Equal(t, expect, want)
}

func TestCreatePathMultiFileTorrent(t *testing.T) {
	multi, err := Open("testdata/bocchi.torrent")
	require.Nil(t, err)

	for _, file := range multi.Files {
		want := createPath(".", file.Path[0], &multi)
		expect := fmt.Sprintf("%s/%s", multi.Name, file.Path[0])
		assert.Equal(t, expect, want)
	}
}

func TestBencodeMultifileTorrent(t *testing.T) {
	path := "testdata/bocchi.torrent"
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	bto := bencodeTorrent{}
	err = bencode.Unmarshal(file, &bto)
	if err != nil {
		t.Fatal(err)
	}

	want := fileInfo{
		Length: 63109789,
		Path:   []string{"14 転がる岩、君に朝が降る.flac"},
	}

	got := bto.Info.Files[0]

	assert.Equal(t, want, got)

}

func TestSetLengthOfMultifileTorrent(t *testing.T) {
	torrent, err := Open("testdata/bocchi.torrent")
	assert.Nil(t, err)
	assert.Equal(t, torrent.Length, 764345370)
}

func TestToTorrentFile(t *testing.T) {
	tests := map[string]struct {
		input  *bencodeTorrent
		output TorrentFile
		fails  bool
	}{
		"correct conversion": {
			input: &bencodeTorrent{
				Announce: "http://bttracker.debian.org:6969/announce",
				Info: bencodeInfo{
					Pieces:      "1234567890abcdefghijabcdefghij1234567890",
					PieceLength: 262144,
					Length:      351272960,
					Name:        "debian-10.2.0-amd64-netinst.iso",
				},
			},
			output: TorrentFile{
				Announce: "http://bttracker.debian.org:6969/announce",
				InfoHash: [20]byte{216, 247, 57, 206, 195, 40, 149, 108, 204, 91, 191, 31, 134, 217, 253, 207, 219, 168, 206, 182},
				PieceHashes: [][20]byte{
					{49, 50, 51, 52, 53, 54, 55, 56, 57, 48, 97, 98, 99, 100, 101, 102, 103, 104, 105, 106},
					{97, 98, 99, 100, 101, 102, 103, 104, 105, 106, 49, 50, 51, 52, 53, 54, 55, 56, 57, 48},
				},
				PieceLength: 262144,
				Length:      351272960,
				Name:        "debian-10.2.0-amd64-netinst.iso",
				Files: []fileInfo{
					{
						Length: 351272960,
						Path:   []string{"debian-10.2.0-amd64-netinst.iso"},
					},
				},
			},
			fails: false,
		},
		"not enough bytes in pieces": {
			input: &bencodeTorrent{
				Announce: "http://bttracker.debian.org:6969/announce",
				Info: bencodeInfo{
					Pieces:      "1234567890abcdefghijabcdef", // Only 26 bytes
					PieceLength: 262144,
					Length:      351272960,
					Name:        "debian-10.2.0-amd64-netinst.iso",
				},
			},
			output: TorrentFile{},
			fails:  true,
		},
	}

	for _, test := range tests {
		to, err := test.input.toTorrentFile()
		if test.fails {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, test.output, to)
	}
}
