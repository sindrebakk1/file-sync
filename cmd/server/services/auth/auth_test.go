package auth_test

import (
	"filesync/enums"
	"filesync/models"
	"github.com/stretchr/testify/assert"
	"net"
	"os"
	"server/pkg/cache"
	"server/services/auth"
	"server/services/file"
	"server/services/user"
	"testing"
)

const (
	BaseDir      = "test"
	ChallengeLen = 32
)

var (
	userService        user.Service
	authConfig         *auth.Config
	fileServiceFactory file.Factory
	client             net.Conn
	server             net.Conn

	// Test users
	testUser1   = "test1"
	testSecret1 = []byte("secret1")
	testUser2   = "test2"
	testSecret2 = []byte("secret2")
)

func TestMain(m *testing.M) {
	fileCache := cache.NewCache(100)
	metaCache := cache.NewCache(100)
	fileServiceFactory = file.NewFactory(BaseDir, fileCache, metaCache)
	userService = user.New(fileServiceFactory)
	authConfig = &auth.Config{
		ChallengeLen: ChallengeLen,
	}

	client, server = net.Pipe()
	defer func() {
		client.Close()
		server.Close()
	}()

	os.Exit(m.Run())
}

// TestAuthenticateClientNewUser tests the authentication of a new user.
func TestAuthenticateClientNewUser(t *testing.T) {
	go testClient(client, t, testUser1, testSecret1)

	authenticator := auth.New(userService, authConfig)

	err := authenticator.AuthenticateClient(server)
	assert.NoError(t, err, "Error authenticating client")
	assert.True(t, authenticator.IsAuthenticated(), "Expected authenticated to be true")
	assert.Equal(t, testUser1, authenticator.GetUsername(), "Expected user name to be %s, got %s", testUser1, authenticator.GetUsername())
}

// TestAuthenticateClientExistingUser tests the authentication of an existing user.
func TestAuthenticateClientExistingUser(t *testing.T) {
	go testClient(client, t, testUser1, testSecret1)

	authenticator := auth.New(userService, authConfig)

	err := authenticator.AuthenticateClient(server)
	assert.NoError(t, err, "Error authenticating client")
	assert.True(t, authenticator.IsAuthenticated(), "Expected authenticated to be true")
	assert.Equal(t, testUser1, authenticator.GetUsername(), "Expected user name to be %s, got %s", testUser1, authenticator.GetUsername())
}

// TestAuthenticateClientFailed tests the authentication of a client with an incorrect secret.
func TestAuthenticateClientFailed(t *testing.T) {
	go testClient(client, t, testUser1, testSecret2)

	authenticator := auth.New(userService, authConfig)

	err := authenticator.AuthenticateClient(server)
	assert.Error(t, err, "Expected authentication error")
	assert.False(t, authenticator.IsAuthenticated(), "Expected authenticated to be false")
	assert.Equal(t, "", authenticator.GetUsername(), "Expected user name to be empty, got %s", authenticator.GetUsername())
}

// TestAuthenticateClientNewUser2 tests the authentication of a new user.
func TestAuthenticateClientNewUser2(t *testing.T) {
	go testClient(client, t, testUser2, testSecret2)

	authenticator := auth.New(userService, authConfig)

	err := authenticator.AuthenticateClient(server)
	assert.NoError(t, err, "Error authenticating client")
	assert.Equal(t, testUser2, authenticator.GetUsername(), "Expected user name to be %s, got %s", testUser2, authenticator.GetUsername())

	secret, found := userService.GetSharedKey(testUser2)
	assert.True(t, found, "Expected shared key to be found")
	assert.Equal(t, testSecret2, secret, "Expected shared key to be %v, got %v", testSecret2, secret)
}

func testClient(conn net.Conn, t *testing.T, testUser string, testSecret []byte) {
	var challengeMessage models.Message
	var err error
	_, err = challengeMessage.Receive(conn)
	assert.NoError(t, err, "Error receiving challenge message")

	// Calculate the challenge response.
	var challengeResponse []byte
	challengeResponse, err = auth.CalculateResponse(challengeMessage.Body.([]byte), testSecret)
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
