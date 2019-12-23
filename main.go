package main

import (
	"log"
	"os"

	"github.com/veggiedefender/torrent-client/torrent"
)

func main() {
	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	t, err := torrent.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	err = t.Download()
	if err != nil {
		log.Fatal(err)
	}
}
