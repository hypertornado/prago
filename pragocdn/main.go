package main

import (
	"compress/gzip"
	"embed"
	"errors"
	"fmt"
	"image"
	"io"
	"mime"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"

	_ "image/jpeg"
	_ "image/png"

	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/pragocdn/cdnclient"
)

const version = "2022.18"

var app *prago.App

//var homePath = os.Getenv("HOME")

var uuidRegex = regexp.MustCompile("^[a-zA-Z0-9]{10,}$")
var filenameRegex = regexp.MustCompile("^[a-zA-Z0-9_.-]{1,150}$")
var extensionRegex = regexp.MustCompile("^[a-zA-Z0-9]{1,10}$")

var vipsMutexes []*sync.Mutex

var fileExtensionMap = map[string]string{
	"jpeg": "jpg",
}

//go:embed resources/icons/*
var iconsFS embed.FS

func main() {

	for i := 0; i < 10; i++ {
		vipsMutexes = append(vipsMutexes, &sync.Mutex{})
	}

	app = prago.New("pragocdn", version)

	app.SetIcons(iconsFS, "resources/icons/")

	initCDNProjectResource()
	bindStats(app)
	bindCDNFiles(app)

	app.GET("/", func(request *prago.Request) {
		out := fmt.Sprintf("Prago CDN\nhttps://www.prago-cdn.com\nversion %s\nadmin Ondřej Odcházel, https//www.odchazel.com", version)
		http.Error(request.Response(), out, 200)
	})

	app.POST("/:account/upload/:extension", func(request *prago.Request) {
		//defer request.Request().Body.Close()
		project := getCDNProject(request.Param("account"))
		if project == nil {
			panic("no account")
		}

		if project.Password != request.Request().Header.Get("X-Authorization") {
			panic("wrong authorization")
		}

		extension := normalizeExtension(request.Param("extension"))
		data, err := project.uploadFile(extension, request.Request().Body)
		if err != nil {
			panic(err)
		}

		request.RenderJSON(data)
	})

	app.GET("/:account/:uuid/metadata", func(request *prago.Request) {
		cdnFile := getCDNFile(request.Param("account"), request.Param("uuid"))
		if cdnFile == nil {
			render404(request)
			return
		}
		metadata, err := cdnFile.getMetadata()
		if err != nil {
			panic(err)
		}
		request.RenderJSON(metadata)
	})

	app.GET("/:account/:uuid/:format/:hash/:name", func(request *prago.Request) {
		cdnFile := getCDNFile(request.Param("account"), request.Param("uuid"))
		if cdnFile == nil {
			render404(request)
			return
		}

		errCode, err, stream, mimeExtension, size := cdnFile.getFileDataInFormat(
			request.Param("format"),
			request.Param("hash"),
			request.Param("name"),
		)
		if stream != nil {
			defer stream.Close()
		}

		if err != nil {
			if app.DevelopmentMode() {
				panic(err)
			} else {
				path := request.Request().URL.Path
				fmt.Printf("getFile error on path %s: %s\n", path, err)
				return
			}
		}

		//https://gist.github.com/the42/1956518
		switch errCode {
		case 404:
			render404(request)
			return
		case 498:
			render498(request)
			return
		}

		if err != nil {
			panic(err)
		}

		request.Response().Header().Set("Cache-Control", "public, max-age=31536000")
		request.Response().Header().Set("X-Content-Type-Options", "nosniff")
		request.Response().Header().Set("Content-Type", mimeExtension)

		if strings.HasPrefix(mimeExtension, "text/") && strings.Contains(request.Request().Header.Get("Accept-Encoding"), "gzip") {
			request.Response().Header().Set("Content-Encoding", "gzip")
			gz := gzip.NewWriter(request.Response())
			_, err = io.Copy(gz, stream)
			defer gz.Close()
			if err != nil {
				panic(err)
			}
		} else {
			request.Response().Header().Set("Content-Length", fmt.Sprintf("%d", size))
			_, err = io.Copy(request.Response(), stream)
			if err != nil {
				panic(err)
			}
		}
	})

	app.DELETE("/:account/:uuid", func(request *prago.Request) {

		file := getCDNFile(request.Param("account"), request.Param("uuid"))

		err := file.deleteFile(
			request.Request().Header.Get("X-Authorization"),
		)
		if err != nil {
			panic(err)
		}
		request.RenderJSON(true)
	})

	app.Run()
}

func (account *CDNProject) uploadFile(extension string, inData io.Reader) (*cdnclient.CDNFileData, error) {
	uuid := RandomString(20)
	cdnFile := account.createFile(uuid, extension)

	//tempFilePath

	tmpPath := cdnFile.tempFilePath()
	defer os.Remove(tmpPath)
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return nil, err
	}
	defer tmpFile.Close()

	_, err = io.Copy(tmpFile, inData)
	if err != nil {
		return nil, err
	}

	cdnFile.Checksum = checksum(tmpPath)

	filePath := cdnFile.getDataPath()

	_, err = os.Stat(filePath)

	//file does not exist
	if err != nil {
		dirPath := cdnFile.getDataDirectoryPath()
		err := os.MkdirAll(dirPath, 0777)
		if err != nil {
			return nil, err
		}
		err = os.Rename(tmpPath, filePath)
		if err != nil {
			return nil, err
		}

		/*file, err := os.Create(filePath)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		_, err = io.Copy(file, inData)
		if err != nil {
			return nil, err
		}*/
	}

	cdnFile.update()
	return cdnFile.getMetadata()
}

func (file *CDNFile) deleteFile(password string) error {
	panic("not reimplemented")

	/*uuid := file.UUID
	project := file.Project()

	if !uuidRegex.MatchString(uuid) {
		return errors.New("wrongs uuid format: " + uuid)
	}

	if project.Password != password {
		return errors.New("wrong password")
	}

	filePath, _, err := file.getFilePathFromUUID()
	if err != nil {
		return err
	}

	deletedDir := fmt.Sprintf("%s/deleted/%s", cdnDirPath(), project.Name)

	cmd := exec.Command("mv", filePath, deletedDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()*/
}

/*func (file *CDNFile) getFilePathFromUUID() (filePath, extension string, err error) {
	dirPath := file.getDataDirectoryPath()
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return "", "", err
	}

	var name string
	for _, v := range files {
		fileName := v.Name()
		if strings.HasPrefix(fileName, file.UUID+".") {
			_, extension, _ = getNameAndExtension(fileName)
			name = fileName
			break
		}
	}

	if name == "" {
		return "", "", errors.New("no file found for uuid: " + file.UUID)
	}

	return fmt.Sprintf("%s/%s", dirPath, name), extension, nil
}*/

// TODO: cache somewhere
func (cdnFile *CDNFile) getMetadata() (*cdnclient.CDNFileData, error) {
	/*filePath, _, err := cdnFile.getFilePathFromUUID()
	if err != nil {
		return nil, fmt.Errorf("getting file path from uuid: %s", err)
	}*/

	file, err := os.Open(cdnFile.getDataPath())
	if err != nil {
		return nil, fmt.Errorf("opening file: %s", err)
	}
	defer file.Close()

	var width, height int

	if isImageExtension(cdnFile.Suffix) {
		i, _, err := image.Decode(file)
		if err == nil {
			bounds := i.Bounds()
			width = bounds.Max.X
			height = bounds.Max.Y
		} else {
			app.Log().Errorf("decoding: %s", err)
		}
	}

	filestat, _ := file.Stat()

	return &cdnclient.CDNFileData{
		UUID:      cdnFile.UUID,
		Extension: cdnFile.Suffix,
		IsImage:   isImageExtension(cdnFile.Suffix),
		Filesize:  filestat.Size(),
		Width:     int64(width),
		Height:    int64(height),
	}, nil

}

func (file *CDNFile) getFileDataInFormat(format, hash, name string) (errCode int, err error, source io.ReadCloser, mimeExtension string, size int64) {
	project := file.Project()

	_, fileExtension, err := getNameAndExtension(name)
	if err != nil {
		return 404, err, nil, "", -1
	}

	mimeExtension = mime.TypeByExtension("." + fileExtension)

	expectedHash := cdnclient.GetHash(
		project.Name,
		project.Password,
		file.UUID,
		format,
		name,
	)
	if expectedHash != hash {
		return 498, errors.New("wrong hash"), nil, "", -1
	}

	originalPath := file.getDataPath()

	var path string
	if format == "file" {
		path = originalPath
	} else {
		path, err = file.convertedFilePath(fileExtension, format)
		if err != nil {
			return 404, err, nil, "", -1
		}
		mimeExtension = mime.TypeByExtension(".webp")
	}

	f, err := os.Open(path)
	if err != nil {
		return 500, err, nil, "", -1
	}

	stat, err := f.Stat()
	if err != nil {
		return 500, err, nil, "", -1
	}

	return 200, nil, f, mimeExtension, stat.Size()
}

func (file *CDNFile) getDataDirectoryPath() string {
	checksum := file.Checksum
	firstPrefix := strings.ToLower(checksum[0:2])
	secondPrefix := strings.ToLower(checksum[2:4])
	return fmt.Sprintf("%s/data/%s/%s",
		cdnDirPath(),
		firstPrefix,
		secondPrefix,
	)
}

func (file *CDNFile) getDataPath() string {
	return fmt.Sprintf("%s/%s.data",
		file.getDataDirectoryPath(),
		file.Checksum,
	)
}

func (file *CDNFile) getFileDirectoryPathOLD() string {
	uuid := file.UUID
	firstPrefix := strings.ToLower(uuid[0:2])
	secondPrefix := strings.ToLower(uuid[2:4])
	return fmt.Sprintf("%s/files/%s/%s/%s",
		cdnDirPath(),
		file.Project().Name,
		firstPrefix,
		secondPrefix,
	)
}

func (file *CDNFile) getFilePathOLD() string {
	return fmt.Sprintf("%s/%s.%s",
		file.getFileDirectoryPathOLD(),
		file.UUID,
		file.Suffix,
	)
}

func (file *CDNFile) getCacheDirectoryPath(format string) string {
	checksum := file.Checksum
	firstPrefix := strings.ToLower(checksum[0:2])
	secondPrefix := strings.ToLower(checksum[2:4])
	return fmt.Sprintf("%s/cache/%s/%s/%s",
		cdnDirPath(),
		//file.Project().Name,
		format,
		firstPrefix,
		secondPrefix,
	)
}

func (file *CDNFile) getCacheFilePath(format, extension string) string {
	return fmt.Sprintf("%s/%s.%s",
		file.getCacheDirectoryPath(format),
		file.UUID,
		extension,
	)
}

var singleSizeRegexp = regexp.MustCompile("^[1-9][0-9]*$")
var sizeRegexp = regexp.MustCompile("^[1-9][0-9]{0,3}x[1-9][0-9]{0,3}$")

func isImageExtension(extension string) bool {
	if extension == "jpg" || extension == "png" || extension == "webp" {
		return true
	}
	return false
}

func (file *CDNFile) convertedFilePath(extension, format string) (string, error) {
	if !isImageExtension(extension) {
		return "", errors.New("cant resize non images")
	}

	originalPath := file.getDataPath()
	outputFilePath := file.getCacheFilePath(format, "webp")
	outputDirectoryPath := file.getCacheDirectoryPath(format)

	if singleSizeRegexp.MatchString(format) {
		return outputFilePath, vipsThumbnail(originalPath, outputDirectoryPath, outputFilePath, format, false)
	}

	if sizeRegexp.MatchString(format) {
		return outputFilePath, vipsThumbnail(originalPath, outputDirectoryPath, outputFilePath, format, true)
	}

	return "", errors.New("wrong file convert format")
}

func render404(request *prago.Request) {
	http.Error(request.Response(), "Not Found", 404)
}

func render498(request *prago.Request) {
	http.Error(request.Response(), "Wrong Hash", 498)
}
