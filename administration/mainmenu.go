package administration

import (
	"strings"

	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/administration/messages"
	"github.com/hypertornado/prago/utils"
)

type MainMenu struct {
	Logo             string
	AdminHomepageURL string
	SearchQuery      string
	Sections         []MainMenuSection
	HasSearch        bool
}

type MainMenuSection struct {
	Name  string
	Items []MainMenuItem
}

type MainMenuItem struct {
	Name     string
	Subname  string
	URL      string
	Selected bool
}

func (admin *Administration) getMainMenu(request prago.Request) (ret MainMenu) {
	user := GetUser(request)

	var selectedAdminSection bool
	if request.Request().URL.Path == admin.GetURL("") {
		selectedAdminSection = true
	}

	adminSectionName := admin.HumanName
	if admin.Logo != "" {
		adminSectionName = ""
	}
	adminSection := MainMenuSection{
		Name: adminSectionName,
		Items: []MainMenuItem{
			{
				Name:     messages.Messages.Get(user.Locale, "admin_signpost"),
				URL:      admin.GetURL(""),
				Selected: selectedAdminSection,
			},
		},
	}

	for _, v := range admin.rootActions {
		if v.Method != "GET" && v.Method != "" {
			continue
		}

		var selected bool
		fullURL := admin.GetURL(v.URL)
		if request.Request().URL.Path == fullURL {
			selected = true
		}

		adminSection.Items = append(adminSection.Items, MainMenuItem{
			Name:     v.getName(user.Locale),
			URL:      fullURL,
			Selected: selected,
		})
	}

	ret.Sections = append(ret.Sections, adminSection)

	resourceSection := MainMenuSection{
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

			resourceSection.Items = append(resourceSection.Items, MainMenuItem{
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
	randomness := admin.App.Config.GetString("random")
	userSection := MainMenuSection{
		Name: userName,
		Items: []MainMenuItem{
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
