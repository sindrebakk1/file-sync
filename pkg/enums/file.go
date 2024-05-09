package enums

type FileStatus uint8

const (
	Unknown FileStatus = iota
	Stale
	Dirty
	Syncing
	Synced
)

func (f FileStatus) String() string {
	return [...]string{"Unknown", "Stale", "Dirty", "Syncing", "Synced"}[f]
}
