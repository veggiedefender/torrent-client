package main

import (
	"log"
	"os"

	"github.com/veggiedefender/torrent-client/torrent"
)

func main() {
	inPath := os.Args[1]
	outPath := os.Args[2]

	inFile, err := os.Open(inPath)
	if err != nil {
		log.Fatal(err)
	}
	defer inFile.Close()

	t, err := torrent.Open(inFile)
	if err != nil {
		log.Fatal(err)
	}
	buf, err := t.Download()
	if err != nil {
		log.Fatal(err)
	}

	outFile, err := os.Create(outPath)
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()
	_, err = outFile.Write(buf)
	if err != nil {
		log.Fatal(err)
	}
}
