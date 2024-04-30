package enums

type MessageType uint8

const (
	Status MessageType = iota
	Upload
	Download
	Delete
	List
	Auth
	Error
)

func (a MessageType) String() string {
	return [...]string{"Status", "Upload", "Download", "Delete", "List", "Auth", "Error"}[a]
}

type Done uint8

const (
	No Done = iota
	Yes
)

func (d Done) String() string {
	return [...]string{"No", "Yes"}[d]
}

type Sender uint8

const (
	Client Sender = iota
	Server
)

func (s Sender) String() string {
	return [...]string{"Client", "Server"}[s]
}
