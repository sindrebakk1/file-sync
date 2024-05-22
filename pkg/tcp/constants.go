package tcp

type Version uint8

const (
	V1 Version = 1
)

// TODO: update version to 2 bytes and make HeaderSize 40
const HeaderSize = 39

type Flags uint16

const (
	FlagError   Flags = 1 << iota
	FlagHuffman Flags = 1 << 1
)
