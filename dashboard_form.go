package prago

type DashboardFormItem struct {
	ID       string
	Name     string
	Template string
	Value    string
	Options  [][2]string
}

func (dashboard *Dashboard) AddFormItemOptions(id, name string, options [][2]string) {

}
