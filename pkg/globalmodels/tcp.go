package globalmodels

import (
	"file-sync/pkg/globalenums"
)

// ChallengeMessage contains a challenge for the client.
type ChallengeMessage struct {
	Payload []byte
}

// ChallengeResponse is the payload for a ChallengeResponseMessage.
type ChallengeResponse struct {
	Response []byte
	User     string
}

// ChallengeResponseMessage contains a challenge response for the server.
type ChallengeResponseMessage struct {
	Payload ChallengeResponse
}

// AuthResponseMessage contains the result of an authentication request.
type AuthResponseMessage struct {
	Payload globalenums.AuthResult
}

// SharedKeyMessage contains a shared key for a new user.
type SharedKeyMessage struct {
	Payload []byte
}

// Message is the base struct for all messages.
type Message struct {
	TransactionID string
	Payload       interface{}
}

type FileMapResponse struct {
	FileMap map[string]File
}

// SyncRequest is the payload for a SyncRequestMessage.
type SyncRequest struct {
	Hash string
	Info File
}

// SyncResponse is the payload for a SyncResponseMessage.
type SyncResponse struct {
	Info File
}

// DownloadRequest is the payload for a DownloadRequestMessage.
type DownloadRequest struct {
	Hash string
}

// UploadRequest is the payload for an upload request.
type UploadRequest struct {
	Hash     string
	Checksum string
}

// UploadResult is the payload for an upload success message.
type UploadResult struct {
	Hash    string
	Success bool
}

// FileChunk contains information about a file chunk message.
type FileChunk struct {
	Chunk []byte
	Done  bool
}

// CloseConnection is a message to close the connection.
type CloseConnection struct{}

// Error contains information about an error message.
type Error struct {
	Message string
}
