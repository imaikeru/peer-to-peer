package internal

// Client is a struct that contains:
//     - username          - string
//     - miniServerAddress - the address of the mini server, which is linked to the username
type Client struct {
	miniServerAddress string
	username          string
}

// CreateEmptyClient is a factory method that:
//    - creates and returns a pointer to an empty Client struct
func CreateEmptyClient() *Client {
	return &Client{
		miniServerAddress: "",
		username:          "",
	}
}
