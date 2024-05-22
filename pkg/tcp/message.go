package tcp

const (
	VersionSize       = 1
	FlagsSize         = 1
	TypeIDSize        = 2
	TransactionIDSize = 16
	LengthSize        = 2
)

type Version uint8

const (
	V1 Version = 1
)

type Flag uint8

const (
	FError Flag = 1 << iota
	FHuff  Flag = 1 << 1
)

const HeaderSize = VersionSize + FlagsSize + TypeIDSize + TransactionIDSize + LengthSize

type TypeID uint16

type TransactionID [TransactionIDSize]byte

type Length uint16

type Header struct {
	Version       Version
	Flags         Flag
	Type          TypeID
	TransactionID TransactionID
	Length        Length
}

type Message struct {
	Header Header
	Body   interface{}
}
