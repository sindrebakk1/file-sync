package models

import (
	"bytes"
	"encoding/binary"
	globalenums "file-sync/enums"
	"fmt"
	"io"
	"net"
	"server/constants"
	"server/models"
)

const (
	TransactionIDSize = 32
	HeaderSize        = TransactionIDSize + 1 + 1 + 1 + 2
)

type Header struct {
	Version       uint8
	Sender        globalenums.Sender
	Action        globalenums.MessageType
	Length        uint16
	TransactionID [TransactionIDSize]byte
}

type Message struct {
	Header Header
	Body   interface{}
}

// StatusRequest is the type of Message.Body for a Status request.
type StatusRequest models.FileInfoBytes

// StatusResponse is the type of Message.Body for a Status response.
type StatusResponse globalenums.FileStatus

// UploadRequest is the type of Message.Body for an Upload request.
type UploadRequest []byte

// UploadResponse is the type of Message.Body for an Upload response.
type UploadResponse globalenums.Done

// DownloadRequest is the type of Message.Body for a Download request.
type DownloadRequest models.FileHash

// DownloadResponse is the type of Message.Body for a Download response.
type DownloadResponse []byte

// DeleteRequest is the type of Message.Body for a Delete request.
type DeleteRequest models.FileHash

// DeleteResponse is the type of Message.Body for a Delete response.
type DeleteResponse globalenums.Done

// ListResponse is the type of Message.Body for a List response.
type ListResponse []models.FileInfoBytes

func (m *Message) unmarshalBinary(reader io.Reader) (int, error) {
	header := make([]byte, HeaderSize)
	n, err := reader.Read(header)
	if err != nil {
		return 0, err
	}
	if n != len(header) {
		return 0, fmt.Errorf("expected %d bytes, got %d", len(header), n)
	}
	m.Header.Version = header[0]
	if m.Header.Version != constants.Version {
		return 0, fmt.Errorf("unsupported version: %d", m.Header.Version)
	}
	m.Header.Sender = globalenums.Sender(header[1])
	m.Header.Action = globalenums.MessageType(header[2])
	m.Header.Length = binary.BigEndian.Uint16(header[3:5])
	transactionID := header[5 : 5+TransactionIDSize]
	copy(m.Header.TransactionID[:], transactionID)

	body := make([]byte, m.Header.Length)
	n, err = reader.Read(body)
	if err != nil {
		return 0, err
	}
	if n != len(body) {
		return 0, fmt.Errorf("expected %d bytes, got %d", len(body), n)
	}

	switch m.Header.Action {
	case globalenums.Auth:
		m.Body = body
	case globalenums.Status:
		if m.Header.Sender == globalenums.Client {
			m.Body = StatusRequest(body)
			break
		}
		if m.Header.Sender == globalenums.Server {
			m.Body = StatusResponse(body[0])
			break
		}
	case globalenums.Upload:
		if m.Header.Sender == globalenums.Client {
			m.Body = UploadRequest(body)
			break
		}
		if m.Header.Sender == globalenums.Server {
			m.Body = UploadResponse(body[0])
			break
		}
	case globalenums.Download:
		if m.Header.Sender == globalenums.Client {
			m.Body = DownloadRequest(body)
			break
		}
		if m.Header.Sender == globalenums.Server {
			m.Body = DownloadResponse(body)
			break
		}
	case globalenums.Delete:
		if m.Header.Sender == globalenums.Client {
			m.Body = DeleteRequest(body)
			break
		}
		if m.Header.Sender == globalenums.Server {
			m.Body = DeleteResponse(body[0])
			break
		}
	case globalenums.List:
		if m.Header.Sender == globalenums.Client {
			m.Body = nil
			break
		}
		if m.Header.Sender == globalenums.Server {
			// loop over body and create FileInfoBytes
			var fileInfoBytes models.FileInfoBytes
			for i := 0; i < len(body); i += models.FileInfoSize {
				copy(fileInfoBytes[:], body[i:i+models.FileInfoSize])
				m.Body = append(m.Body.([]models.FileInfoBytes), fileInfoBytes)
			}
			break
		}
	case globalenums.Cancel:
		m.Body = nil
	default:
		return 0, fmt.Errorf("unknown action: %s", m.Header.Action)
	}

	return n, nil
}

func (m *Message) marshalBinary() ([]byte, error) {
	body := new(bytes.Buffer)
	err := binary.Write(body, binary.BigEndian, m.Body)
	if err != nil {
		return nil, err
	}

	header := make([]byte, HeaderSize)
	header[0] = constants.Version
	header[1] = uint8(m.Header.Sender)
	header[2] = uint8(m.Header.Action)
	binary.BigEndian.PutUint16(header[3:5], uint16(body.Len()))
	copy(header[5:5+TransactionIDSize], m.Header.TransactionID[:])

	return append(header, body.Bytes()...), nil
}

func (m *Message) Receive(conn net.Conn) (n int, err error) {
	n, err = m.unmarshalBinary(conn)
	return n, err
}

func (m *Message) Send(conn net.Conn) (n int, err error) {
	var messageBytes []byte
	messageBytes, err = m.marshalBinary()
	if err != nil {
		return 0, err
	}
	n, err = conn.Write(messageBytes)
	if err != nil {
		return 0, err
	}
	return n, nil
}
