package prago

type breadcrumbs struct {
	Items []*breadcrumb
}

type breadcrumb struct {
	Logo     string
	Icon     string
	Image    string
	Name     string
	URL      string
	Title    string
	Selected bool
}

func (menu menu) GetBreadcrumbs() *breadcrumbs {
	items := getBreadcrumbsFromMenuItems(menu.Items)

	items = append([]*breadcrumb{{
		Logo: "/admin/logo",
		URL:  "/admin",
	}}, items...)

	return &breadcrumbs{
		Items: items,
	}
}

func getBreadcrumbsFromMenuItems(items []*menuItem) []*breadcrumb {
	for _, v := range items {
		if v.Selected {
			return []*breadcrumb{menuItemToBreadcrumb(v, true)}
		}
		items := getBreadcrumbsFromMenuItems(v.Subitems)
		if len(items) > 0 {
			return append([]*breadcrumb{menuItemToBreadcrumb(v, false)}, items...)
		}
	}
	return nil
}

func menuItemToBreadcrumb(menuItem *menuItem, selected bool) *breadcrumb {
	return &breadcrumb{
		Icon:     menuItem.Icon,
		Image:    menuItem.Image,
		Name:     menuItem.Name,
		URL:      menuItem.URL,
		Title:    menuItem.Name,
		Selected: selected,
	}
}
