package prago

type breadcrumbs struct {
	Items []*breadcrumb
}

type breadcrumb struct {
	Logo  string
	Icon  string
	Image string
	Name  string
	URL   string
	Title string
}

func (pd *PageData) GetBreadcrumbs() *breadcrumbs {
	items := getBreadcrumbsFromMenuItems(pd.MenuItems)

	if len(items) > 0 && pd.LogoURL != "" {
		items = append([]*breadcrumb{{
			Logo: pd.LogoURL,
			URL:  items[0].URL,
		}}, items...)
	}

	if len(items) > 0 {
		items = items[0 : len(items)-1]
	}

	return &breadcrumbs{
		Items: items,
	}
}

func getBreadcrumbsFromMenuItems(items []*MenuItem) []*breadcrumb {
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

func menuItemToBreadcrumb(menuItem *MenuItem, selected bool) *breadcrumb {
	return &breadcrumb{
		Icon:  menuItem.Icon,
		Image: menuItem.Image,
		Name:  menuItem.Name,
		URL:   menuItem.URL,
		Title: menuItem.Name,
		//Selected: selected,
	}
}
