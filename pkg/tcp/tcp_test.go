package tcp_test

import (
	"github.com/stretchr/testify/assert"
	"net"
	"tcp"
	"testing"
)

func TestEncodeDecodeMessage_String(t *testing.T) {
	tcs := []testCase{
		{
			value: tcp.Message{
				Header: tcp.Header{
					Version:       tcp.V1,
					Flags:         tcp.FError | tcp.FHuff,
					TransactionID: tcp.TransactionID(make([]byte, tcp.TransactionIDSize)),
				},
				Body: "Hello",
			},
			name: "Hello",
		},
		{
			value: tcp.Message{
				Header: tcp.Header{
					Version:       tcp.V1,
					Flags:         tcp.FError | tcp.FHuff,
					TransactionID: tcp.TransactionID(make([]byte, tcp.TransactionIDSize)),
				},
				Body: "World",
			},
			name: "World",
		},
		{
			value: tcp.Message{
				Header: tcp.Header{
					Version:       tcp.V1,
					Flags:         tcp.FError | tcp.FHuff,
					TransactionID: tcp.TransactionID(make([]byte, tcp.TransactionIDSize)),
				},
				Body: "GoLang!%&",
			},
			name: "GoLang",
		},
		{
			value: tcp.Message{
				Header: tcp.Header{
					Version:       tcp.V1,
					Flags:         tcp.FError | tcp.FHuff,
					TransactionID: tcp.TransactionID(make([]byte, tcp.TransactionIDSize)),
				},
				Body: "12345",
			},
			name: "number string",
		},
		{
			value: tcp.Message{
				Header: tcp.Header{
					Version:       tcp.V1,
					Flags:         tcp.FError | tcp.FHuff,
					TransactionID: tcp.TransactionID(make([]byte, tcp.TransactionIDSize)),
				},
				Body: "",
			},
			name: "empty string",
		},
	}
	testEncodeDecodeMessages(t, tcs)
}

func TestEncodeDecodeMessage_Numbers(t *testing.T) {
	tcs := []testCase{
		{
			value: tcp.Message{
				Header: tcp.Header{
					Version:       tcp.V1,
					Flags:         tcp.FError | tcp.FHuff,
					TransactionID: tcp.TransactionID(make([]byte, tcp.TransactionIDSize)),
				},
				Body: 1234,
			},
			name: "int",
		},
		{
			value: tcp.Message{
				Header: tcp.Header{
					Version:       tcp.V1,
					Flags:         tcp.FError | tcp.FHuff,
					TransactionID: tcp.TransactionID(make([]byte, tcp.TransactionIDSize)),
				},
				Body: uint8(123),
			},
			name: "uint8",
		},
		{
			value: tcp.Message{
				Header: tcp.Header{
					Version:       tcp.V1,
					Flags:         tcp.FError | tcp.FHuff,
					TransactionID: tcp.TransactionID(make([]byte, tcp.TransactionIDSize)),
				},
				Body: uint16(1234),
			},
			name: "uint16",
		},
		{
			value: tcp.Message{
				Header: tcp.Header{
					Version:       tcp.V1,
					Flags:         tcp.FError | tcp.FHuff,
					TransactionID: tcp.TransactionID(make([]byte, tcp.TransactionIDSize)),
				},
				Body: uint32(1234),
			},
			name: "uint32",
		},
		{
			value: tcp.Message{
				Header: tcp.Header{
					Version:       tcp.V1,
					Flags:         tcp.FError | tcp.FHuff,
					TransactionID: tcp.TransactionID(make([]byte, tcp.TransactionIDSize)),
				},
				Body: uint64(1234),
			},
			name: "uint64",
		},
		{
			value: tcp.Message{
				Header: tcp.Header{
					Version:       tcp.V1,
					Flags:         tcp.FError | tcp.FHuff,
					TransactionID: tcp.TransactionID(make([]byte, tcp.TransactionIDSize)),
				},
				Body: int8(123),
			},
			name: "int8",
		},
		{
			value: tcp.Message{
				Header: tcp.Header{
					Version:       tcp.V1,
					Flags:         tcp.FError | tcp.FHuff,
					TransactionID: tcp.TransactionID(make([]byte, tcp.TransactionIDSize)),
				},
				Body: int16(1234),
			},
			name: "int16",
		},
		{
			value: tcp.Message{
				Header: tcp.Header{
					Version:       tcp.V1,
					Flags:         tcp.FError | tcp.FHuff,
					TransactionID: tcp.TransactionID(make([]byte, tcp.TransactionIDSize)),
				},
				Body: int32(12345),
			},
			name: "int32",
		},
		{
			value: tcp.Message{
				Header: tcp.Header{
					Version:       tcp.V1,
					Flags:         tcp.FError | tcp.FHuff,
					TransactionID: tcp.TransactionID(make([]byte, tcp.TransactionIDSize)),
				},
				Body: int64(123456),
			},
			name: "int64",
		},
		{
			value: tcp.Message{
				Header: tcp.Header{
					Version:       tcp.V1,
					Flags:         tcp.FError | tcp.FHuff,
					TransactionID: tcp.TransactionID(make([]byte, tcp.TransactionIDSize)),
				},
				Body: float32(1234.567),
			},
			name: "float32",
		},
		{
			value: tcp.Message{
				Header: tcp.Header{
					Version:       tcp.V1,
					Flags:         tcp.FError | tcp.FHuff,
					TransactionID: tcp.TransactionID(make([]byte, tcp.TransactionIDSize)),
				},
				Body: uint16(1234),
			},
			name: "uint16",
		},
		{
			value: tcp.Message{
				Header: tcp.Header{
					Version:       tcp.V1,
					Flags:         tcp.FError | tcp.FHuff,
					TransactionID: tcp.TransactionID(make([]byte, tcp.TransactionIDSize)),
				},
				Body: float64(1234.567),
			},
			name: "float64",
		},
	}
	testEncodeDecodeMessages(t, tcs)
}

func TestEncodeDecodeMessage_Boolean(t *testing.T) {
	tcs := []testCase{
		{
			value: tcp.Message{
				Header: tcp.Header{
					Version:       tcp.V1,
					Flags:         tcp.FError | tcp.FHuff,
					TransactionID: tcp.TransactionID(make([]byte, tcp.TransactionIDSize)),
				},
				Body: false,
			},
			name: "false",
		},
		{
			value: tcp.Message{
				Header: tcp.Header{
					Version:       tcp.V1,
					Flags:         tcp.FError | tcp.FHuff,
					TransactionID: tcp.TransactionID(make([]byte, tcp.TransactionIDSize)),
				},
				Body: true,
			},
			name: "true",
		},
	}
	testEncodeDecodeMessages(t, tcs)
}

func TestEncodeDecodeMessage_Slice(t *testing.T) {
	type testStruct3 struct {
		A int
		B string
		C bool
	}
	tcp.RegisterType(testStruct3{})
	tcs := []testCase{
		{
			value: tcp.Message{
				Header: tcp.Header{
					Version:       tcp.V1,
					Flags:         tcp.FError | tcp.FHuff,
					TransactionID: tcp.TransactionID(make([]byte, tcp.TransactionIDSize)),
				},
				Body: []int{1, 2, 3},
			},
			name: "int slice",
		},
		{
			value: tcp.Message{
				Header: tcp.Header{
					Version:       tcp.V1,
					Flags:         tcp.FError | tcp.FHuff,
					TransactionID: tcp.TransactionID(make([]byte, tcp.TransactionIDSize)),
				},
				Body: []uint{4, 5, 6},
			},
			name: "uint slice",
		},
		{
			value: tcp.Message{
				Header: tcp.Header{
					Version:       tcp.V1,
					Flags:         tcp.FError | tcp.FHuff,
					TransactionID: tcp.TransactionID(make([]byte, tcp.TransactionIDSize)),
				},
				Body: []float32{1.1, 2.2, 3.3},
			},
			name: "float32 slice",
		},
		{
			value: tcp.Message{
				Header: tcp.Header{
					Version:       tcp.V1,
					Flags:         tcp.FError | tcp.FHuff,
					TransactionID: tcp.TransactionID(make([]byte, tcp.TransactionIDSize)),
				},
				Body: []float64{4.4, 5.5, 6.6},
			},
			name: "float64 slice",
		},
		{
			value: tcp.Message{
				Header: tcp.Header{
					Version:       tcp.V1,
					Flags:         tcp.FError | tcp.FHuff,
					TransactionID: tcp.TransactionID(make([]byte, tcp.TransactionIDSize)),
				},
				Body: []string{"a", "b", "c"},
			},
			name: "string slice",
		},
		{
			value: tcp.Message{
				Header: tcp.Header{
					Version:       tcp.V1,
					Flags:         tcp.FError | tcp.FHuff,
					TransactionID: tcp.TransactionID(make([]byte, tcp.TransactionIDSize)),
				},
				Body: []bool{true, false, true},
			},
			name: "bool slice",
		},
		{
			value: tcp.Message{
				Header: tcp.Header{
					Version:       tcp.V1,
					Flags:         tcp.FError | tcp.FHuff,
					TransactionID: tcp.TransactionID(make([]byte, tcp.TransactionIDSize)),
				},
				Body: []testStruct3{{1, "a", true}, {2, "B", false}},
			},
			name: "struct slice",
		},
	}
	testEncodeDecodeMessages(t, tcs)
}

func TestEncodeDecodeMessage_Struct(t *testing.T) {
	type testStruct3 struct {
		A int
		B string
		C bool
	}
	type TestStruct3 struct {
		A int
		B string
		C bool
	}
	type testStructSliceField3 struct {
		Slice []int
	}
	type testStructStructField3 struct {
		Struct testStruct3
		B      string
	}
	type testStructStructFieldSliceField3 struct {
		Structs []testStruct3
	}
	type testStructEmbeddedPrivateStruct3 struct {
		testStruct3
		B string
	}
	type testStructEmbeddedStruct3 struct {
		TestStruct3
		B string
	}
	type testStructPrivateFields3 struct {
		a int
		b string
		c bool
	}
	tcp.RegisterType(testStruct3{})
	tcp.RegisterType(testStructSliceField3{})
	tcp.RegisterType(testStructStructField3{})
	tcp.RegisterType(testStructStructFieldSliceField3{})
	tcp.RegisterType(testStructEmbeddedPrivateStruct3{})
	tcp.RegisterType(testStructEmbeddedStruct3{})
	tcp.RegisterType(testStructPrivateFields3{})

	var (
		testStructEmbeddedPrivateStruct3ID tcp.TypeID
		testStructPrivateFields3ID         tcp.TypeID
		err                                error
	)
	testStructEmbeddedPrivateStruct3ID, err = tcp.GetIDFromType(testStructEmbeddedPrivateStruct3{})
	assert.NoError(t, err)
	testStructPrivateFields3ID, err = tcp.GetIDFromType(testStructPrivateFields3{})
	assert.NoError(t, err)

	tcs := []testCase{
		{
			value: tcp.Message{
				Header: tcp.Header{
					Flags:         tcp.FError | tcp.FHuff,
					TransactionID: tcp.TransactionID(make([]byte, tcp.TransactionIDSize)),
				},
				Body: testStruct3{1, "a", true},
			},
			name: "struct",
		},
		{
			value: tcp.Message{
				Header: tcp.Header{
					Flags:         tcp.FError | tcp.FHuff,
					TransactionID: tcp.TransactionID(make([]byte, tcp.TransactionIDSize)),
				},
				Body: testStructSliceField3{[]int{1, 2, 3}},
			},
			name: "struct with slice",
		},
		{
			value: tcp.Message{
				Header: tcp.Header{
					Flags:         tcp.FError | tcp.FHuff,
					TransactionID: tcp.TransactionID(make([]byte, tcp.TransactionIDSize)),
				},
				Body: testStructStructField3{testStruct3{1, "a", true}, "b"},
			},
			name: "struct with struct",
		},
		{
			value: tcp.Message{
				Header: tcp.Header{
					Flags:         tcp.FError | tcp.FHuff,
					TransactionID: tcp.TransactionID(make([]byte, tcp.TransactionIDSize)),
				},
				Body: testStructStructFieldSliceField3{[]testStruct3{{1, "a", true}, {1, "a", true}}},
			},
			name: "struct with struct slice",
		},
		{
			value: tcp.Message{
				Header: tcp.Header{
					Flags:         tcp.FError | tcp.FHuff,
					TransactionID: tcp.TransactionID(make([]byte, tcp.TransactionIDSize)),
				},
				Body: testStructEmbeddedPrivateStruct3{testStruct3{1, "a", true}, "b"},
			},
			expected: tcp.Message{
				Header: tcp.Header{
					Flags:         tcp.FError | tcp.FHuff,
					Type:          testStructEmbeddedPrivateStruct3ID,
					TransactionID: tcp.TransactionID(make([]byte, tcp.TransactionIDSize)),
					Length:        0x3,
				},
				Body: testStructEmbeddedPrivateStruct3{testStruct3{}, "b"},
			},
			name: "struct with embedded private struct",
		},
		{
			value: tcp.Message{
				Header: tcp.Header{
					Flags:         tcp.FError | tcp.FHuff,
					TransactionID: tcp.TransactionID(make([]byte, tcp.TransactionIDSize)),
				},
				Body: testStructEmbeddedStruct3{TestStruct3{1, "a", true}, "b"},
			},
			name: "struct with embedded struct",
		},
		{
			value: tcp.Message{
				Header: tcp.Header{
					Flags:         tcp.FError | tcp.FHuff,
					TransactionID: tcp.TransactionID(make([]byte, tcp.TransactionIDSize)),
				},
				Body: testStructPrivateFields3{1, "a", true},
			},
			expected: tcp.Message{
				Header: tcp.Header{
					Flags:         tcp.FError | tcp.FHuff,
					Type:          testStructPrivateFields3ID,
					TransactionID: tcp.TransactionID(make([]byte, tcp.TransactionIDSize)),
					Length:        0x0,
				},
				Body: testStructPrivateFields3{},
			},
			name: "struct with private fields",
		},
	}
	testEncodeDecodeMessages(t, tcs)
}

func testEncodeDecodeMessages(t *testing.T, tcs []testCase) {
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			testEncodeMessage(t, tc)
		})
	}
}

func testEncodeMessage(t *testing.T, tc testCase) {
	msg := tc.value.(tcp.Message)
	client, server := net.Pipe()
	// Write to server
	go func() {
		encoder := tcp.NewEncoder(server)
		err := encoder.Encode(&msg)
		assert.NoError(t, err)
		server.Close()
	}()

	// Read from client
	decoder := tcp.NewDecoder(client)
	var res tcp.Message
	err := decoder.Decode(&res)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	expected := msg
	if tc.expected != nil {
		expected = tc.expected.(tcp.Message)
	}
	assert.Equal(t, expected, res)
	client.Close()
}
