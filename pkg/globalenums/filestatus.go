package globalenums

type FileStatus int

const (
	Unknown FileStatus = iota
	Error
	Dirty
	Syncing
	Synced
)

func (c FileStatus) String() string {
	return [...]string{"Unknown", "Synced", "Dirty", "Syncing", "Error"}[c]
}
