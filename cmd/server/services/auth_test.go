package services_test

import (
	"filesync/enums"
	"filesync/models"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net"
	"os"
	"server/pkg/cache"
	"server/services"
	"testing"
)

const (
	BaseDir      = "test"
	ChallengeLen = 32
	Port         = 5000
)

var (
	userService        services.UserService
	authConfig         *services.Config
	fileServiceFactory services.FileServiceFactory
	listener           net.Listener

	// Test users
	testUser1   = "test1"
	testSecret1 = []byte("secret1")
	testUser2   = "test2"
	testSecret2 = []byte("secret2")
)

func TestMain(m *testing.M) {
	fileCache := cache.NewCache(100)
	metaCache := cache.NewCache(100)
	fileServiceFactory = services.NewFileServiceFactory(BaseDir, fileCache, metaCache)
	userService = services.NewUserService(fileServiceFactory)
	authConfig = &services.Config{
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

	authenticator := services.NewAuthService(userService, authConfig)

	err = authenticator.AuthenticateClient(conn)
	assert.NoError(t, err, "Error authenticating client")
	assert.True(t, authenticator.IsAuthenticated(), "Expected authenticated to be true")
	assert.Equal(t, testUser1, authenticator.GetUsername(), "Expected user name to be %s, got %s", testUser1, authenticator.GetUsername())
}

// TestAuthenticateClientExistingUser tests the authentication of an existing user.
func TestAuthenticateClientExistingUser(t *testing.T) {
	go testClient(t, testUser1, testSecret1)

	conn, err := listener.Accept()
	assert.NoError(t, err)
	defer conn.Close()

	authenticator := services.NewAuthService(userService, authConfig)

	err = authenticator.AuthenticateClient(conn)
	assert.NoError(t, err, "Error authenticating client")
	assert.True(t, authenticator.IsAuthenticated(), "Expected authenticated to be true")
	assert.Equal(t, testUser1, authenticator.GetUsername(), "Expected user name to be %s, got %s", testUser1, authenticator.GetUsername())
}

// TestAuthenticateClientFailed tests the authentication of a client with an incorrect secret.
func TestAuthenticateClientFailed(t *testing.T) {
	go testClient(t, testUser1, testSecret2)

	conn, err := listener.Accept()
	assert.NoError(t, err)
	defer conn.Close()

	authenticator := services.NewAuthService(userService, authConfig)

	err = authenticator.AuthenticateClient(conn)
	assert.Error(t, err, "Expected authentication error")
	assert.False(t, authenticator.IsAuthenticated(), "Expected authenticated to be false")
	assert.Equal(t, "", authenticator.GetUsername(), "Expected user name to be empty, got %s", authenticator.GetUsername())
}

// TestAuthenticateClientNewUser2 tests the authentication of a new user.
func TestAuthenticateClientNewUser2(t *testing.T) {
	go testClient(t, testUser2, testSecret2)

	conn, err := listener.Accept()
	assert.NoError(t, err)
	defer conn.Close()

	authenticator := services.NewAuthService(userService, authConfig)

	err = authenticator.AuthenticateClient(conn)
	assert.NoError(t, err, "Error authenticating client")
	assert.Equal(t, testUser2, authenticator.GetUsername(), "Expected user name to be %s, got %s", testUser2, authenticator.GetUsername())

	secret, found := userService.GetSharedKey(testUser2)
	assert.True(t, found, "Expected shared key to be found")
	assert.Equal(t, testSecret2, secret, "Expected shared key to be %v, got %v", testSecret2, secret)
}

func testClient(t *testing.T, testUser string, testSecret []byte) {
	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", Port))
	assert.NoError(t, err)
	defer conn.Close()

	var challengeMessage models.Message
	_, err = challengeMessage.Receive(conn)
	assert.NoError(t, err, "Error receiving challenge message")

	// Calculate the challenge response.
	var challengeResponse []byte
	challengeResponse, err = services.CalculateResponse(challengeMessage.Body.([]byte), testSecret)
	assert.NoError(t, err, "Error calculating response")

	// Send the challenge response to the server.
	challengeResponsePayload := append(challengeResponse, []byte(testUser)...)
	challengeResponseMessage := models.Message{
		Header: models.Header{
			Action: enums.Auth,
		},
		Body: challengeResponsePayload,
	}
	_, err = challengeResponseMessage.Send(conn)
	assert.NoError(t, err, "Error sending challenge response message")

	// Receive the new user message from the server.
	var authResponseMessage models.Message
	_, err = authResponseMessage.Receive(conn)
	assert.NoError(t, err)

	result := enums.AuthResult(authResponseMessage.Body.([]byte)[0])
	switch result {
	case enums.Authenticated:
		t.Logf("Client: Received authenticated message: %v\n", authResponseMessage)
	case enums.NewUser:
		t.Logf("Client: Received new user message: %v\n", authResponseMessage)
		// Send the shared key to the server.
		secretMessage := models.Message{
			Header: models.Header{
				Action: enums.Auth,
			},
			Body: testSecret,
		}
		_, err = secretMessage.Send(conn)
		assert.NoError(t, err)

		var responseMessage models.Message
		_, err = responseMessage.Receive(conn)
		assert.NoError(t, err, "Error receiving new user response message")
		t.Logf("Client: Received authenticated message: %v\n", responseMessage)
		result = enums.AuthResult(responseMessage.Body.([]byte)[0])
		assert.Equal(t, enums.Authenticated, result, "Expected authenticated message, got %v", result)
	case enums.Unauthorized:
		t.Logf("Client: Received authentication failed message: %v\n", authResponseMessage)
	default:
		t.Errorf("Client: Received unexpected message: %v\n", authResponseMessage)
	}
}
