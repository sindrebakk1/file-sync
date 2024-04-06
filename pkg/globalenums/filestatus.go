package globalenums

type FileStatus int

const (
	Unknown FileStatus = iota
	Dirty
	Syncing
	Synced
)

func (c FileStatus) String() string {
	return [...]string{"Unknown", "Synced", "Dirty", "Syncing"}[c]
}
