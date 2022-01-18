package prago

import (
	"fmt"
	"sort"
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

	app.Permission(nobodyPermission)
	app.Role(sysadminRoleName, nil)
	app.Permission(loggedPermission)
	app.Permission(sysadminPermission)
	app.Role("", []Permission{loggedPermission})
}

//Permission for access
type Permission string

func (app *App) validatePermission(permission Permission) error {
	if !app.accessManager.permissions[permission] {
		return fmt.Errorf("unknown permission '%s'", permission)
	}
	return nil
}

func (app *App) createRoleFieldType() *fieldType {
	var fp = func(field, *user) interface{} {
		var roleNames []string
		for k := range app.accessManager.roles {
			roleNames = append(roleNames, k)
		}
		sort.Strings(roleNames)

		vals := [][2]string{}
		for _, v := range roleNames {
			vals = append(vals, [2]string{v, v})
		}
		return vals
	}
	return &fieldType{
		formTemplate:   "admin_item_select",
		formDataSource: fp,

		filterLayoutTemplate: "filter_layout_select",
	}
}

//Role adds role to admin
func (app *App) Role(role string, permissions []Permission) *App {
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
	return app
}

//Permission adds permission to admin
func (app *App) Permission(permission Permission) *App {
	if app.accessManager.permissions[Permission(permission)] {
		panic(fmt.Sprintf("Permission '%s' already added", permission))
	}
	if permission != nobodyPermission {
		app.accessManager.roles[sysadminRoleName][permission] = true
	}
	app.accessManager.permissions[Permission(permission)] = true
	return app
}

func (app *App) authorize(user *user, permission Permission) bool {
	return app.accessManager.roles[user.Role][permission]
}

func (resource *Resource[T]) getResourceViewRoles() []string {
	var ret []string
	for roleName, permissions := range resource.app.accessManager.roles {
		for permission := range permissions {
			if permission == resource.getPermissionView() {
				ret = append(ret, roleName)
			}
		}
	}
	return ret
}
