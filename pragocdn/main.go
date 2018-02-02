package main

import (
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/extensions"
	"github.com/hypertornado/prago/pragocdn/cdnclient"
	"github.com/hypertornado/prago/utils"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const version = "1.0.1"

var config CDNConfig

var accounts = map[string]*CDNConfigAccount{}
var homePath = os.Getenv("HOME")

var uuidRegex = regexp.MustCompile("^[a-zA-Z0-9]{10,}$")
var filenameRegex = regexp.MustCompile("^[a-zA-Z0-9_-]{1,50}$")
var extensionRegex = regexp.MustCompile("^[a-zA-Z0-9]{1,10}$")

func main() {
	var err error
	config, err = loadCDNConfig()
	if err != nil {
		panic(err)
	}

	for k, v := range config.Accounts {
		err := prepareAccountDirectories(v.Name)
		if err != nil {
			panic(err)
		}
		accounts[v.Name] = &config.Accounts[k]
	}

	app := prago.NewApp("pragocdn", version)
	app.AddMiddleware(extensions.BuildMiddleware{})
	app.AddMiddleware(prago.MiddlewareServer{Fn: start})
	prago.Must(app.Init())
}

func uploadFile(account CDNConfigAccount, extension string, inData io.Reader) (*cdnclient.CDNUploadData, error) {
	uuid := utils.RandomString(20)
	dirPath := getFileDirectoryPath(account.Name, uuid)

	err := os.MkdirAll(dirPath, 0777)
	if err != nil {
		return nil, err
	}

	filePath := getFilePath(account.Name, uuid, extension)

	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(file, inData)
	if err != nil {
		return nil, err
	}

	return &cdnclient.CDNUploadData{
		UUID:      uuid,
		Extension: extension,
	}, nil
}

func start(app *prago.App) {
	app.MainController().Get("/", func(request prago.Request) {
		out := fmt.Sprintf("Prago CDN\nhttps://www.prago-cdn.com\nversion %s\nadmin Ondřej Odcházel, https//www.odchazel.com", version)
		http.Error(request.Response(), out, 200)
		request.SetProcessed()
	})

	app.MainController().Post("/:account/upload/:extension", func(request prago.Request) {
		defer request.Request().Body.Close()
		accountName := request.Params().Get("account")
		account := accounts[accountName]
		if account == nil {
			panic("no account")
		}

		authorization := request.Request().Header.Get("X-Authorization")
		if account.Password != authorization {
			panic("wrong authorization")
		}

		extension := normalizeExtension(request.Params().Get("extension"))
		if !extensionRegex.MatchString(extension) {
			panic("wrong extension")
		}

		data, err := uploadFile(*account, extension, request.Request().Body)
		if err != nil {
			panic(err)
		}
		prago.WriteAPI(request, data, 200)
	})

	app.MainController().Get("/:account/:uuid/:format/:hash/:name", func(request prago.Request) {
		errCode, err, stream, mimeExtension, size := getFile(
			request.Params().Get("account"),
			request.Params().Get("uuid"),
			request.Params().Get("format"),
			request.Params().Get("hash"),
			request.Params().Get("name"),
		)

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

		request.SetProcessed()
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

	app.MainController().Delete("/:account/:uuid", func(request prago.Request) {
		err := deleteFile(
			request.Params().Get("account"),
			request.Request().Header.Get("X-Authorization"),
			request.Params().Get("uuid"),
		)
		if err != nil {
			panic(err)
		}
		prago.WriteAPI(request, true, 200)
	})
}

var fileExtensionMap = map[string]string{
	"jpeg": "jpg",
}

func deleteFile(accountName, password, uuid string) error {
	account := accounts[accountName]
	if account == nil {
		return errors.New("account not found")
	}

	if !uuidRegex.MatchString(uuid) {
		return errors.New("wrongs uuid format: " + uuid)
	}

	if account.Password != password {
		return errors.New("wrong password")
	}

	dirPath := getFileDirectoryPath(accountName, uuid)
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return err
	}

	var name string
	for _, v := range files {
		fileName := v.Name()
		if strings.HasPrefix(fileName, uuid+".") {
			name = fileName
			break
		}
	}

	if name == "" {
		return errors.New("no file found " + name)
	}

	filePath := fmt.Sprintf("%s/%s", dirPath, name)
	deletedDir := fmt.Sprintf("%s/.pragocdn/deleted/%s", homePath, accountName)

	cmd := exec.Command("mv", filePath, deletedDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func getFile(accountName, uuid, format, hash, name string) (eddCode int, err error, source io.Reader, mimeExtension string, size int64) {
	account := accounts[accountName]
	if account == nil {
		return 404, errors.New("account not found"), nil, "", -1
	}

	if !uuidRegex.MatchString(uuid) {
		return 404, errors.New("wrongs uuid format: " + uuid), nil, "", -1
	}

	splited := strings.Split(name, ".")
	if len(splited) != 2 {
		return 404, errors.New("wrong name format"), nil, "", -1
	}
	fileName := splited[0]
	fileExtension := splited[1]
	fileExtension = normalizeExtension(fileExtension)
	mimeExtension = mime.TypeByExtension("." + fileExtension)

	if !filenameRegex.MatchString(fileName) || !extensionRegex.MatchString(fileExtension) {
		return 404, errors.New("wrong name format"), nil, "", -1
	}

	expectedHash := cdnclient.GetHash(
		account.Name,
		account.Password,
		uuid,
		format,
		name,
	)
	if expectedHash != hash {
		return 498, errors.New("wrong hash"), nil, "", -1
	}

	originalPath := getFilePath(accountName, uuid, fileExtension)

	var path string
	if format == "file" {
		path = originalPath
	} else {
		path, err = convertedFilePath(accountName, uuid, fileExtension, format)
		if err != nil {
			return 404, err, nil, "", -1
		}
	}

	file, err := os.Open(path)
	if err != nil {
		return 500, err, nil, "", -1
	}

	stat, err := file.Stat()
	if err != nil {
		return 500, err, nil, "", -1
	}

	return 200, nil, file, mimeExtension, stat.Size()
}

func prepareAccountDirectories(name string) error {
	var err error
	err = os.MkdirAll(
		fmt.Sprintf("%s/.pragocdn/files/%s",
			homePath,
			name,
		),
		0777,
	)
	if err != nil {
		return fmt.Errorf("preparing files dir for %s: %s", name, err)
	}

	err = os.MkdirAll(
		fmt.Sprintf("%s/.pragocdn/cache/%s",
			homePath,
			name,
		),
		0777,
	)
	if err != nil {
		return fmt.Errorf("preparing cache dir for %s: %s", name, err)
	}

	err = os.MkdirAll(
		fmt.Sprintf("%s/.pragocdn/deleted/%s",
			homePath,
			name,
		),
		0777,
	)
	if err != nil {
		return fmt.Errorf("preparing deleted dir for %s: %s", name, err)
	}

	return nil
}

func getFileDirectoryPath(account, uuid string) string {
	firstPrefix := uuid[0:2]
	secondPrefix := uuid[2:4]
	return fmt.Sprintf("%s/.pragocdn/files/%s/%s/%s",
		homePath,
		account,
		firstPrefix,
		secondPrefix,
	)
}

func getFilePath(account, uuid, extension string) string {
	return fmt.Sprintf("%s/%s.%s",
		getFileDirectoryPath(account, uuid),
		uuid,
		extension,
	)
}

func getCacheDirectoryPath(account, uuid, format string) string {
	firstPrefix := uuid[0:2]
	secondPrefix := uuid[2:4]
	return fmt.Sprintf("%s/.pragocdn/cache/%s/%s/%s/%s",
		homePath,
		account,
		format,
		firstPrefix,
		secondPrefix,
	)
}

func getCacheFilePath(account, uuid, format, extension string) string {
	return fmt.Sprintf("%s/%s.%s",
		getCacheDirectoryPath(account, uuid, format),
		uuid,
		extension,
	)
}

var singleSizeRegexp = regexp.MustCompile("^[1-9][0-9]*$")
var sizeRegexp = regexp.MustCompile("^[1-9][0-9]{0,3}x[1-9][0-9]{0,3}$")

func isImageExtension(extension string) bool {
	if extension == "jpg" || extension == "png" {
		return true
	}
	return false
}

func normalizeExtension(extension string) string {
	extension = strings.ToLower(extension)
	fileExtensionChanged := fileExtensionMap[extension]
	if fileExtensionChanged != "" {
		extension = fileExtensionChanged
	}
	return extension
}

func convertedFilePath(account, uuid, extension, format string) (string, error) {
	if !isImageExtension(extension) {
		return "", errors.New("cant resize non images")
	}

	originalPath := getFilePath(account, uuid, extension)
	outputFilePath := getCacheFilePath(account, uuid, format, extension)
	outputDirectoryPath := getCacheDirectoryPath(account, uuid, format)

	if singleSizeRegexp.MatchString(format) {
		return outputFilePath, vipsThumbnail(originalPath, outputDirectoryPath, outputFilePath, format, extension, false)
	}

	if sizeRegexp.MatchString(format) {
		return outputFilePath, vipsThumbnail(originalPath, outputDirectoryPath, outputFilePath, format, extension, true)
	}

	return "", errors.New("wrong file convert format")
}

func vipsThumbnail(originalPath, outputDirectoryPath, outputFilePath, size, extension string, crop bool) error {
	_, err := os.Open(outputFilePath)
	if err == nil {
		return nil
	}

	err = os.MkdirAll(outputDirectoryPath, 0777)
	if err != nil {
		return err
	}

	outputParameters := "[strip]"
	if extension == "jpg" {
		outputParameters = "[optimize_coding,strip]"
	}

	cmdAr := []string{
		originalPath,
		"-s",
		size,
		"-o",
		outputFilePath + outputParameters,
	}

	if config.Profile != "" {
		cmdAr = append(cmdAr, "--delete", "--eprofile", config.Profile)
	}

	if crop {
		cmdAr = append(cmdAr, "-m", "attention")
	}

	cmd := exec.Command("vipsthumbnail", cmdAr...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func render404(request prago.Request) {
	http.Error(request.Response(), "Not Found", 404)
	request.SetProcessed()
}

func render498(request prago.Request) {
	http.Error(request.Response(), "Wrong Hash", 498)
	request.SetProcessed()
}
