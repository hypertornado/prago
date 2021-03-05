package prago

import (
	"strings"

	"github.com/hypertornado/prago/messages"
	"github.com/hypertornado/prago/utils"
)

type mainMenu struct {
	Logo             string
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

func (admin *App) getMainMenu(request Request) (ret mainMenu) {
	user := GetUser(request)

	var selectedAdminSection bool
	if request.Request().URL.Path == admin.GetURL("") {
		selectedAdminSection = true
	}

	adminSectionName := admin.HumanName
	if admin.Logo != "" {
		adminSectionName = ""
	}
	adminSection := mainMenuSection{
		Name: adminSectionName,
		Items: []mainMenuItem{
			{
				Name:     messages.Messages.Get(user.Locale, "admin_signpost"),
				URL:      admin.GetURL(""),
				Selected: selectedAdminSection,
			},
		},
	}

	var selectedTasks bool
	if request.Request().URL.Path == admin.GetURL("_tasks") {
		selectedTasks = true
	}
	adminSection.Items = append(adminSection.Items, mainMenuItem{
		Name:     messages.Messages.Get(user.Locale, "tasks"),
		URL:      admin.GetURL("_tasks"),
		Selected: selectedTasks,
	})

	for _, v := range admin.rootActions {
		if v.Method != "GET" && v.Method != "" {
			continue
		}

		var selected bool
		fullURL := admin.GetURL(v.URL)
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
	for _, resource := range admin.getSortedResources(user.Locale) {
		if admin.Authorize(user, resource.CanView) {
			resourceURL := admin.GetURL(resource.ID)

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
	if request.Request().URL.Path == admin.GetURL("user/settings") || request.Request().URL.Path == admin.GetURL("user/password") {
		userSettingsSection = true
	}
	userName := user.Name
	if userName == "" {
		userName = user.Email
	}
	randomness := admin.Config.GetString("random")
	userSection := mainMenuSection{
		Name: userName,
		Items: []mainMenuItem{
			{
				Name:     messages.Messages.Get(user.Locale, "admin_settings"),
				URL:      admin.GetURL("user/settings"),
				Selected: userSettingsSection,
			},
			{
				Name: messages.Messages.Get(user.Locale, "admin_homepage"),
				URL:  "/",
			},
			{
				Name: messages.Messages.Get(user.Locale, "admin_log_out"),
				URL:  admin.GetURL("logout") + "?_csrfToken=" + user.CSRFToken(randomness),
			},
		},
	}
	ret.Sections = append(ret.Sections, userSection)

	ret.Logo = admin.Logo
	ret.AdminHomepageURL = admin.GetURL("")

	if admin.search != nil {
		ret.HasSearch = true
	}

	return ret
}