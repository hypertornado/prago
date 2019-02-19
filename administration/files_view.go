package administration

type filesViewData struct {
	Error       string
	UUID        string
	OriginalURL string
	MediumURL   string
	SmallURL    string
	Paths       []filesViewDataPath
}

type filesViewDataPath struct {
	Name string
	URL  string
}

func filesViewDataSource(resource Resource, user User, f Field, value interface{}) interface{} {
	ret := filesViewData{}

	var file File
	err := resource.Admin.Query().WhereIs("UID", value.(string)).Get(&file)
	if err != nil {
		ret.Error = "Can't find file."
		return ret
	}

	ret.UUID = file.UID

	ret.Paths = []filesViewDataPath{
		{"original", file.GetOriginal()},
	}

	ret.OriginalURL = file.GetOriginal()

	if file.IsImage() {
		ret.MediumURL = file.GetMedium()
		ret.SmallURL = file.GetSmall()
		ret.Paths = append(ret.Paths,
			filesViewDataPath{"large", file.GetLarge()},
			filesViewDataPath{"medium", file.GetMedium()},
			filesViewDataPath{"small", file.GetSmall()},
			filesViewDataPath{"metadata", file.GetMetadataPath()},
		)
	}

	return ret
}
