package auth

import (
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
	BaseDir      = "test"
	ChallengeLen = 32
	Port         = 8081
)

var (
	userService        services.UserService
	authConfig         *Config
	fileServiceFactory services.FileServiceFactory
	listener           net.Listener

	// Test users
	testUser1   = "test1"
	testSecret1 = []byte("secret1")
	testUser2   = "test2"
	testSecret2 = []byte("secret2")
)

func TestMain(m *testing.M) {
	fileServiceFactory = services.NewFileServiceFactory(BaseDir)
	userService = services.NewUserService(fileServiceFactory)
	authConfig = &Config{
		ChallengeLen: ChallengeLen,
	}

	var err error
	listener, err = net.Listen("tcp", fmt.Sprintf(":%d", Port))
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	os.Exit(m.Run())
}

// TestAuthenticateClientNewUser tests the authentication of a new user.
func TestAuthenticateClientNewUser(t *testing.T) {
	go testClient(t, testUser1, testSecret1)

	conn, err := listener.Accept()
	assert.NoError(t, err)
	defer conn.Close()

	authenticator := NewAuthenticator(conn, userService, authConfig)

	var userName string
	userName, err = authenticator.AuthenticateClient()
	assert.NoError(t, err, "Error authenticating client")
	assert.Equal(t, testUser1, userName, "Expected user name to be %s, got %s", testUser1, userName)

	t.Logf("Server: Authenticated user: %s\n", userName)
}

// TestAuthenticateClientExistingUser tests the authentication of an existing user.
func TestAuthenticateClientExistingUser(t *testing.T) {
	go testClient(t, testUser1, testSecret1)

	conn, err := listener.Accept()
	assert.NoError(t, err)
	defer conn.Close()

	authenticator := NewAuthenticator(conn, userService, authConfig)

	var userName string
	userName, err = authenticator.AuthenticateClient()
	assert.NoError(t, err, "Error authenticating client")
	assert.Equal(t, testUser1, userName, "Expected user name to be %s, got %s", testUser1, userName)

	t.Logf("Server: Authenticated user: %s\n", userName)
}

// TestAuthenticateClientFailed tests the authentication of a client with an incorrect secret.
func TestAuthenticateClientFailed(t *testing.T) {
	go testClient(t, testUser1, testSecret2)

	conn, err := listener.Accept()
	assert.NoError(t, err)
	defer conn.Close()

	authenticator := NewAuthenticator(conn, userService, authConfig)

	var userName string
	userName, err = authenticator.AuthenticateClient()
	assert.Error(t, err, "Expected authentication error")
	assert.Equal(t, "", userName, "Expected user name to be empty, got %s", userName)

	t.Logf("Server: Authentication failed\n")
}

// TestAuthenticateClientNewUser2 tests the authentication of a new user.
func TestAuthenticateClientNewUser2(t *testing.T) {
	go testClient(t, testUser2, testSecret2)

	conn, err := listener.Accept()
	assert.NoError(t, err)
	defer conn.Close()

	authenticator := NewAuthenticator(conn, userService, authConfig)

	var userName string
	userName, err = authenticator.AuthenticateClient()
	assert.NoError(t, err, "Error authenticating client")
	assert.Equal(t, testUser2, userName, "Expected user name to be %s, got %s", testUser2, userName)

	secret, found := userService.GetSharedKey(testUser2)
	assert.True(t, found, "Expected shared key to be found")
	assert.Equal(t, testSecret2, secret, "Expected shared key to be %v, got %v", testSecret2, secret)

	t.Logf("Server: Authenticated user: %s\n", userName)
}

func sendMessage(conn net.Conn, message interface{}) error {
	encoder := gob.NewEncoder(conn)
	return encoder.Encode(message)
}

func receiveMessage(conn net.Conn, message interface{}) error {
	decoder := gob.NewDecoder(conn)
	return decoder.Decode(message)
}

func testClient(t *testing.T, testUser string, testSecret []byte) {
	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", Port))
	assert.NoError(t, err)
	defer conn.Close()

	var challengeMessage globalmodels.ChallengeMessage
	err = receiveMessage(conn, &challengeMessage)
	assert.NoError(t, err, "Error receiving challenge message")
	t.Logf("Client: Received challenge message: %v\n", challengeMessage)

	// Calculate the challenge response.
	var challengeResponse []byte
	challengeResponse, err = calculateResponse(challengeMessage.Payload, testSecret)
	assert.NoError(t, err, "Error calculating response")

	// Send the challenge response to the server.
	challengeResponsePayload := globalmodels.ChallengeResponse{
		User:     testUser,
		Response: challengeResponse,
	}
	challengeResponseMessage := globalmodels.ChallengeResponseMessage{
		Payload: challengeResponsePayload,
	}
	err = sendMessage(conn, &challengeResponseMessage)
	assert.NoError(t, err, "Error sending challenge response message")
	t.Logf("Client: Sent challenge response message: %v\n", challengeResponseMessage)

	// Receive the new user message from the server.
	var authResponseMessage globalmodels.AuthResponseMessage
	err = receiveMessage(conn, &authResponseMessage)
	assert.NoError(t, err)
	switch authResponseMessage.Payload {
	case globalenums.Authenticated:
		t.Logf("Client: Received authenticated message: %v\n", authResponseMessage)
	case globalenums.NewUser:
		t.Logf("Client: Received new user message: %v\n", authResponseMessage)
		// Send the shared key to the server.
		err = sendMessage(conn, &testSecret)
		assert.NoError(t, err)
		t.Logf("Client: Sent shared key: %v\n", testSecret)

		var authResponseMessageNewUser globalmodels.AuthResponseMessage
		err = receiveMessage(conn, &authResponseMessageNewUser)
		assert.NoError(t, err, "Error receiving new user message")
		assert.Equal(t, globalenums.Authenticated, authResponseMessageNewUser.Payload, "Expected authenticated message, got %v", authResponseMessageNewUser.Payload)
		t.Logf("Client: Received authenticated message: %v\n", authResponseMessageNewUser)
	case globalenums.AuthFailed:
		t.Logf("Client: Received authentication failed message: %v\n", authResponseMessage)
	default:
		t.Errorf("Client: Received unexpected message: %v\n", authResponseMessage)
	}
}
