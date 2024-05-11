package user

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"server/services/file"
)

type Service interface {
	// Create creates a new user with the given username and shared key.
	Create(username string, sharedKey []byte) (err error)
	// GetSharedKey returns the shared key of the user with the given username.
	GetSharedKey(username string) (sharedKey []byte, found bool)
	// Delete deletes the user with the given username.
	Delete(username string)
	// GetFileService returns the file service of the user with the given username.
	GetFileService(username string) (fileService file.Service, err error)
}

type concreteService struct {
	userMap            map[string][]byte
	fileServiceFactory file.Factory
}

func New(fileServiceFactory file.Factory) Service {
	return &concreteService{
		make(map[string][]byte),
		fileServiceFactory,
	}
}

func (u *concreteService) Create(username string, sharedKey []byte) (err error) {
	_, exists := u.userMap[username]
	if exists {
		return fmt.Errorf("user already exists")
	}
	u.userMap[username] = sharedKey
	return nil
}

func (u *concreteService) GetSharedKey(username string) (sharedKey []byte, found bool) {
	sharedKey, found = u.userMap[username]
	return sharedKey, found
}

func (u *concreteService) Delete(username string) {
	delete(u.userMap, username)
}

func (u *concreteService) GetFileService(username string) (fileService file.Service, err error) {
	sharedKey, found := u.GetSharedKey(username)
	if !found {
		return nil, fmt.Errorf("user not found")
	}

	return u.fileServiceFactory.New(generateHash(sharedKey))
}

func generateHash(data []byte) (hash string) {
	hasher := sha256.New()
	hasher.Write(data)
	hash = hex.EncodeToString(hasher.Sum(nil))

	return hash
}
