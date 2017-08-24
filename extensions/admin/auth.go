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

func (a *Admin) AddAuthRole(roleName string, permissions []string) {
	if a.roles == nil {
		a.roles = make(map[string]map[string]bool)
	}
	perms := map[string]bool{}
	for _, v := range permissions {
		perms[v] = true
	}
	a.roles[roleName] = perms
}

func AuthenticatePermission(a *Admin, role string) Authenticatizer {
	return func(u *User) bool {
		if u.IsSysadmin {
			return true
		}
		if !u.IsAdmin {
			return false
		}
		if a.roles == nil {
			return false
		}
		if a.roles[u.Role] == nil {
			return false
		}
		return a.roles[u.Role][role]
	}
}
