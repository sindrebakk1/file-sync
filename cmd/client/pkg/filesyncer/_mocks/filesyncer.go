package _mocks

import (
	"client/pkg/filesyncer"
	"file-sync/pkg/globalmodels"
)

type mockFileSyncer struct {
	syncedFileMap map[string]*globalmodels.File
}

func NewMockFileSyncer(syncedFileMap map[string]*globalmodels.File) filesyncer.FileSyncer {
	return &mockFileSyncer{
		syncedFileMap,
	}
}

func (m *mockFileSyncer) SyncFile(_ string, _ *globalmodels.File) error {
	return nil
}

func (m *mockFileSyncer) GetSyncedFileMap() map[string]*globalmodels.File {
	return m.syncedFileMap
}

func (m *mockFileSyncer) Close() error {
	return nil
}
