package globalenums

//type StatusCode int

//const (
//	OK             StatusCode = 200
//	Created                   = 201
//	InvalidMessage            = 400
//	Unauthorized              = 401
//	InternalError             = 500
//)
//
//func (c StatusCode) String() string {
//	return [...]string{"OK", "InvalidMessage", "InternalError"}[c]
//}

type AuthResult int

const (
	Authenticated AuthResult = iota
	NewUser
	AuthFailed
)

func (c AuthResult) String() string {
	return [...]string{"Authenticated", "NewUser", "Unauthorized"}[c]
}
