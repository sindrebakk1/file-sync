package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"file-sync/pkg/globalenums"
	"file-sync/pkg/globalmodels"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net"
	"os"
	"server/services"
	"testing"
)

const (
	Port         = 8081
	BaseDir      = "test"
	ChallengeLen = 4
)

var (
	authenticator      Authenticator
	listener           net.Listener
	userService        services.UserService
	authConfig         *Config
	fileServiceFactory services.FileServiceFactory
	conn               net.Conn
	userName           string
)

func TestMain(m *testing.M) {
	// Initialize services.
	fileServiceFactory = services.NewFileServiceFactory(BaseDir)
	userService = services.NewUserService(fileServiceFactory)
	authConfig = &Config{
		ChallengeLen: ChallengeLen,
	}
	os.Exit(m.Run())
}

func TestSetup(t *testing.T) {
	var (
		err error
	)

	listener, err = net.Listen("tcp", fmt.Sprintf("localhost:%d", Port))
	if err != nil {
		t.Fatalf("Failed to start listener: %v", err)
	}
	go func() {
		for {
			conn, err = listener.Accept()
			if err != nil {
				t.Errorf("Failed to accept connection: %v", err)
				return
			}

			encoder := gob.NewEncoder(conn)
			decoder := gob.NewDecoder(conn)
			authenticator = NewAuthenticator(encoder, decoder, userService, authConfig)
			userName, err = authenticator.AuthenticateClient()
			if err != nil {
				t.Errorf("Failed to authenticate client: %v", err)
				return
			}
			t.Logf("Authenticated client: %s", userName)

			conn.Close()
		}
	}()
}

func TestTeardown(t *testing.T) {
	listener.Close()
	authenticator = nil
}

// TestAuthenticateClientNewUser tests the authentication of a new user.
func TestAuthenticateClientNewUser(t *testing.T) {
	var (
		err           error
		testUser      = "test"
		testSharedKey = []byte("testsecret")
	)
	conn, err = net.Dial("tcp", fmt.Sprintf("localhost:%d", Port))
	if err != nil {
		t.Fatalf("Failed to dial connection: %v", err)
	}
	encoder := gob.NewEncoder(conn)
	decoder := gob.NewDecoder(conn)

	// Receive the challenge from the server.
	var challengeMessage globalmodels.ChallengeMessage
	err = decoder.Decode(&challengeMessage)
	if err != nil {
		t.Fatalf("Failed to receive challenge: %v", err)
	}

	// Calculate the response to the challenge.
	response := make([]byte, base64.StdEncoding.EncodedLen(sha256.Size))
	response, err = calculateChallengeResponse(challengeMessage.Payload, testSharedKey)
	if err != nil {
		t.Fatalf("Failed to calculate challenge response: %v", err)
	}

	// Respond to the challenge.
	challengeResponseMessage := globalmodels.ChallengeResponseMessage{
		Message: globalmodels.Message{
			MessageType: globalenums.Authentication,
		},
		Payload: globalmodels.ChallengeResponse{
			User:     testUser,
			Response: response,
		},
	}
	err = encoder.Encode(&challengeResponseMessage)
	if err != nil {
		t.Fatalf("Failed to send challenge response: %v", err)
	}

	// Receive the authentication result from the server.
	var authResponseMessage globalmodels.AuthResponseMessage
	err = decoder.Decode(&authResponseMessage)
	if err != nil {
		t.Fatalf("Failed to receive authentication result: %v", err)
	}

	assert.Equal(t, globalenums.NewUser, authResponseMessage.Payload, "Expected new user to be created")

	err = encoder.Encode(&testSharedKey)

	err = decoder.Decode(&authResponseMessage)

	assert.Equal(t, globalenums.Authenticated, authResponseMessage.Payload, "Expected OK status")
	assert.Equal(t, testUser, userName, "Expected user name to be set correctly")

	conn.Close()
}

func calculateChallengeResponse(challenge []byte, sharedKey []byte) (response []byte, err error) {
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
