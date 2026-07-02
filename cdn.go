package prago

import (
	"fmt"

	"github.com/hypertornado/prago/pragocdn/cdnclient"
)

var filesCDN cdnclient.CDNAccount

func initCDN(app *App) {
	filesCDN = cdnclient.NewCDNAccount(
		app.mustGetSetting("cdn_url"),
		app.mustGetSetting("cdn_account"),
		app.mustGetSetting("cdn_password"),
	)
}

type cdnFileData struct {
	DownloadURL string
	MediumURL   string
	GiantURL    string
	OriginalURL string
}

func getCDNViewData(app *App, uid string) (ret *cdnFileData) {
	file := Query[File](app).Is("UID", uid).First()
	if file == nil {
		return nil
	}
	ret = &cdnFileData{}
	ret.DownloadURL = fmt.Sprintf("/admin/file/%d/download", file.ID)
	if file.IsImage() {
		ret.MediumURL = file.GetMedium()
		ret.GiantURL = file.GetGiant()
		ret.OriginalURL = file.GetOriginal()
	}
	return ret

}

func (file *File) getCDNNamedDownloadPaths() (ret [][2]string) {
	ret = append(ret, [2]string{"original", file.GetOriginal()})
	if file.IsImage() {
		ret = append(ret, [2]string{"large", file.GetLarge()})
		ret = append(ret, [2]string{"medium", file.GetMedium()})
		ret = append(ret, [2]string{"small", file.GetSmall()})
	}
	return ret
}

func (app *App) GetCDNVideoURL(uuid string) string {
	return filesCDN.GetVideoURL(uuid)
}
