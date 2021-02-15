package main

import (
	"flag"
	"fmt"
	"log"
)

func main() {

	filePathPtr := flag.String("file_path", "/path/to/file/where/users/and/their/addresses/are/saved", "string")

	flag.Parse()

	fmt.Print(*filePathPtr)

	client := InitializeClient(*filePathPtr, "13337")

	if err := client.start(); err != nil {
		log.Fatalln(err)
	}
}
