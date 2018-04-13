package administration

//Authenticatizer is function for user authenticatication
type Authenticatizer func(*User) bool

//AuthenticateAdmin authenticaticatizer for admin
func AuthenticateAdmin(user *User) bool {
	if user.IsSysadmin {
		return true
	}
	if user.IsAdmin {
		return true
	}
	return false
}

//AuthenticateSysadmin authenticaticatizer for sysadmin
func AuthenticateSysadmin(user *User) bool {
	if user.IsSysadmin {
		return true
	}
	return false
}

func (admin *Administration) createRoleFieldType() FieldType {
	var fp = func() interface{} {
		roleNames := []string{""}
		for k, _ := range admin.roles {
			roleNames = append(roleNames, k)
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

func (admin *Administration) AddAuthRole(role string, permissions []string) {
	perms := map[string]bool{}
	for _, v := range permissions {
		perms[v] = true
	}
	admin.roles[role] = perms
}

func (admin *Administration) AuthenticatePermission(permission string) Authenticatizer {
	return func(u *User) bool {
		if u.IsSysadmin {
			return true
		}
		if !u.IsAdmin {
			return false
		}
		if admin.roles == nil {
			return false
		}
		if admin.roles[u.Role] == nil {
			return false
		}
		return admin.roles[u.Role][permission]
	}
}
