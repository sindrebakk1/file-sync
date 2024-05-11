package auth

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"filesync/enums"
	"filesync/models"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"server/services/file"
	"server/services/user"
)

type Service interface {
	AuthenticateClient(net.Conn) error
	GetFileService() (file.Service, error)
	IsAuthenticated() bool
	GetUsername() string
}

type Config struct {
	ChallengeLen int
}

type concreteService struct {
	userService   user.Service
	config        *Config
	authenticated bool
	userName      string
}

func New(userService user.Service, config *Config) Service {
	return &concreteService{
		userService,
		config,
		false,
		"",
	}
}

// AuthenticateClient authenticates the client, creating a new user if necessary.
func (a *concreteService) AuthenticateClient(conn net.Conn) (err error) {
	var challenge []byte
	challenge, err = generateChallenge(a.config.ChallengeLen)
	log.Debugf("Generated challenge: %s", challenge)
	if err != nil {
		return err
	}
	challengeMessage := models.Message{
		Header: models.Header{
			Action: enums.Auth,
			Sender: enums.Server,
		},
		Body: challenge,
	}
	_, err = challengeMessage.Send(conn)
	if err != nil {
		return err
	}

	var n int
	var challengeResponseMessage models.Message
	n, err = challengeResponseMessage.Receive(conn)
	if err != nil {
		return err
	}
	if n < 1+sha256.Size {
		return fmt.Errorf("expected at least %d bytes in challengeResponseMessage, got %d", 1+sha256.Size, n)
	}
	challengeResponse := challengeResponseMessage.Body.([]byte)[0:sha256.Size]
	userName := string(challengeResponseMessage.Body.([]byte)[sha256.Size:])
	log.Debugf("Received challenge response, response: %s, username: %s", string(challengeResponse), userName)

	// Get the shared key for the user.
	sharedKey, found := a.userService.GetSharedKey(userName)
	// If the user is not found, create a new user and receive the shared key.
	if !found {
		log.Debugf("User %s not found, creating new user", userName)
		newUserMessage := models.Message{
			Header: models.Header{
				Action: enums.Auth,
				Sender: enums.Server,
			},
			Body: enums.NewUser,
		}
		_, err = newUserMessage.Send(conn)
		if err != nil {
			return err
		}
		var sharedKeyMessage models.Message
		_, err = sharedKeyMessage.Receive(conn)
		if err != nil {
			return err
		}
		sharedKey = sharedKeyMessage.Body.([]byte)
		err = a.userService.Create(userName, sharedKey)
		if err != nil {
			return err
		}
		log.Debugf("Created new user %s", userName)
	}

	// Compare the expected response with the received response.
	var expectedResponse []byte
	expectedResponse, err = CalculateResponse(challenge, sharedKey)
	if !bytes.Equal(expectedResponse, challengeResponse) {
		// Send the authenticated message to the client.
		authFailedMessage := models.Message{
			Header: models.Header{
				Action: enums.Auth,
				Sender: enums.Server,
			},
			Body: enums.Unauthorized,
		}
		_, err = authFailedMessage.Send(conn)
		if err != nil {
			return err
		}
		log.Debugf("Authentication failed for user %s", userName)
		return fmt.Errorf("challenge failed")
	}

	// Send the authenticated message to the client.
	authenticatedMessage := models.Message{
		Header: models.Header{
			Action: enums.Auth,
			Sender: enums.Server,
		},
		Body: enums.Authenticated,
	}
	_, err = authenticatedMessage.Send(conn)
	if err != nil {
		return err
	}

	a.userName = userName
	a.authenticated = true
	log.Debugf("Authenticated user %s", a.userName)

	return nil
}

// GetFileService returns the file service of the authenticated user.
func (a *concreteService) GetFileService() (file.Service, error) {
	if !a.authenticated {
		return nil, fmt.Errorf("not authenticated")
	}
	return a.userService.GetFileService(a.userName)
}

func (a *concreteService) IsAuthenticated() bool {
	return a.authenticated
}

func (a *concreteService) GetUsername() string {
	return a.userName
}

// generateChallenge generates a random challenge of the specified length.
func generateChallenge(length int) (challenge []byte, err error) {
	challenge = make([]byte, base64.StdEncoding.EncodedLen(length))
	randomBytes := make([]byte, length)
	_, err = rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	base64.StdEncoding.Encode(challenge, randomBytes)
	return challenge, nil
}

// CalculateResponse calculates the expected response to the challenge using the shared key.
func CalculateResponse(challenge []byte, sharedKey []byte) (response []byte, err error) {
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
