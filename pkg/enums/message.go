package enums

type MessageType uint8

const (
	Status MessageType = iota
	Upload
	Download
	Delete
	Chunk
	List
	Auth
	Error
	Cancel
	Echo
)

func (a MessageType) String() string {
	return [...]string{"Status", "Upload", "Download", "Delete", "Chunk", "List", "Auth", "Error", "Cancel", "Echo"}[a]
}

type Done uint8

const (
	Yes Done = iota
	No
)

func (d Done) String() string {
	return [...]string{"Yes", "No"}[d]
}

type Sender uint8

const (
	Server Sender = iota
	Client
)

func (s Sender) String() string {
	return [...]string{"Server", "Client"}[s]
}
