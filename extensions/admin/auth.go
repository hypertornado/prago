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

func (admin *Admin) createRoleFieldType() FieldType {
	var fp = func() interface{} {
		roleNames := []string{""}
		if admin.roles != nil {
			for k, _ := range admin.roles {
				roleNames = append(roleNames, k)
			}
		}

		vals := [][2]string{}
		for _, v := range roleNames {
			vals = append(vals, [2]string{v, v})
		}
		return vals
	}
	return FieldType{
		FormSubTemplate: "admin_item_select",
		ValuesSource:    &fp,
	}
}

func (a *Admin) AddAuthRole(roleName string, permissions []string) {
	perms := map[string]bool{}
	for _, v := range permissions {
		perms[v] = true
	}
	a.roles[roleName] = perms
}

func (a *Admin) AuthenticatePermission(permission string) Authenticatizer {
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
		return a.roles[u.Role][permission]
	}
}
