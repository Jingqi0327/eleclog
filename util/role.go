package util

const (
	UserRole    = "user"
	ManagerRole = "manager"
	AdminRole   = "admin"
)

func IsSupportedRole(role string) bool {
	switch role {
	case UserRole, ManagerRole, AdminRole:
		return true
	}
	return false
}