package administration

import (
	"fmt"
)

var permissionSysadmin Permission = "sysadmin"
var permissionNobody Permission = ""

type Permission string

func (admin Administration) getSysadminPermissions() []string {
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

func (admin *Administration) createRoleFieldType() FieldType {
	var fp = func(field, User) interface{} {
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
		FormTemplate:   "admin_item_select",
		FormDataSource: fp,
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

func (admin Administration) Authorize(user User, permission Permission) bool {
	if !user.IsAdmin {
		return false
	}
	if permission == "" {
		return true
	}

	//TODO: remove issysadmin after fixed this
	if user.IsSysadmin && user.Role == "" {
		user.Role = "sysadmin"
	}

	return admin.roles[user.Role][string(permission)]
}
