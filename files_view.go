package prago

import (
	"strings"
)

type filesViewData struct {
	Error       string
	UUID        string
	Filename    string
	Filesize    int64
	OriginalURL string
	MediumURL   string
	SmallURL    string
	IsImage     bool
	Paths       []filesViewDataPath
}

type filesViewDataPath struct {
	Name string
	URL  string
}

func getFilesViewData(app *App, uid string) (ret filesViewData) {
	var file File
	err := app.Query().WhereIs("UID", uid).Get(&file)
	if err != nil {
		ret.Error = "Can't find file."
		return ret
	}

	metadata, err := filesCDN.GetMetadata(uid)
	if err != nil {
		ret.Error = "Can't get medtadata"
		return ret
	}

	ret.UUID = file.UID
	ret.Filesize = metadata.Filesize

	ret.Paths = []filesViewDataPath{
		{"original", file.GetOriginal()},
	}

	ret.OriginalURL = file.GetOriginal()

	if file.isImage() {
		ret.MediumURL = file.GetMedium()
		ret.SmallURL = file.GetSmall()
		ret.Paths = append(ret.Paths,
			filesViewDataPath{"large", file.GetLarge()},
			filesViewDataPath{"medium", file.GetMedium()},
			filesViewDataPath{"small", file.GetSmall()},
			filesViewDataPath{"metadata", file.GetMetadataPath()},
		)
		ret.IsImage = true
	}

	return ret

}

func filesViewDataSource(resource Resource, user *User, f field, value interface{}) interface{} {
	var ret []filesViewData
	ar := strings.Split(value.(string), ",")
	for _, v := range ar {
		item := getFilesViewData(resource.app, v)
		ret = append(ret, item)
	}
	return ret
}
