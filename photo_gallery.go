package prago

import (
	"strings"
)

type PhotoGalleryImage struct {
	URL   string
	UUID  string
	Title string
}

func (app *App) GetPhotoGalleryImageData(idsStr string) (ret []*PhotoGalleryImage) {

	if len(idsStr) == 0 {
		return nil
	}
	ids := strings.SplitSeq(idsStr, ",")

	for v := range ids {
		file := Query[File](app).Is("uid", v).First()
		if file != nil {
			if file.IsImage() {

				item := &PhotoGalleryImage{
					URL:   file.GetGiant(),
					UUID:  file.UID,
					Title: file.Description,
				}
				ret = append(ret, item)
			}
		}
	}
	return
}
