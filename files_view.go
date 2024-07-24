package prago

import (
	"context"
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

func getFilesViewData(ctx context.Context, app *App, uid string) (ret filesViewData) {
	file := Query[File](app).Context(ctx).Is("UID", uid).First()
	if file == nil {
		ret.Error = "Can't find file."
		return ret
	}

	metadata, err := filesCDN.GetMetadata(uid)
	if err != nil {
		ret.Error = "Can't get metadata"
		return ret
	}

	ret.UUID = file.UID
	ret.Filesize = metadata.Filesize

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
			filesViewDataPath{"metadata", file.getMetadataPath()},
		)
		ret.IsImage = true
	}

	return ret

}

func filesViewDataSource(request *Request, f *Field, value interface{}) interface{} {
	app := f.resource.app
	var ret []filesViewData
	ar := strings.Split(value.(string), ",")
	for _, v := range ar {
		item := getFilesViewData(request.Request().Context(), app, v)
		ret = append(ret, item)
	}
	return ret
}
