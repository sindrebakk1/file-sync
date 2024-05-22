package tcp

type TypeID uint16

type TransactionID [32]byte

type Length uint16

type Header struct {
	Version       Version
	Flags         Flags
	Type          TypeID
	TransactionID TransactionID
	Length        Length
}

type Message struct {
	Header Header
	Body   interface{}
}
