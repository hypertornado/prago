package admin

type exportFormData struct {
	formats []string
}

func (cache structCache) getExportFormData(user User, visible structFieldFilter) exportFormData {
	ret := exportFormData{
		formats: []string{"csv", "json"},
	}
	return ret
}
