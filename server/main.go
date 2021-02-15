package main

import (
	"log"

	"github.com/imaikeru/peer-to-peer/server/server"
)

const port = "13337"

func main() {
	ts := server.CreateNewServer(port)

	if err := ts.Start(); err != nil {
		log.Fatalln(err)
	}
}
