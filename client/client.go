package main

import (
	"bufio"
	// "flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/imaikeru/peer-to-peer/client/validator"
)

const (
	userIndex             = 1
	pathToFileOnUserIndex = 2
	pathToSaveIndex       = 3

	commandsList = "Wrong command, choose between:\n" + "list-files\n" +
		"download user \"path to file on user\" \"path to save\"\n" +
		"register user \"file1\" \"file2\" \"file3\" …. \"fileN\"\n" +
		"unregister user \"file1\" \"file2\" \"file3\" …. \"fileN\"\n"
)

func (c *Client) getAddressToDownloadFrom(username string) (string, error) {
	c.fileMutex.Lock()
	defer c.fileMutex.Unlock()
	file, err := os.OpenFile(c.usersAndAddressesFileName, os.O_RDWR, 0777)
	if err != nil {
		log.Println("Error while opening file for reading.")
		return "", err
	}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		text := scanner.Text()
		if strings.Contains(text, username) {
			address := strings.Split(text, "-")[1]
			return address, nil
		}
	}

	file.Close()

	return "", fmt.Errorf("There is no record containing this username and its address")
}

func (c *Client) miniServerHandleDownloadRequest(conn net.Conn) {
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
					rw.Flush()
				}
			}
			fileToSend.Close()
		}
	}
}

func (c *Client) downloadFile(address *string, pathToFileOnUser, pathToSave string) {
	fmt.Println("Address is " + *address)
	split := strings.SplitN(*address, ":", 2)
	downloadConnection, connectToMiniserverErr := net.Dial("tcp", split[0]+":"+split[1])
	if connectToMiniserverErr != nil {
		log.Printf("Failed to connect to miniserver with address: %s", *address)
		log.Printf(connectToMiniserverErr.Error())
	} else {

		requestWriter := bufio.NewWriter(downloadConnection)
		requestWriter.WriteString(pathToFileOnUser + "\n")
		requestWriter.Flush()

		newFile, createFileError := os.Create(pathToSave)
		if createFileError != nil {
			log.Printf("Could not create file with name %s", pathToSave)
		} else {
			fileReadWriter := bufio.NewReadWriter(bufio.NewReader(downloadConnection), bufio.NewWriter(newFile))
			for {
				buf := make([]byte, 4096)
				for {
					if bytesRead, readingErr := fileReadWriter.Read(buf); readingErr != nil {
						if readingErr != io.EOF {
							log.Printf("Error reading file %s.", pathToFileOnUser)
						}
						break
					} else {
						fileReadWriter.Write(buf[0:bytesRead])
						fileReadWriter.Flush()
					}
				}

			}
		}
		newFile.Close()
	}
}

func (c *Client) getUsersInformationFromServerPeriodically(conn net.Conn) {
	serverWriter := bufio.NewWriter(conn)

	for {
		serverWriter.Flush()
		_, err := serverWriter.WriteString("list-users" + "\n")
		serverWriter.Flush()

		if err != nil {
			log.Printf("Error occurred while trying to write to server, %s", err.Error())
		}
		time.Sleep(30 * time.Second)
	}
}

func (c *Client) updateUsersAndAddresses(newData string) {
	data := strings.SplitN(newData, ":", 2)
	newInfo := ""
	if len(data) == 2 {
		newInfo = data[1]
		newInfo = strings.ReplaceAll(newInfo, ";", "\n")
	}

	c.fileMutex.Lock()
	defer c.fileMutex.Unlock()

	if file, err := os.Create(c.usersAndAddressesFileName); err != nil {
		log.Printf("Error when attempting to open %s to update users data.", c.usersAndAddressesFileName)
	} else {
		writer := bufio.NewWriter(file)
		if _, writingError := writer.WriteString(newInfo); writingError != nil {
			log.Printf("%s", writingError.Error())
		}
		writer.Flush()
		file.Close()
	}
}

func (c *Client) operateMiniServer(miniServer net.Listener) {
	for {
		if conn, err := miniServer.Accept(); err != nil {
			log.Println("Error accepting connection.")
		} else {
			go c.miniServerHandleDownloadRequest(conn)
		}
	}
}

// Client is a struct that contains:
// 	  - usersAndAddressesFileName - path to file which will contain the information about other users and their addresses that are connected to the main server
//    - fileMutex                 - a Mutex that is used for working with "usersAndAddressesFileName"
//    - centralServerport         - the port of the central server, to which the client connects
//    - validator                 - used for validating the user commands
type Client struct {
	fileMutex                 sync.Mutex
	usersAndAddressesFileName string
	centralServerPort         string
	validator                 *validator.Validator
}

// CreateNewClient is a factory function that:
//   - accepts:
//        - usersAndAddressesFileName - path to file which will contain the information about other users and their addresses that are connected to the main server
//        - centralServerPort         - the port of the central server, to which the client will connect
//   - creates and returns:
//        - a pointer to Client struct
func CreateNewClient(usersAndAddressesFileName, centralServerPort string) *Client {
	return &Client{
		usersAndAddressesFileName: usersAndAddressesFileName,
		centralServerPort:         centralServerPort,
		validator:                 validator.CreateValidator(),
	}
}

func (c *Client) parseListFiles(response string) *string {
	splitResponse := strings.SplitN(response, ":", 2)
	data := splitResponse[1]
	data = strings.Trim(data, ";")
	data = strings.ReplaceAll(data, ";", "\n")
	return &data
}

func (c *Client) start() error {
	server, err := net.Dial("tcp", ":13337")
	if err != nil {
		return fmt.Errorf("Failed to connect to server. %w", err)
	}
	consoleToServerRw := bufio.NewReadWriter(bufio.NewReaderSize(os.Stdin, 4096), bufio.NewWriterSize(server, 4096))

	fmt.Println("Server address " + server.RemoteAddr().String())

	miniServer, errServerCreated := net.Listen("tcp", "localhost:0")
	if errServerCreated != nil {
		return fmt.Errorf("Could not initialize MiniServer. %w", errServerCreated)
	}

	fmt.Println(miniServer.Addr().String())
	go c.operateMiniServer(miniServer)

	miniServerAddress := miniServer.Addr().String()
	log.Printf("MiniServer started. Listening on: %s", miniServerAddress)
	consoleToServerRw.Flush()
	consoleToServerRw.WriteString("register-miniserver " + miniServerAddress + "\n")
	consoleToServerRw.Flush()

	go c.getUsersInformationFromServerPeriodically(server)

	go func() {
		for {
			consoleToServerRw.Flush()
			if request, err := consoleToServerRw.ReadString('\n'); err != nil {
				log.Println("Failed to read from stdin.")
				break
			} else {
				if !c.validator.Validate(strings.ReplaceAll(request, "\n", "")) {
					log.Printf(commandsList)
				} else {
					if strings.Contains(request, "download") {
						splitRequest := strings.Fields(request)
						username := splitRequest[userIndex]
						pathToFileOnUser := splitRequest[pathToFileOnUserIndex]
						pathToSave := splitRequest[pathToSaveIndex]
						addressToDownloadFrom, userErr := c.getAddressToDownloadFrom(username)
						if userErr != nil {
							log.Printf("The user %s is not an active one. %s", username, userErr.Error())
						} else {
							go c.downloadFile(&addressToDownloadFrom, pathToFileOnUser, pathToSave)
						}
					} else {
						consoleToServerRw.Flush()
						if _, err2 := consoleToServerRw.WriteString(request); err2 != nil {
							log.Println(err2)
						}
						consoleToServerRw.Flush()
					}
				}
			}
		}
	}()

	serverReader := bufio.NewReader(server)
	for {
		response, err := serverReader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				return fmt.Errorf("Failed to read from server. %w", err)
			}
			log.Println("Disconnected from server.")
			return nil
		}
		if strings.Contains(response, "list-users:") {
			go c.updateUsersAndAddresses(response)
		} else if strings.Contains(response, "list-files:") {
			log.Printf("\n" + *c.parseListFiles(response))
		} else {
			log.Printf("From server: %s", response)
		}

	}
}

// func main() {

// 	filePathPtr := flag.String("file_path", "/path/to/file/where/users/and/their/addresses/are/saved", "string")

// 	flag.Parse()

// 	fmt.Print(*filePathPtr)

// 	client := CreateNewClient(*filePathPtr, "13337")

// 	if err := client.start(); err != nil {
// 		log.Fatalln(err)
// 	}
// }
