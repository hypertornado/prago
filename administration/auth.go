package administration

import (
	"fmt"
)

var permissionEverybody Permission = "__everybody"
var permissionSysadmin Permission = "sysadmin"
var permissionNobody Permission = ""

type Permission string

func authNobody(p Permission) bool {
	if p == "" {
		return true
	}
	return false
}

//Authenticatizer is function for user authenticatication
type Authenticatizer func(*User) bool

func (admin Administration) getAllPermissions() []string {
	m := map[string]bool{}
	for _, v1 := range admin.roles {
		for v2, _ := range v1 {
			m[v2] = true
		}
	}
	m[string(permissionSysadmin)] = true
	var ret []string
	for k, _ := range m {
		ret = append(ret, k)
	}
	return ret
}

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

func (admin *Administration) AddRole(role string, permissions []string) {
	perms := map[string]bool{}
	for _, v := range permissions {
		perms[v] = true
	}
	_, ok := admin.roles[role]
	if ok {
		panic(fmt.Sprintf("role '%s' already added", role))
	}
	admin.roles[role] = perms
}

func (admin *Administration) Authorize(user User, permission Permission) bool {
	if !user.IsAdmin {
		return false
	}
	if authNobody(permission) {
		return false
	}
	if permission == permissionEverybody {
		return true
	}
	if user.IsSysadmin {
		return true
	}

	if admin.roles == nil {
		return false
	}
	if admin.roles[user.Role] == nil {
		return false
	}
	return admin.roles[user.Role][string(permission)]
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
