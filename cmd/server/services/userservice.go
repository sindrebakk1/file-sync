package services

type UserService interface {
	// Create creates a new user with the given username and shared key.
	Create(username string, sharedKey []byte) (err error)
	// GetSharedKey returns the shared key of the user with the given username.
	GetSharedKey(username string) (sharedKey []byte, found bool)
	// Delete deletes the user with the given username.
	Delete(username string) (err error)
	// GetFileService returns the file service of the user with the given username.
	GetFileService(username string) (fileService FileService, found bool)
}
