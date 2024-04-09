package auth

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"file-sync/pkg/globalenums"
	"file-sync/pkg/globalmodels"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"server/services"
)

type Authenticator interface {
	AuthenticateClient() (userName string, err error)
	GetFileService() (services.FileService, error)
}

type Config struct {
	ChallengeLen int
}

type concreteAuthenticator struct {
	conn          net.Conn
	userService   services.UserService
	config        *Config
	authenticated bool
	userName      string
}

func NewAuthenticator(conn net.Conn, userService services.UserService, config *Config) Authenticator {
	return &concreteAuthenticator{
		conn,
		userService,
		config,
		false,
		"",
	}
}

// AuthenticateClient authenticates the client, creating a new user if necessary.
func (a *concreteAuthenticator) AuthenticateClient() (userName string, err error) {
	var challenge []byte
	challenge, err = generateChallenge(a.config.ChallengeLen)
	if err != nil {
		return "", err
	}
	challengeMessage := globalmodels.ChallengeMessage{
		Payload: challenge,
	}

	// Send the challenge to the client.
	err = a.sendMessage(&challengeMessage)
	if err != nil {
		return "", err
	}

	// Receive the response from the client.
	var challengeResponseMessage globalmodels.ChallengeResponseMessage
	err = a.receiveMessage(&challengeResponseMessage)
	if err != nil {
		return "", err
	}
	a.userName = challengeResponseMessage.Payload.User
	challengeResponse := challengeResponseMessage.Payload.Response

	// Get the shared key for the user.
	sharedKey, found := a.userService.GetSharedKey(userName)
	// If the user is not found, create a new user and receive the shared key.
	if !found {
		newUserMessage := globalmodels.AuthResponseMessage{
			Payload: globalenums.NewUser,
		}
		err = a.sendMessage(&newUserMessage)
		if err != nil {
			return "", err
		}
		// Receive the shared key from the client.
		err = a.receiveMessage(&sharedKey)
		if err != nil {
			return "", err
		}
		err = a.userService.Create(userName, sharedKey)
		if err != nil {
			return "", err
		}
	}

	// Compare the expected response with the received response.
	var expectedResponse []byte
	expectedResponse, err = calculateResponse(challenge, sharedKey)
	if !bytes.Equal(expectedResponse, challengeResponse) {
		// Send the authenticated message to the client.
		authFailedMessage := globalmodels.AuthResponseMessage{
			Payload: globalenums.AuthFailed,
		}
		err = a.sendMessage(&authFailedMessage)
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf("challenge failed")
	}

	// Send the authenticated message to the client.
	authenticatedMessage := globalmodels.AuthResponseMessage{
		Payload: globalenums.Authenticated,
	}
	err = a.sendMessage(&authenticatedMessage)
	if err != nil {
		return "", err
	}

	a.authenticated = true

	return a.userName, nil
}

// GetFileService returns the file service of the authenticated user.
func (a *concreteAuthenticator) GetFileService() (services.FileService, error) {
	if !a.authenticated {
		return nil, fmt.Errorf("not authenticated")
	}
	return a.userService.GetFileService(a.userName)
}

// generateChallenge generates a random challenge of the specified length.
func generateChallenge(length int) (challenge []byte, err error) {
	challenge = make([]byte, base64.StdEncoding.EncodedLen(length))
	log.Debugf("Generating challenge with lenght %d", length)
	randomBytes := make([]byte, length)
	_, err = rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	base64.StdEncoding.Encode(challenge, randomBytes)
	return challenge, nil
}

// calculateResponse calculates the expected response to the challenge using the shared key.
func calculateResponse(challenge []byte, sharedKey []byte) (response []byte, err error) {
	challengeBytes := make([]byte, base64.StdEncoding.DecodedLen(len(challenge)))
	_, err = base64.StdEncoding.Decode(challengeBytes, challenge)
	if err != nil {
		return nil, err
	}

	// Calculate the HMAC-SHA256 hash of the challenge using the shared secret
	mac := hmac.New(sha256.New, sharedKey)
	_, err = mac.Write(challengeBytes)
	if err != nil {
		return nil, err
	}

	return mac.Sum(nil), nil
}

func (a *concreteAuthenticator) sendMessage(message interface{}) error {
	encoder := gob.NewEncoder(a.conn)
	return encoder.Encode(message)
}

func (a *concreteAuthenticator) receiveMessage(message interface{}) error {
	decoder := gob.NewDecoder(a.conn)
	return decoder.Decode(message)
}
