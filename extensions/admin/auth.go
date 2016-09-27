package admin

//Authenticatizer is function for user authenticatication
type Authenticatizer func(*User) bool

//AuthenticateAdmin authenticaticatizer for admin
func AuthenticateAdmin(u *User) bool {
	if u.IsSysadmin {
		return true
	}
	if u.IsAdmin {
		return true
	}
	return false
}

//AuthenticateSysadmin authenticaticatizer for sysadmin
func AuthenticateSysadmin(u *User) bool {
	if u.IsSysadmin {
		return true
	}
	return false
}
