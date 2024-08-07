package prago

import (
	"fmt"
	"sort"
)

// Permission for access
type Permission string

const (
	everybodyPermission Permission = "everybody"
	nobodyPermission    Permission = "nobody"
	loggedPermission    Permission = "logged"
	sysadminPermission  Permission = "sysadmin"
	sysadminRoleName               = "sysadmin"
)

type accessManager struct {
	roles       map[string]map[Permission]bool
	permissions map[Permission]bool
}

func (app *App) initAccessManager() {
	app.accessManager = &accessManager{
		roles:       make(map[string]map[Permission]bool),
		permissions: make(map[Permission]bool),
	}

	app.Permission(everybodyPermission)
	app.Permission(nobodyPermission)
	app.Role(sysadminRoleName, nil)
	app.Permission(loggedPermission)
	app.Permission(sysadminPermission)
	app.Role("", []Permission{loggedPermission})
}

func (app *App) validatePermission(permission Permission) error {
	if !app.accessManager.permissions[permission] {
		return fmt.Errorf("unknown permission '%s'", permission)
	}
	return nil
}

func (app *App) createRoleFieldType() *fieldType {
	var formDataSource = func(*Field, UserData, string) interface{} {
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
		formTemplate:   "form_input_select",
		formDataSource: formDataSource,

		filterLayoutTemplate: "filter_layout_select",
		filterLayoutDataSource: func(f *Field, ud UserData) interface{} {
			return formDataSource(f, ud, "")
		},
	}
}

// Role adds role to admin
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

// Permission adds permission to admin
func (app *App) Permission(permission Permission) *App {
	if app.accessManager.permissions[Permission(permission)] {
		panic(fmt.Sprintf("Permission '%s' already added", permission))
	}
	if permission != nobodyPermission && permission != everybodyPermission {
		app.accessManager.roles[sysadminRoleName][permission] = true
	}
	app.accessManager.permissions[Permission(permission)] = true
	return app
}

func (app *App) authorize(isLogged bool, role string, permission Permission) bool {
	if permission == everybodyPermission {
		return true
	}
	if permission == loggedPermission && isLogged {
		return true
	}
	if !isLogged {
		return false
	}
	ret := app.accessManager.roles[role][permission]
	return ret
}
