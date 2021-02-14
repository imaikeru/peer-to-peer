package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

// read file to see what address you need to download from
func getAddressToDownloadFrom(username string) (string, error) {
	file, err := os.Open("file.txt")
	if err != nil {
		log.Println("Error while opening file")
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		text := scanner.Text()
		if strings.Contains(text, username) {
			address := strings.Split(text, "-")[1]
			return address, nil
		}
	}

	return "", fmt.Errorf("There is no record containing this username and its address")
}

// download file
func handleDownloadRequest(conn net.Conn) {
	log.Println("Accepted download request from: ", conn.RemoteAddr().String())

	defer conn.Close()
	messageReader := bufio.NewReader(conn)

	if fileToDownloadMessage, err := messageReader.ReadString('\n'); err != nil {
		log.Println("Error when reading message from connection.")
	} else {
		if fileToSend, fileErr := os.Open(fileToDownloadMessage); fileErr != nil {
			log.Printf("Could not open file %s for reading", fileToDownloadMessage)
		} else {
			buf := make([]byte, 4096)
			rw := bufio.NewReadWriter(bufio.NewReaderSize(fileToSend, 4096), bufio.NewWriterSize(conn, 4096))
			for {

				if bytesRead, readingErr := rw.Read(buf); readingErr != nil {
					if readingErr != io.EOF {
						log.Printf("Error reading file %s.", fileToDownloadMessage)
					}
					break
				} else {
					rw.Write(buf[0:bytesRead])
				}
			}

		}
	}
}

// start mini server that listens to download requests
func operateMiniServer(miniServer net.Listener) {
	for {
		if conn, err := miniServer.Accept(); err != nil {
			log.Println("Error accepting connection.")
		} else {
			go handleDownloadRequest(conn)
		}
	}
}

type Client struct {
	usersAndAddressesFileName string
	centralServerPort         string
}

func initializeClient(usersAndAddressesFileName, centralServerPort string) *Client {
	return &Client{
		usersAndAddressesFileName: usersAndAddressesFileName,
		centralServerPort:         centralServerPort,
	}
}

func (c *Client) start() error {
	if conn, err := net.Dial("tcp", ":13337"); err != nil {
		return fmt.Errorf("Failed to connect to server. %w", err)
	} else {
		consoleToServerRw := bufio.NewReadWriter(bufio.NewReaderSize(os.Stdin, 4096), bufio.NewWriterSize(conn, 4096))

		if miniServer, errServerCreated := net.Listen("tcp", "localhost:0"); errServerCreated != nil {
			fmt.Errorf("Could not initialize MiniServer. %w", errServerCreated)
		} else {
			miniServerAddress := miniServer.Addr().String()
			log.Printf("MiniServer started. Listening on: %s", miniServerAddress)
			consoleToServerRw.Flush()
			consoleToServerRw.WriteString("register-miniserver " + miniServerAddress)
			consoleToServerRw.Flush()
			go func() {
				for {
					if request, err := consoleToServerRw.ReadString('\n'); err != nil {
						log.Println("Failed to read from stdin.")
						break
					} else {
						if written, err2 := consoleToServerRw.WriteString(request); err2 != nil {
							log.Println(err2)
						} else {
							log.Println(written)
						}
					}
					consoleToServerRw.Flush()
				}
			}()

			serverReader := bufio.NewReader(conn)
			for {
				if response, err := serverReader.ReadString('\n'); err != nil {
					return fmt.Errorf("Failed to read from server. %w", err)
				} else {
					fmt.Println("From server: ", response)
				}
			}
		}
	}

}

func main() {

	// if conn, err := net.Dial("tcp", ":13337"); err != nil {
	// 	log.Println("Failed to connect to server, exitting.")
	// } else {
	// 	consoleToServerRw := bufio.NewReadWriter(bufio.NewReaderSize(os.Stdin, 4096), bufio.NewWriterSize(conn, 4096))

	// 	if miniServer, errServerCreated := net.Listen("tcp", "localhost:0"); errServerCreated != nil {
	// 		log.Println("Could not initialize MiniServer. Exitting.")
	// 	} else {
	// 		miniServerAddress := miniServer.Addr().String()
	// 		log.Printf("MiniServer started. Listening on: %s", miniServerAddress)
	// 		consoleToServerRw.Flush()
	// 		consoleToServerRw.WriteString("register-miniserver " + miniServerAddress)
	// 		consoleToServerRw.Flush()
	// 		go func() {
	// 			for {
	// 				if request, err := consoleToServerRw.ReadString('\n'); err != nil {
	// 					log.Println("Failed to read from stdin.")
	// 					break
	// 				} else {
	// 					if written, err2 := consoleToServerRw.WriteString(request); err2 != nil {
	// 						log.Println(err2)
	// 					} else {
	// 						log.Println(written)
	// 					}
	// 				}
	// 				consoleToServerRw.Flush()
	// 			}
	// 		}()

	// 		serverReader := bufio.NewReader(conn)
	// 		for {
	// 			if response, err := serverReader.ReadString('\n'); err != nil {
	// 				log.Println("Failed to read from server.")
	// 				break
	// 			} else {
	// 				fmt.Println("From server: ", response)
	// 			}
	// 		}
	// 	}
	// }
}
