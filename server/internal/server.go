package internal

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

	commandIndex           = 0
	userIndex              = 1
	miniServerAddressIndex = 1
	filesStartIndex        = 2
)

// TorrentServer is a struct that contains:
//     - port                - the port on which the server listens
//     - usedUsernames       - a map whose keys are usernames(strings) that are already being used and values are the addresses of the clients that ue them(*strings)
//     - usedUsernamesMutex  - a Mutex that is used for working with "usedUsernames"
//     - clients             - a map whose keys are user addresses(string) and values are
//     - clientsMutex        -
//     - files               -
//     - filesMutex          -
type TorrentServer struct {
	port               string
	usedUsernames      map[string]*string
	usedUsernamesMutex sync.RWMutex
	clients            map[string]*Client
	clientsMutex       sync.RWMutex
	files              map[string]map[string]struct{}
	filesMutex         sync.RWMutex
}

func (t *TorrentServer) listUsersAndTheirAddresses() string {
	var sb strings.Builder

	sb.WriteString("list-users:")

	t.clientsMutex.RLock()
	defer t.clientsMutex.RUnlock()

	for _, info := range t.clients {
		if info.miniServerAddress != "" && info.username != "" {
			sb.WriteString(info.username + " - " + info.miniServerAddress + ";")
		}
	}

	return sb.String()
}

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

func (t *TorrentServer) checkIfUsernameIsUsedByDifferentAddressAndAddItOtherwise(senderAddress, username string) bool {
	t.usedUsernamesMutex.Lock()
	defer t.usedUsernamesMutex.Unlock()

	if user, ok := t.usedUsernames[username]; ok {
		if *user == senderAddress {
			return false
		}
		return true
	}

	t.usedUsernames[username] = &senderAddress
	return false
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

func (t *TorrentServer) unregisterFiles(username string, files ...string) {
	t.filesMutex.Lock()
	defer t.filesMutex.Unlock()

	for _, fileToDelete := range files {
		trimmedFileToDelete := strings.ReplaceAll(fileToDelete, `"`, "")
		delete(t.files[username], trimmedFileToDelete)
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

func (t *TorrentServer) registerFiles(username string, files ...string) {
	t.filesMutex.Lock()
	defer t.filesMutex.Unlock()

	if _, ok := t.files[username]; !ok {
		t.files[username] = make(map[string]struct{})
	}

	for _, fileToAdd := range files {
		fileToAdd = strings.ReplaceAll(fileToAdd, `"`, "")
		t.files[username][fileToAdd] = struct{}{}
	}

}

func (t *TorrentServer) registerFilesCommandHelper(senderAddress, username string, files ...string) string {
	if t.checkIfUsernameIsUsedByDifferentAddressAndAddItOtherwise(senderAddress, username) {
		return fmt.Sprintf("Another user has already registered as %s.", username)
	}

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
	rw.WriteString("list-files:" + t.listFiles() + "\n")
}

func (t *TorrentServer) handleListUsersCommand(rw *bufio.ReadWriter) {
	rw.WriteString(t.listUsersAndTheirAddresses() + "\n")
}

func (t *TorrentServer) handleRegisterMiniServerCommand(rw *bufio.ReadWriter, clientAddress, miniServerAddress string) {
	t.registerMiniServer(clientAddress, miniServerAddress)
	rw.WriteString("Successfully registered miniServerAddress." + "\n")
}

func (t *TorrentServer) registerClient(address string) {
	t.clientsMutex.Lock()
	defer t.clientsMutex.Unlock()

	t.clients[address] = CreateEmptyClient()
}

func (t *TorrentServer) getUsernameFor(clientAddress string) (string, error) {
	t.clientsMutex.Lock()
	defer t.clientsMutex.Unlock()

	if client, ok := t.clients[clientAddress]; ok {
		return client.username, nil
	}

	return "", fmt.Errorf("There is no such username")
}

func (t *TorrentServer) deleteFilesFor(username string) {
	t.filesMutex.Lock()
	defer t.filesMutex.Unlock()

	delete(t.files, username)
}

func (t *TorrentServer) disconnect(clientAddress string) {
	username, err := t.getUsernameFor(clientAddress)

	if err == nil {
		t.usedUsernamesMutex.Lock()
		defer t.usedUsernamesMutex.Unlock()

		delete(t.usedUsernames, username)

		t.clientsMutex.Lock()
		defer t.clientsMutex.Unlock()
		delete(t.clients, clientAddress)

		t.deleteFilesFor(username)
	}
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
			t.disconnect(clientAddress)
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
			}
			readerWriter.Flush()
		}
	}
}

// CreateNewServer is a factory method that:
//    - accepts
//         - port - a string representation of the port on which the server will listen
//    - creates and returns
//         - a pointer to TorrentServer struct
func CreateNewServer(port string) *TorrentServer {
	return &TorrentServer{
		port:          port,
		usedUsernames: make(map[string]*string),
		clients:       make(map[string]*Client),
		files:         make(map[string]map[string]struct{}),
	}
}

// Start is a function that:
//    1. Creates a listener using the port from TorrentSerevr
//    2. Accepts and handles connections
func (t *TorrentServer) Start() error {
	listener, err := net.Listen(protocol, ":"+t.port)
	if err != nil {
		return fmt.Errorf("Error starting server on port %s. %w", t.port, err)
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
