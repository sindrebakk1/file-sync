package filestatus

type FileStatus int

const (
	Stale FileStatus = iota
	Updated
	Synced
	Syncing
	Error
)

func (c FileStatus) String() string {
	return [...]string{"Unknown", "Error", "Syncing", "Synced"}[c]
}
