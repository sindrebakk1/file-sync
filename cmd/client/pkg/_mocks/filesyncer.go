package _mocks

import (
	"client/pkg/filesyncer"
	"file-sync/pkg/globalmodels"
)

type mockFileSyncer struct {
	syncedFileMap map[string]*globalmodels.FileInfo
}

func NewMockFileSyncer(syncedFileMap map[string]*globalmodels.FileInfo) filesyncer.FileSyncer {
	return &mockFileSyncer{
		syncedFileMap,
	}
}

func (m *mockFileSyncer) SyncFile(_ string, _ *globalmodels.FileInfo) error {
	return nil
}

func (m *mockFileSyncer) GetSyncedFileMap() map[string]*globalmodels.FileInfo {
	return m.syncedFileMap
}

func (m *mockFileSyncer) Close() error {
	return nil
}
