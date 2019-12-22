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

	to, err := torrent.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	to.Download()
}
