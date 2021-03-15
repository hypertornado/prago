package prago

import (
	"fmt"
)

var permissionSysadmin Permission = "sysadmin"

//Permission for access
type Permission string

func (app *App) initSysadminPermissions() {
	app.AddRole("sysadmin", app.getSysadminPermissions())
}

func (app App) getSysadminPermissions() []string {
	m := map[string]bool{}
	for _, v1 := range app.roles {
		for v2 := range v1 {
			m[v2] = true
		}
	}
	m[string(permissionSysadmin)] = true
	var ret []string
	for k := range m {
		ret = append(ret, k)
	}
	return ret
}

func (app *App) getRoleFieldTypeData() [][2]string {
	roleNames := []string{""}
	for k := range app.roles {
		roleNames = append(roleNames, k)
	}

	vals := [][2]string{}
	for _, v := range roleNames {
		vals = append(vals, [2]string{v, v})
	}
	return vals
}

func (app *App) createRoleFieldType() FieldType {
	var fp = func(Field, User) interface{} {
		roleNames := []string{""}
		for k := range app.roles {
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

		FilterLayoutTemplate: "filter_layout_select",
	}
}

//AddRole adds role to admin
func (app *App) AddRole(role string, permissions []string) {
	perms := map[string]bool{}
	for _, v := range permissions {
		perms[v] = true
	}
	_, ok := app.roles[role]
	if ok {
		panic(fmt.Sprintf("role '%s' already added", role))
	}
	app.roles[role] = perms
}

//Authorize user for task
func (app App) Authorize(user User, permission Permission) bool {
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

	return app.roles[user.Role][string(permission)]
}

func (app App) getResourceViewRoles(resource Resource) []string {
	var ret []string
	for roleName, permissions := range app.roles {
		for permission := range permissions {
			if permission == string(resource.canView) {
				ret = append(ret, roleName)
			}
		}
	}
	return ret
}
