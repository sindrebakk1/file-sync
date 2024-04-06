package globalmodels

import "file-sync/pkg/globalenums"

// Message is the base struct for all messages.
type Message struct {
	MessageType globalenums.MessageType
	Payload     interface{}
}

// ChallengeMessage contains a challenge for the client.
type ChallengeMessage struct {
	Message
	Payload []byte
}

// ChallengeResponse is the payload for a ChallengeResponseMessage.
type ChallengeResponse struct {
	Response []byte
	User     string
}

// ChallengeResponseMessage contains information about a challenge response message.
type ChallengeResponseMessage struct {
	Message
	Payload ChallengeResponse
}

// AuthResponseMessage contains information about an authentication response message.
type AuthResponseMessage struct {
	Message
	Payload globalenums.AuthResult
}

// NewUserResponseMessage contains a shared key for a new user.
type NewUserResponseMessage struct {
	Message
	Payload []byte
}

// SyncRequest is the payload for a SyncRequestMessage.
type SyncRequest struct {
	Hash string
	Info File
}

// SyncRequestMessage contains information about a file message.
type SyncRequestMessage struct {
	Message
	Payload SyncRequest
}

// SyncResponse is the payload for a SyncResponseMessage.
type SyncResponse struct {
	Info File
}

// SyncResponseMessage contains information about a file message.
type SyncResponseMessage struct {
	Message
	Payload SyncResponse
}

// FileChunkMessage contains information about a file chunk message.
type FileChunkMessage struct {
	Message
	Payload []byte
}

// ErrorMessage contains information about an error message.
type ErrorMessage struct {
	Message
	Payload string
}
