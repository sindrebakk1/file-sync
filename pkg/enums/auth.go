package enums

type AuthResult uint8

const (
	Authenticated AuthResult = iota
	NewUser
	Unauthorized
)

func (c AuthResult) String() string {
	return [...]string{"Authenticated", "NewUser", "Unauthorized"}[c]
}
