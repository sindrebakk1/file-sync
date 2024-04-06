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
	"server/services"
)

type Authenticator interface {
	AuthenticateClient() (userName string, err error)
}

type Config struct {
	ChallengeLen int
}

type concreteAuthenticator struct {
	encoder     *gob.Encoder
	decoder     *gob.Decoder
	userService services.UserService
	config      *Config
}

func NewAuthenticator(encoder *gob.Encoder, decoder *gob.Decoder, userService services.UserService, config *Config) Authenticator {
	return &concreteAuthenticator{
		encoder,
		decoder,
		userService,
		config,
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
		Message: globalmodels.Message{
			MessageType: globalenums.Authentication,
		},
		Payload: challenge,
	}

	// Send the challenge to the client.
	err = a.encoder.Encode(&challengeMessage)

	// Receive the response from the client.
	var challengeResponseMessage globalmodels.ChallengeResponseMessage
	err = a.decoder.Decode(&challengeResponseMessage)
	if err != nil {
		return "", err
	}
	userName = challengeResponseMessage.Payload.User
	challengeResponse := challengeResponseMessage.Payload.Response

	// Get the shared key for the user.
	sharedKey, found := a.userService.GetSharedKey(userName)
	// If the user is not found, create a new user and receive the shared key.
	if !found {
		newUserMessage := globalmodels.AuthResponseMessage{
			Message: globalmodels.Message{
				MessageType: globalenums.Authentication,
			},
			Payload: globalenums.NewUser,
		}
		err = a.encoder.Encode(&newUserMessage)
		if err != nil {
			return "", err
		}
		// Receive the shared key from the client.
		err = a.decoder.Decode(&sharedKey)
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
			Message: globalmodels.Message{
				MessageType: globalenums.Authentication,
			},
			Payload: globalenums.AuthFailed,
		}
		err = a.encoder.Encode(&authFailedMessage)
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf("challenge failed")
	}

	// Send the authenticated message to the client.
	authenticatedMessage := globalmodels.AuthResponseMessage{
		Message: globalmodels.Message{
			MessageType: globalenums.Authentication,
		},
		Payload: globalenums.Authenticated,
	}
	err = a.encoder.Encode(&authenticatedMessage)
	if err != nil {
		return "", err
	}

	return userName, nil
}

// generateChallenge generates a random challenge of the specified length.
func generateChallenge(length int) (challenge []byte, err error) {
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
