package prago

import "html/template"

// FormItem represents item of form
type FormItem struct {
	ID                 string
	Icon               string
	Name               string
	Description        string
	DescriptionsBefore []string
	DescriptionsAfter  []string
	Placeholder        string
	Required           bool
	Focused            bool
	Readonly           bool
	HiddenName         bool
	Hidden             bool
	Template           string
	Value              string
	Data               any

	Style string
	Color string

	Content template.HTML

	TextOver string

	UUID string
	form *Form

	Autocomplete string
	InputMode    string

	HelpURL string

	FileMultiple bool
	FileAccept   string

	FormFilterID string

	SuggestionURL string

	IsWide                 bool
	PreventPasswordManager bool

	UnitBefore string
	UnitAfter  string
}

func (fi *FormItem) GetContent() template.HTML {
	if fi.Content != "" {
		return fi.Content
	}
	if fi.Template != "" {
		return fi.form.app.adminTemplates.ExecuteToHTML(fi.Template, fi)
	}
	return ""
}

func (fi *FormItem) GetColor() string {
	if fi.Color != "" {
		return fi.Color
	}
	return getStyleColor(fi.Style)
}

func (fi *FormItem) SetFromFilter(formFilter *FormFilter) *FormItem {
	fi.FormFilterID = formFilter.uuid
	return fi
}

func (fi *FormItem) SetUUID() *FormItem {
	fi.UUID = "id-" + randomString(5)
	return fi
}

func (fi *FormItem) SetValue(value string) *FormItem {
	fi.Value = value
	return fi
}

func (fi *FormItem) SetDescription(description string) *FormItem {
	fi.Description = description
	return fi
}

func (fi *FormItem) SetContent(content template.HTML) *FormItem {
	fi.Content = content
	return fi
}

func (fi *FormItem) SetTextOver(textOver string) *FormItem {
	if textOver == "" {
		textOver = fi.Name
	}
	fi.TextOver = textOver
	return fi
}

func (fi *FormItem) SetDescriptionsBefore(before []string) *FormItem {
	fi.DescriptionsBefore = before
	return fi
}

func (fi *FormItem) SetDescriptionsAfter(after []string) *FormItem {
	fi.DescriptionsAfter = after
	return fi
}

func (fi *FormItem) SetUnitBefore(unit string) *FormItem {
	fi.UnitBefore = unit
	return fi
}

func (fi *FormItem) SetUnitAfter(unit string) *FormItem {
	fi.UnitAfter = unit
	return fi
}

func (fi *FormItem) SetIcon(icon string) *FormItem {
	fi.Icon = icon
	return fi
}

func (fi *FormItem) SetFocused() *FormItem {
	fi.Focused = true
	return fi
}

func (fi *FormItem) SetReadonly() *FormItem {
	fi.Readonly = true
	return fi
}
