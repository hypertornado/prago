package main

//https://stackoverflow.com/questions/29476611/go-simple-api-gateway-proxy
//https://golang.org/pkg/net/http/httputil/#NewSingleHostReverseProxy

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type Server struct {
	config *ConfigServer
	port   int
}

func NewServer(config ConfigServer) *Server {
	ret := &Server{
		config: &config,
	}
	return ret
}

func (server *Server) Start() error {
	largest, err := server.GetLatestVersion()
	if err != nil {
		return err
	}
	dirPath := server.path() + "/versions/" + largest
	execName := server.config.Name + "." + platform
	execPath := dirPath + "/" + execName

	server.port, err = Freeport()
	if err != nil {
		return err
	}

	cmd := exec.Cmd{
		Path:   execPath,
		Dir:    dirPath,
		Stdout: os.Stdout,
		Stderr: os.Stdout,
		Args:   []string{execName, "server", "-p", fmt.Sprintf("%d", server.port)},
	}

	err = cmd.Start()
	if err != nil {
		panic(err)
	}

	//fmt.Println(dirPath, executablePath)
	return nil
}

func (server Server) path() string {
	return os.Getenv("HOME") + "/." + server.config.Name
}

func (server Server) GetLatestVersion() (string, error) {
	files, err := ioutil.ReadDir(server.path() + "/versions")
	if err != nil {
		return "", err
	}

	var versions []version
	for _, v := range files {
		version := newVersion(v.Name(), server.config.Name)
		if version != nil {
			versions = append(versions, *version)
		}
	}

	if len(versions) == 0 {
		return "", errors.New("no versions set")
	}

	largest := versions[0]
	for _, v := range versions {
		if v.id() > largest.id() {
			largest = v
		}
	}

	return largest.human(), nil
}

type version struct {
	prefix string
	major  int
	minor  int
	patch  int
}

func (v version) human() string {
	return fmt.Sprintf("%s.%d.%d.%d", v.prefix, v.major, v.minor, v.patch)
}

func (v version) id() int {
	return (1000000 * v.major) + (1000 * v.minor) + v.patch
}

func newVersion(versionName, prefix string) *version {
	splited := strings.Split(versionName, ".")
	if len(splited) != 4 {
		return nil
	}

	if splited[0] != prefix {
		return nil
	}

	major, err := strconv.Atoi(splited[1])
	if err != nil {
		return nil
	}

	minor, err := strconv.Atoi(splited[2])
	if err != nil {
		return nil
	}

	patch, err := strconv.Atoi(splited[3])
	if err != nil {
		return nil
	}

	return &version{prefix, major, minor, patch}
}
