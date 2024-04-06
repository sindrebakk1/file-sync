package globalenums

//type StatusCode int

//const (
//	OK             StatusCode = 200
//	Created                   = 201
//	InvalidMessage            = 400
//	Unauthorized              = 401
//	InternalError             = 500
//)
//
//func (c StatusCode) String() string {
//	return [...]string{"OK", "InvalidMessage", "InternalError"}[c]
//}

type AuthResult int

const (
	Authenticated AuthResult = iota
	NewUser
	AuthFailed
)

func (c AuthResult) String() string {
	return [...]string{"Authenticated", "Unauthorized", "NewUser"}[c]
}

type MessageType int

const (
	Exit MessageType = iota
	Error
	Authentication
	FileInfo
	FileInfoResponse
	UpdateFile
	UpdateFileResponse
	DownloadFile
	DownloadFileResponse
	FileInfoMapRequest
	FileInfoMapResponse
)

func (c MessageType) String() string {
	return [...]string{
		"Exit",
		"Error",
		"Authentication",
		"FileInfo",
		"FileInfoResponse",
		"UpdateRequest",
		"UpdateResponse",
		"DownloadRequest",
		"DownloadResponse",
		"FileInfoMapRequest",
		"FileInfoMapResponse",
	}[c]
}
