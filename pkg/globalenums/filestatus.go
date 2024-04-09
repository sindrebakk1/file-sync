package globalenums

type FileStatus int

const (
	Unknown FileStatus = iota
	Synced
	Syncing
	New
	Dirty
	Stale
)

func (c FileStatus) String() string {
	return [...]string{"Unknown", "Synced", "Syncing", "New", "Dirty", "Stale"}[c]
}
