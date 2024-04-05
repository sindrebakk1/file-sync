package globalmodels

type StatusCode int

const (
	OK             StatusCode = 200
	InvalidMessage            = 400
	InternalError             = 500
)

type Message struct {
	StatusCode StatusCode
}

// HandshakeMessage contains information about a handshake message.
type HandshakeMessage struct {
	Message
	ClientID  string
	ServerID  string
	SharedKey []byte
}

// HandshakeResponseMessage contains information about a handshake response message.
type HandshakeResponseMessage struct {
	Message
	ClientID string
	ServerID string
}

// FileInfoMessage contains information about a file message.
type FileInfoMessage struct {
	Message
	Hash     string
	FileInfo FileInfo
}

// FileChunkMessage contains information about a file chunk message.
type FileChunkMessage struct {
	Message
	chunk []byte
}
