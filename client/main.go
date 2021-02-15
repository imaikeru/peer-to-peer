package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/imaikeru/peer-to-peer/client/internal"
)

func main() {

	filePathPtr := flag.String("file_path", "/path/to/file/where/users/and/their/addresses/are/saved", "string")

	flag.Parse()

	fmt.Print(*filePathPtr)

	client := internal.CreateNewClient(*filePathPtr, "13337")

	if err := client.Start(); err != nil {
		log.Fatalln(err)
	}
}
