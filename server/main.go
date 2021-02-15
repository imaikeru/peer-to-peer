package main

import (
	"log"

	"github.com/imaikeru/peer-to-peer/server/internal"
)

const port = "13337"

func main() {
	ts := internal.CreateNewServer(port)

	if err := ts.Start(); err != nil {
		log.Fatalln(err)
	}
}
