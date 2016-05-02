package admin

type Authenticatizer func(*User) bool

func AuthenticateAdmin(u *User) bool {
	if u.IsSysadmin {
		return true
	}
	if u.IsAdmin {
		return true
	}
	return false
}

func AuthenticateSysadmin(u *User) bool {
	if u.IsSysadmin {
		return true
	}
	return false
}
