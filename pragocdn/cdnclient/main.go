package cdnclient

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type CDNUploadData struct {
	UUID      string
	Extension string
	IsImage   bool
}

type CDNAccount struct {
	URL      string
	Account  string
	Password string
	client   http.Client
}

func NewCDNAccount(url, account, password string) CDNAccount {
	return CDNAccount{
		URL:      url,
		Account:  account,
		Password: password,
		client:   http.Client{},
	}
}

func GetHash(account, password, uuid, format, filename string) string {
	h := sha256.New()
	h.Write([]byte(account))
	h.Write([]byte(password))
	h.Write([]byte(uuid))
	h.Write([]byte(format))
	h.Write([]byte(filename))
	return fmt.Sprintf("%x", h.Sum(nil)[0:5])
}

func (a CDNAccount) getHash(uuid, format, filename string) string {

	return GetHash(
		a.Account,
		a.Password,
		uuid,
		format,
		filename,
	)
}

func (a CDNAccount) GetFileURL(uuid, filename string) string {
	hash := a.getHash(uuid, "file", filename)
	return fmt.Sprintf("%s/%s/%s/file/%s/%s", a.URL, a.Account, uuid, hash, filename)
}

func (a CDNAccount) GetImageURL(uuid, filename string, size int) string {
	hash := a.getHash(uuid, strconv.Itoa(size), filename)
	return fmt.Sprintf("%s/%s/%s/%d/%s/%s", a.URL, a.Account, uuid, size, hash, filename)
}

func (a CDNAccount) GetImageCropURL(uuid, filename string, width, height int) string {
	hash := a.getHash(uuid, fmt.Sprintf("%dx%d", width, height), filename)
	return fmt.Sprintf("%s/%s/%s/%dx%d/%s/%s", a.URL, a.Account, uuid, width, height, hash, filename)
}

func (a CDNAccount) UploadFileFromPath(filePath string) (*CDNUploadData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening file: %s", err)
	}
	defer file.Close()

	extension := filepath.Ext(filePath)
	extension = strings.Replace(extension, ".", "", -1)
	return a.UploadFile(file, extension)
}

func (a CDNAccount) DeleteFile(uuid string) error {
	u, err := url.Parse(fmt.Sprintf("%s/%s/%s", a.URL, a.Account, uuid))
	if err != nil {
		return fmt.Errorf("parsing url: %s", err)
	}

	req := &http.Request{}
	req.Method = "DELETE"
	req.URL = u
	req.Header = map[string][]string{}
	req.Header.Set("X-Authorization", a.Password)

	response, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return fmt.Errorf("wrong request with status code %d", response.StatusCode)
	}
	return nil
}

func (a CDNAccount) UploadFile(reader io.ReadCloser, extension string) (*CDNUploadData, error) {

	u, err := url.Parse(fmt.Sprintf("%s/%s/upload/%s", a.URL, a.Account, extension))
	if err != nil {
		return nil, fmt.Errorf("parsing url: %s", err)
	}

	req := &http.Request{}
	req.Method = "POST"
	req.Body = reader
	req.URL = u
	req.Header = map[string][]string{}
	req.Header.Set("X-Authorization", a.Password)

	response, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("opening file: %s", err)
	}
	defer response.Body.Close()

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response data: %s", err)
	}

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("wrong status code %d: %s", response.StatusCode, string(data))
	}

	var ret CDNUploadData
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling file: %s", err)
	}

	return &ret, nil
}
