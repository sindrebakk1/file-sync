package hasher

import (
	"crypto/sha256"
	"hash"
	"io"
	"os"
)

type Sha256Hasher interface {
	// Hash hashes the given data.
	Hash(data []byte) (hash []byte, err error)
	// HashPath hashes the file at the given path.
	HashPath(path string) (hash []byte, err error)
}

type concreteSha256Hasher struct {
	hasher hash.Hash
}

func NewSha256Hasher() Sha256Hasher {
	return &concreteSha256Hasher{
		sha256.New(),
	}
}

func (h *concreteSha256Hasher) Hash(data []byte) (hash []byte, err error) {
	defer h.hasher.Reset()
	_, err = h.hasher.Write(data)
	if err != nil {
		return nil, err
	}
	return h.hasher.Sum(nil), nil
}

func (h *concreteSha256Hasher) HashPath(path string) (hash []byte, err error) {
	defer h.hasher.Reset()
	var file *os.File
	file, err = os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	_, err = io.Copy(h.hasher, file)
	if err != nil {
		return nil, err
	}
	return h.hasher.Sum(nil), nil
}
