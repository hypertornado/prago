package prago

import (
	"fmt"
)

var nobodyPermission Permission = "nobody"
var loggedPermission Permission = "logged"
var sysadminPermission Permission = "sysadmin"
var sysadminRoleName = "sysadmin"

type accessManager struct {
	roles       map[string]map[Permission]bool
	permissions map[Permission]bool
}

func (app *App) initAccessManager() {
	app.accessManager = &accessManager{
		roles:       make(map[string]map[Permission]bool),
		permissions: make(map[Permission]bool),
	}

	app.AddPermission(nobodyPermission)
	app.AddRole(sysadminRoleName, nil)
	app.AddPermission(loggedPermission)
	app.AddPermission(sysadminPermission)
	app.AddRole("", []Permission{loggedPermission})
}

//Permission for access
type Permission string

/*
func (app App) getSysadminPermissions() []string {
	m := map[string]bool{}
	for _, v1 := range app.accessManager.roles {
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
}*/

func (app *App) getRoleFieldTypeData() [][2]string {
	roleNames := []string{""}
	for k := range app.accessManager.roles {
		roleNames = append(roleNames, k)
	}

	vals := [][2]string{}
	for _, v := range roleNames {
		vals = append(vals, [2]string{v, v})
	}
	return vals
}

func (app *App) createRoleFieldType() FieldType {
	var fp = func(field, *User) interface{} {
		roleNames := []string{""}
		for k := range app.accessManager.roles {
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
func (app *App) AddRole(role string, permissions []Permission) {
	_, ok := app.accessManager.roles[role]
	if ok {
		panic(fmt.Sprintf("Role '%s' already added", role))
	}

	perms := map[Permission]bool{}
	for _, v := range permissions {
		if v == nobodyPermission {
			panic(fmt.Sprintf("Can't add permission nobody to role %s.", role))
		}
		if !app.accessManager.permissions[Permission(v)] {
			panic(fmt.Sprintf("Permission '%s' not found, you should add it before adding to role.", v))
		}
		perms[v] = true
	}
	perms[loggedPermission] = true
	app.accessManager.roles[role] = perms
}

//AddPermission adds permission to admin
func (app *App) AddPermission(permission Permission) {
	if app.accessManager.permissions[Permission(permission)] {
		panic(fmt.Sprintf("Permission '%s' already added", permission))
	}
	if permission != nobodyPermission {
		app.accessManager.roles[sysadminRoleName][permission] = true
	}
	app.accessManager.permissions[Permission(permission)] = true
}

func (app *App) authorize(user *User, permission Permission) bool {
	return app.accessManager.roles[user.Role][permission]
}
func (request *Request) authorize(permission Permission) bool {
	return request.app.authorize(request.user, permission)
}

func (app App) getResourceViewRoles(resource Resource) []string {
	var ret []string
	for roleName, permissions := range app.accessManager.roles {
		for permission := range permissions {
			if permission == resource.canView {
				ret = append(ret, roleName)
			}
		}
	}
	return ret
}
