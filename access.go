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
	roles          map[string]map[Permission]bool
	roleNames      map[string]func(string) string
	permissions    map[Permission]bool
	canManageRoles map[[2]string]bool
}

func (app *App) initAccessManager() {
	app.accessManager = &accessManager{
		roles:          make(map[string]map[Permission]bool),
		roleNames:      make(map[string]func(string) string),
		permissions:    make(map[Permission]bool),
		canManageRoles: make(map[[2]string]bool),
	}

	app.Permission(everybodyPermission)
	app.Permission(nobodyPermission)
	app.Role(sysadminRoleName, unlocalized("Sysadmin"), nil)
	app.Permission(loggedPermission)
	app.Permission(sysadminPermission)
	app.Role("", unlocalized("Bez oprávnění"), []Permission{loggedPermission})
}

func (app *App) validatePermission(permission Permission) error {
	if !app.accessManager.permissions[permission] {
		return fmt.Errorf("unknown permission '%s'", permission)
	}
	return nil
}

func (app *App) getRoleName(roleID string, locale string) string {
	name := app.accessManager.roleNames[roleID]
	if name == nil {
		return roleID
	}
	return name(locale)
}

func (app *App) createRoleFieldType() *fieldType {
	var formDataSource = func(field *Field, ud UserData, value string) interface{} {
		var roleIDs []string
		for k := range app.accessManager.roles {
			if k == "" {
				continue
			}
			roleIDs = append(roleIDs, k)
		}
		sort.Strings(roleIDs)

		vals := [][2]string{}
		for _, id := range roleIDs {
			vals = append(vals, [2]string{id, app.getRoleName(id, ud.Locale())})
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

		listCellDataSource: func(userData UserData, f *Field, value interface{}) *listCell {
			return &listCell{Name: app.getRoleName(value.(string), userData.Locale()), ItemID: f.id}
		},
	}
}

// Role adds role to admin
func (app *App) Role(roleID string, roleName func(string) string, permissions []Permission) *App {
	_, ok := app.accessManager.roles[roleID]
	if ok {
		panic(fmt.Sprintf("Role '%s' already added", roleID))
	}
	if roleName == nil {
		panic(fmt.Sprintf("No name set for role '%s'", roleID))
	}
	app.accessManager.roleNames[roleID] = roleName

	perms := map[Permission]bool{}
	for _, v := range permissions {
		if v == nobodyPermission {
			panic(fmt.Sprintf("Can't add permission nobody to role %s.", roleID))
		}
		if !app.accessManager.permissions[Permission(v)] {
			panic(fmt.Sprintf("Permission '%s' not found, you should add it before adding to role.", v))
		}
		perms[v] = true
	}
	perms[loggedPermission] = true
	app.accessManager.roles[roleID] = perms
	app.AddManagerOfRole(sysadminRoleName, roleID)
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

func (app *App) AddManagerOfRole(whoRole, whomRole string) {

	if app.accessManager.roles[whoRole] == nil {
		panic(fmt.Sprintf("Role '%s' does not exist", whoRole))
	}
	if app.accessManager.roles[whomRole] == nil {
		panic(fmt.Sprintf("Role '%s' does not exist", whomRole))
	}

	app.accessManager.canManageRoles[[2]string{whoRole, whomRole}] = true
}

func (app *App) canManageRole(who, whom string) bool {
	return app.accessManager.canManageRoles[[2]string{who, whom}]
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
