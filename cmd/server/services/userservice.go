package services

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type UserService interface {
	// Create creates a new user with the given username and shared key.
	Create(username string, sharedKey []byte) (err error)
	// GetSharedKey returns the shared key of the user with the given username.
	GetSharedKey(username string) (sharedKey []byte, found bool)
	// Delete deletes the user with the given username.
	Delete(username string)
	// GetFileService returns the file service of the user with the given username.
	GetFileService(username string) (fileService FileService, err error)
}

type concreteUserService struct {
	userMap            map[string][]byte
	fileServiceFactory FileServiceFactory
}

func NewUserService(fileServiceFactory FileServiceFactory) UserService {
	return &concreteUserService{
		make(map[string][]byte),
		fileServiceFactory,
	}
}

func (u *concreteUserService) Create(username string, sharedKey []byte) (err error) {
	_, exists := u.userMap[username]
	if exists {
		return fmt.Errorf("user already exists")
	}
	u.userMap[username] = sharedKey
	return nil
}

func (u *concreteUserService) GetSharedKey(username string) (sharedKey []byte, found bool) {
	sharedKey, found = u.userMap[username]
	return sharedKey, found
}

func (u *concreteUserService) Delete(username string) {
	delete(u.userMap, username)
}

func (u *concreteUserService) GetFileService(username string) (fileService FileService, err error) {
	sharedKey, found := u.GetSharedKey(username)
	if !found {
		return nil, fmt.Errorf("user not found")
	}

	return u.fileServiceFactory.NewFileService(generateHash(sharedKey))
}

func generateHash(data []byte) (hash string) {
	hasher := sha256.New()
	hasher.Write(data)
	hash = hex.EncodeToString(hasher.Sum(nil))

	return hash
}
