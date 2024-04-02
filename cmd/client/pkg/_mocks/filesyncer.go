package _mocks

import (
	"client/pkg/filesyncer"
	"file-sync/pkg/models"
)

type mockFileSyncer struct {
	syncedFileMap map[string]*models.FileInfo
}

func NewMockFileSyncer(syncedFileMap map[string]*models.FileInfo) filesyncer.FileSyncer {
	return &mockFileSyncer{
		syncedFileMap,
	}
}

func (m *mockFileSyncer) SyncFile(_ string, _ *models.FileInfo) error {
	return nil
}

func (m *mockFileSyncer) GetSyncedFileMap() map[string]*models.FileInfo {
	return m.syncedFileMap
}

func (m *mockFileSyncer) Close() error {
	return nil
}
