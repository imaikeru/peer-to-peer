package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

const (
	protocol = "tcp"
	host     = "localhost"
	port     = "13337"

	commandIndex           = 0
	userIndex              = 1
	miniServerAddressIndex = 1
	filesStartIndex        = 2

	commandList = "list-files;" +
		"download user \"path to file on user\" \"path to save\";" +
		"register user \"file1\" \"file2\" \"file3\" …. \"fileN\";" +
		"unregister user \"file1\" \"file2\" \"file3\" …. \"fileN\";"
)

// client struct containing message connection, mini server address used for donwloading from, username
type Client struct {
	// messagingConnection net.Conn
	miniServerAddress string
	username          string
}

func createEmptyClient() *Client {
	return &Client{
		miniServerAddress: "",
		username:          "",
	}
}

// client struct containing listener, clients and files and the respective mutex for them
type TorrentServer struct {
	// map[conn.remoteAddr().string()]Client{messagingConnection, miniServerAddress, username}
	// map[username]map[fileName]

	port         string
	clients      map[string]*Client
	clientsMutex sync.RWMutex
	files        map[string]map[string]struct{}
	filesMutex   sync.RWMutex
}

// function to list all users and their addresses - called when client asks for it
func (t *TorrentServer) listUsersAndTheirAddresses() string {
	var sb strings.Builder

	t.clientsMutex.RLock()
	defer t.clientsMutex.RUnlock()

	for _, info := range t.clients {
		if info.miniServerAddress != "" && info.username != "" {
			sb.WriteString(info.username + " - " + info.miniServerAddress + ";")
		}
	}

	return sb.String()
}

// function to list all files and their respective owners - called when client asks for it
func (t *TorrentServer) listFiles() string {
	var sb strings.Builder

	t.filesMutex.RLock()
	defer t.filesMutex.RUnlock()

	for username, filePaths := range t.files {
		for filePath := range filePaths {
			sb.WriteString(username + " : " + filePath + ";")
		}
	}

	return sb.String()
}

func (t *TorrentServer) validateAndUpdateUsername(senderAddress, username string) (string, bool) {
	t.clientsMutex.Lock()
	defer t.clientsMutex.Unlock()

	client := t.clients[senderAddress]

	if client.username != username {
		if client.username == "" {
			client.username = username
			return client.username, true
		}
		return client.username, false
	}
	return client.username, true
}

// function to unregister files from a given user, meaning they wont be available for downloading afterwards ( wont be listed at least )
func (t *TorrentServer) unregisterFiles(username string, files ...string) {
	t.filesMutex.Lock()
	defer t.filesMutex.Unlock()

	for _, fileToDelete := range files {
		delete(t.files[username], fileToDelete)
	}
}

func (t *TorrentServer) unregisterFilesCommandHelper(senderAddress, username string, files ...string) string {
	if registeredAs, valid := t.validateAndUpdateUsername(senderAddress, username); !valid {
		return fmt.Sprintf("You have already registered as %s.", registeredAs)
	}

	t.unregisterFiles(username, files...)

	return "Successfully unregistered files."
}

func (t *TorrentServer) handleUnregisterFilesCommand(rw *bufio.ReadWriter, senderAddress, username string, files ...string) {
	rw.WriteString(t.unregisterFilesCommandHelper(senderAddress, username, files...) + "\n")
}

// function that registers files for a given user, meaning they will be available for downloading afterwards
func (t *TorrentServer) registerFiles(username string, files ...string) {
	t.filesMutex.Lock()
	defer t.filesMutex.Unlock()

	if _, ok := t.files[username]; !ok {
		t.files[username] = make(map[string]struct{})
	}

	for _, fileToAdd := range files {
		t.files[username][fileToAdd] = struct{}{}
	}

}

func (t *TorrentServer) registerFilesCommandHelper(senderAddress, username string, files ...string) string {
	if registeredAs, valid := t.validateAndUpdateUsername(senderAddress, username); !valid {
		return fmt.Sprintf("You have already registered as %s.", registeredAs)
	}

	t.registerFiles(username, files...)

	return "Successfully registered files."
}

func (t *TorrentServer) handleRegisterFilesCommand(rw *bufio.ReadWriter, senderAddress, username string, files ...string) {
	rw.WriteString(t.registerFilesCommandHelper(senderAddress, username, files...) + "\n")
}

func (t *TorrentServer) registerMiniServer(clientAddress, miniServerAddress string) {
	t.clientsMutex.Lock()
	defer t.clientsMutex.Unlock()

	t.clients[clientAddress].miniServerAddress = miniServerAddress
}

func (t *TorrentServer) handleListFilesCommand(rw *bufio.ReadWriter) {
	rw.WriteString(t.listFiles() + "\n")
}

func (t *TorrentServer) handleListUsersCommand(rw *bufio.ReadWriter) {
	rw.WriteString(t.listUsersAndTheirAddresses() + "\n")
}

func (t *TorrentServer) handleRegisterMiniServerCommand(rw *bufio.ReadWriter, clientAddress, miniServerAddress string) {
	t.registerMiniServer(clientAddress, miniServerAddress)
	rw.WriteString("Successfully registered miniServerAddress." + "\n")
}

func (t *TorrentServer) handleWrongCommand(rw *bufio.ReadWriter) {
	rw.WriteString("Wrong command. Choose between:;" + commandList + "\n")
}

func (t *TorrentServer) registerClient(address string) {
	t.clientsMutex.Lock()
	defer t.clientsMutex.Unlock()

	t.clients[address] = createEmptyClient()
}

func (t *TorrentServer) getUsernameFor(clientAddress string) string {
	t.clientsMutex.Lock()
	defer t.clientsMutex.Unlock()

	return t.clients[clientAddress].username
}

func (t *TorrentServer) deleteFilesFor(username string) {
	t.filesMutex.Lock()
	defer t.filesMutex.Unlock()

	delete(t.files, username)
}

func (t *TorrentServer) disconnect(clientAddress string) {
	username := t.getUsernameFor(clientAddress)
	t.deleteFilesFor(username)
}

func (t *TorrentServer) handleConnection(conn net.Conn) {
	clientAddress := conn.RemoteAddr().String()
	log.Println("Accepted connection from: ", clientAddress)

	t.registerClient(clientAddress)

	readerWriter := bufio.NewReadWriter(bufio.NewReaderSize(conn, 4096), bufio.NewWriterSize(conn, 4096))
	defer conn.Close()

loop:
	for {
		readerWriter.Flush()
		if data, err := readerWriter.ReadString('\n'); err != nil {
			log.Println(err)
			break
		} else {
			readerWriter.Flush()
			fmt.Print("From client ", clientAddress, ": ", data)
			parsedCommand := strings.Fields(data)

			switch parsedCommand[commandIndex] {
			case "disconnect":
				t.disconnect(clientAddress)
				break loop
			case "unregister":
				t.handleUnregisterFilesCommand(readerWriter, clientAddress, parsedCommand[userIndex], parsedCommand[filesStartIndex:]...)
			case "register-miniserver":
				t.handleRegisterMiniServerCommand(readerWriter, clientAddress, parsedCommand[miniServerAddressIndex])
			case "register":
				t.handleRegisterFilesCommand(readerWriter, clientAddress, parsedCommand[userIndex], parsedCommand[filesStartIndex:]...)
			case "list-files":
				t.handleListFilesCommand(readerWriter)
			case "list-users":
				t.handleListUsersCommand(readerWriter)
			default:
				t.handleWrongCommand(readerWriter)
			}
			readerWriter.Flush()
		}
	}
}

func createNewServer(port string) *TorrentServer {
	return &TorrentServer{
		port:    port,
		clients: make(map[string]*Client),
		files:   make(map[string]map[string]struct{}),
	}
}

func (t *TorrentServer) start() error {
	listener, err := net.Listen(protocol, ":"+t.port)
	if err != nil {
		return fmt.Errorf("Error starting server on port %s. %w", port, err)
	}

	defer listener.Close()

	log.Printf("Server Started. Listening on port %s", t.port)
	for {
		if conn, err := listener.Accept(); err != nil {
			log.Println("Error accepting connection")
		} else {
			go t.handleConnection(conn)
		}
	}
}

func main() {
	// listener, err := net.Listen(protocol, host+":"+port)

	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// defer listener.Close()

	// log.Println("Server Started")
	// for {
	// 	if conn, err := listener.Accept(); err != nil {
	// 		log.Println("Error accepting connection")
	// 	} else {
	// 		go handleConnection(conn)
	// 	}
	// }
	ts := createNewServer(port)

	if err := ts.start(); err != nil {
		log.Fatalln(err)
	}

}
