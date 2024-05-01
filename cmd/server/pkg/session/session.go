package session

import (
	"context"
	"file-sync/models"
	"server/services"
	"sync"
)

type Session struct {
	Username     string
	Transactions *sync.Map
	FileService  services.FileService
}

func NewContext(ctx context.Context, session *Session) (context.Context, context.CancelFunc) {
	cancelCtx, cancel := context.WithCancel(ctx)
	return context.WithValue(cancelCtx, "session", session), cancel
}

func FromContext(ctx context.Context) (*Session, bool) {
	d, ok := ctx.Value("session").(*Session)
	return d, ok
}

func (d *Session) GetTransaction(transactionID [32]byte) (chan models.Message, bool) {
	if v, ok := d.Transactions.Load(transactionID); ok {
		return v.(chan models.Message), true
	}
	return nil, false
}

func (d *Session) NewTransaction(transactionID [32]byte) chan models.Message {
	ch := make(chan models.Message)
	d.Transactions.Store(transactionID, ch)
	return ch
}
