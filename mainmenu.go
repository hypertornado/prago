package prago

import (
	"strings"

	"github.com/hypertornado/prago/messages"
	"github.com/hypertornado/prago/utils"
)

type mainMenu struct {
	Logo             string
	Language         string
	URLPrefix        string
	AdminHomepageURL string
	SearchQuery      string
	Sections         []mainMenuSection
	HasSearch        bool
}

type mainMenuSection struct {
	Name  string
	Items []mainMenuItem
}

type mainMenuItem struct {
	Name     string
	Subname  string
	URL      string
	Selected bool
}

func (menu mainMenu) GetTitle() string {
	for _, v := range menu.Sections {
		for _, v2 := range v.Items {
			if v2.Selected {
				return v2.Name
			}
		}
	}
	return ""
}

func (app *App) getMainMenu(request Request) (ret mainMenu) {
	user := request.GetUser()

	var selectedAdminSection bool
	if request.Request().URL.Path == app.GetAdminURL("") {
		selectedAdminSection = true
	}

	adminSectionName := app.HumanName
	if app.Logo != "" {
		adminSectionName = ""
	}
	adminSection := mainMenuSection{
		Name: adminSectionName,
		Items: []mainMenuItem{
			{
				Name:     messages.Messages.Get(user.Locale, "admin_signpost"),
				URL:      app.GetAdminURL(""),
				Selected: selectedAdminSection,
			},
		},
	}

	var selectedTasks bool
	if request.Request().URL.Path == app.GetAdminURL("_tasks") {
		selectedTasks = true
	}
	adminSection.Items = append(adminSection.Items, mainMenuItem{
		Name:     messages.Messages.Get(user.Locale, "tasks"),
		URL:      app.GetAdminURL("_tasks"),
		Selected: selectedTasks,
	})

	for _, v := range app.rootActions {
		if v.Method != "GET" && v.Method != "" {
			continue
		}

		var selected bool
		fullURL := app.GetAdminURL(v.URL)
		if request.Request().URL.Path == fullURL {
			selected = true
		}

		adminSection.Items = append(adminSection.Items, mainMenuItem{
			Name:     v.getName(user.Locale),
			URL:      fullURL,
			Selected: selected,
		})
	}

	ret.Sections = append(ret.Sections, adminSection)

	resourceSection := mainMenuSection{
		Name: messages.Messages.Get(user.Locale, "admin_tables"),
	}
	for _, resource := range app.getSortedResources(user.Locale) {
		if app.Authorize(user, resource.CanView) {
			resourceURL := app.GetAdminURL(resource.ID)

			var selected bool
			if request.Request().URL.Path == resourceURL {
				selected = true
			}
			if strings.HasPrefix(request.Request().URL.Path, resourceURL+"/") {
				selected = true
			}

			resourceSection.Items = append(resourceSection.Items, mainMenuItem{
				Name:     resource.HumanName(user.Locale),
				Subname:  utils.HumanizeNumber(resource.getCachedCount()),
				URL:      resourceURL,
				Selected: selected,
			})
		}
	}
	ret.Sections = append(ret.Sections, resourceSection)

	var userSettingsSection bool
	if request.Request().URL.Path == app.GetAdminURL("user/settings") || request.Request().URL.Path == app.GetAdminURL("user/password") {
		userSettingsSection = true
	}
	userName := user.Name
	if userName == "" {
		userName = user.Email
	}
	randomness := app.Config.GetString("random")
	userSection := mainMenuSection{
		Name: userName,
		Items: []mainMenuItem{
			{
				Name:     messages.Messages.Get(user.Locale, "admin_settings"),
				URL:      app.GetAdminURL("user/settings"),
				Selected: userSettingsSection,
			},
			{
				Name: messages.Messages.Get(user.Locale, "admin_homepage"),
				URL:  "/",
			},
			{
				Name: messages.Messages.Get(user.Locale, "admin_log_out"),
				URL:  app.GetAdminURL("logout") + "?_csrfToken=" + user.csrfToken(randomness),
			},
		},
	}
	ret.Sections = append(ret.Sections, userSection)

	ret.Logo = app.Logo
	ret.URLPrefix = adminPathPrefix
	ret.Language = user.Locale
	ret.AdminHomepageURL = app.GetAdminURL("")

	if app.search != nil {
		ret.HasSearch = true
	}

	return ret
}
