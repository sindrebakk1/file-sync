package enums

type FileStatus int

const (
	Synced FileStatus = iota
	New
	Dirty
	Syncing
	Error
)

func (c FileStatus) String() string {
	return [...]string{"Synced", "None", "New", "Dirty", "Syncing", "Error"}[c]
}
