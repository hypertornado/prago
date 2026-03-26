package main

import (
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"time"

	"github.com/hypertornado/prago"
)

type CDNVideo struct {
	ID       int64 `prago-order-desc:"true"`
	UUID     string
	Filename string

	CDNProject int64 `prago-type:"relation"`

	CreatedAt time.Time
	UpdatedAt time.Time `prago-can-view:"sysadmin"`
}

func (video *CDNVideo) getCDNFileURL() string {
	project := prago.Query[CDNProject](app).ID(video.CDNProject)
	if project == nil {
		return ""
	}
	return fmt.Sprintf("%s/%s/%s/%s", project.CDNEndpointURL, project.Name, video.UUID, video.Filename)
}

func (video *CDNVideo) getHLSURL() string {
	project := prago.Query[CDNProject](app).ID(video.CDNProject)
	if project == nil || video.UUID == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s/%s/videofiles/master.m3u8", project.CDNEndpointURL, project.Name, video.UUID)
}

func createMultiResolutionHLS_4K(videoURL, outputDir string) error {
	// Vytvoříme hlavní složku
	os.MkdirAll(outputDir, os.ModePerm)

	// Vytvoříme složky pro 5 různých rozlišení
	//os.MkdirAll(outputDir+"/stream_0", os.ModePerm) // 4K (2160p)
	os.MkdirAll(outputDir+"/stream_1", os.ModePerm) // 1080p
	os.MkdirAll(outputDir+"/stream_2", os.ModePerm) // 720p
	os.MkdirAll(outputDir+"/stream_3", os.ModePerm) // 480p
	os.MkdirAll(outputDir+"/stream_4", os.ModePerm) // 360p

	cmd := exec.Command("ffmpeg",
		"-i", videoURL,

		// --- Vytvoříme 5 streamů ze vstupu (Video i Audio pro každý) ---
		//"-map", "0:v:0", "-map", "0:a:0", // Stream 0 (4K)
		"-map", "0:v:0", "-map", "0:a:0", // Stream 1 (1080p)
		"-map", "0:v:0", "-map", "0:a:0", // Stream 2 (720p)
		"-map", "0:v:0", "-map", "0:a:0", // Stream 3 (480p)
		"-map", "0:v:0", "-map", "0:a:0", // Stream 4 (360p)

		// --- Nastavení kódování videa (Kodek H.264) ---
		"-c:v", "libx264",
		"-profile:v", "main", // Zajišťuje dobrou kompatibilitu napříč zařízeními

		// --- ZAROVNÁNÍ KLÍČOVÝCH SNÍMKŮ (Keyframes) ---
		// Předpokládáme 30 fps. Klíčový snímek každé 2 vteřiny = 60 snímků
		"-g", "60",
		"-keyint_min", "60",
		"-sc_threshold", "0",

		// --- Nastavení rozlišení a datového toku (Bitrate) ---

		// Stream 0: 4K (2160p) - Velmi vysoká kvalita
		//"-filter:v:0", "scale=-2:2160", "-b:v:0", "15000k", "-maxrate:v:0", "16000k", "-bufsize:v:0", "20000k",

		// Stream 1: 1080p (Full HD)
		"-filter:v:1", "scale=-2:1080", "-b:v:1", "5000k", "-maxrate:v:1", "5300k", "-bufsize:v:1", "7500k",

		// Stream 2: 720p (HD)
		"-filter:v:2", "scale=-2:720", "-b:v:2", "2800k", "-maxrate:v:2", "3000k", "-bufsize:v:2", "4200k",

		// Stream 3: 480p (SD)
		"-filter:v:3", "scale=-2:480", "-b:v:3", "1400k", "-maxrate:v:3", "1500k", "-bufsize:v:3", "2100k",

		// Stream 4: 360p (Mobilní data / pomalé připojení)
		"-filter:v:4", "scale=-2:360", "-b:v:4", "800k", "-maxrate:v:4", "850k", "-bufsize:v:4", "1200k",

		// --- Nastavení audia (AAC) ---
		"-c:a", "aac",
		//"-b:a:0", "192k", // 4K může mít kvalitní zvuk
		"-b:a:1", "192k", // 1080p také
		"-b:a:2", "128k", // 720p standard
		"-b:a:3", "96k", // 480p lehce osekaný
		"-b:a:4", "64k", // 360p úsporný (zaměřeno na data)

		// --- HLS Nastavení ---
		"-f", "hls",
		"-hls_time", "6",
		"-hls_playlist_type", "vod",
		"-hls_flags", "independent_segments",
		"-master_pl_name", "master.m3u8",

		// --- Propojení videa a audia do HLS variant ---
		"-var_stream_map", "v:0,a:0 v:1,a:1 v:2,a:2 v:3,a:3 v:4,a:4",

		// --- Výstupní názvy segmentů a playlistů ---
		"-hls_segment_filename", outputDir+"/stream_%v/data%03d.ts",
		outputDir+"/stream_%v/playlist.m3u8",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg selhal: %v\nVýstup: %s", err, string(output))
	}

	return nil
}

func (video *CDNVideo) splitToHLS() error {
	videoURL := video.getCDNFileURL()
	if videoURL == "" {
		return fmt.Errorf("no video URL available")
	}
	project := prago.Query[CDNProject](app).ID(video.CDNProject)
	if project == nil {
		return fmt.Errorf("project not found")
	}

	tempDir, err := os.MkdirTemp("", "hls-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}

	defer os.RemoveAll(tempDir)

	err = createMultiResolutionHLS_4K(videoURL, tempDir)
	if err != nil {
		return fmt.Errorf("can't create ffmpeg convert: %s", err)
	}

	return project.uploadFolderToSpaces(tempDir, fmt.Sprintf("%s/videofiles/", video.UUID))
}

func bindCDNVideos(app *prago.App) {
	videoResource := prago.NewResource[CDNVideo](app)
	videoResource.Name(unlocalized("CDN Video"), unlocalized("CDN Videa")).PermissionUpdate("nobody")

	prago.ItemStatistic(app, unlocalized("URL"), "sysadmin", func(video *CDNVideo) string {
		return video.getCDNFileURL()
	})

	prago.ItemStatistic(app, unlocalized("HLS URL"), "sysadmin", func(video *CDNVideo) string {
		return video.getHLSURL()
	})

	prago.ActionResourceItemUI(app, "videoplayer", func(item *CDNVideo, request *prago.Request) template.HTML {
		return app.GetAdminTemplates().ExecuteToHTML("view_video", item.getHLSURL())
	}).Permission("sysadmin").Name(unlocalized("Přehrávač"))

	prago.ActionResourceItemForm(app, "delete-item", func(video *CDNVideo, form *prago.Form, request *prago.Request) {
		form.AddDeleteSubmit("Smazat video")
	}, func(video *CDNVideo, fv prago.FormValidation, request *prago.Request) {
		err := prago.DeleteWithLog(video, request)
		if err != nil {
			fv.AddError(err.Error())
			return
		}
		request.AddFlashMessage("Video smazáno")
		fv.Redirect("/admin/cdnvideo")
	}).Permission("sysadmin").Name(unlocalized("Smazat video"))

	prago.PreviewURLFunction(app, func(video *CDNVideo) string {
		return video.getCDNFileURL()
	})

	prago.ActionForm(app, "upload-video",
		func(form *prago.Form, request *prago.Request) {
			form.AddRelation("project", "Projekt", "cdnproject")

			fileInput := form.AddFileInput("file", "Video soubor")
			fileInput.FileAccept = ".mp4"
			form.AddSubmit("Nahrát video")
		},
		func(fv prago.FormValidation, request *prago.Request) {
			project := prago.Query[CDNProject](app).ID(request.Param("project"))
			if project == nil {
				fv.AddItemError("project", "Vyberte projekt")
			}

			multipartFiles := request.Request().MultipartForm.File["file"]
			if len(multipartFiles) != 1 {
				fv.AddItemError("file", "Vyberte video soubor")
			}

			if !fv.Valid() {
				return
			}

			fileHeader := multipartFiles[0]
			originalFilename := fileHeader.Filename

			uuid := RandomString(20)

			openedFile, err := fileHeader.Open()
			if err != nil {
				fv.AddError(fmt.Sprintf("Chyba při otevírání souboru: %s", err))
				return
			}
			defer openedFile.Close()

			key := fmt.Sprintf("%s/%s", uuid, originalFilename)
			if err := project.uploadVideoToSpaces(key, originalFilename, openedFile, fileHeader.Size); err != nil {
				fv.AddError(fmt.Sprintf("Chyba při nahrávání do Spaces: %s", err))
				return
			}

			video := &CDNVideo{
				UUID:       uuid,
				Filename:   originalFilename,
				CDNProject: project.ID,
			}
			if err := prago.CreateWithLog(video, request); err != nil {
				fv.AddError(fmt.Sprintf("Chyba při vytváření záznamu: %s", err))
				return
			}

			if err := video.splitToHLS(); err != nil {
				fv.AddError(fmt.Sprintf("Chyba při rozdělování na HLS: %s", err))
				return
			}

			fv.Redirect(fmt.Sprintf("/admin/cdnvideo/%d", video.ID))
		},
	).Name(unlocalized("Nahrát video")).Permission("sysadmin")
}
