package session

import (
	"context"
	"server/models"
	"server/services"
	"sync"
)

type Data struct {
	Username     string
	Transactions *sync.Map
	FileService  services.FileService
}

func NewContext(ctx context.Context, session *Data) context.Context {
	return context.WithValue(ctx, "session", session)
}

func FromContext(ctx context.Context) (*Data, bool) {
	d, ok := ctx.Value("session").(*Data)
	return d, ok
}

func (d *Data) GetTransaction(transactionID [32]byte) (chan models.Message, bool) {
	if v, ok := d.Transactions.Load(transactionID); ok {
		return v.(chan models.Message), true
	}
	return nil, false
}

func (d *Data) NewTransaction(transactionID [32]byte) chan models.Message {
	ch := make(chan models.Message)
	d.Transactions.Store(transactionID, ch)
	return ch
}
