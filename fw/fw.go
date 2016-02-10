package fw

import (
	"encoding/json"
	"github.com/hypertornado/lazne/server"
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

func FW() {
	app := kingpin.New("lazne", "Lazne")

	serverCommand := app.Command("server", "Run server")
	port := serverCommand.Flag("port", "server port").Default("8585").Short('p').Int()
	developmentMode := serverCommand.Flag("development", "Is in development mode").Default("false").Short('d').Bool()
	configPath := serverCommand.Flag("config", "Path for config file").Default("").String()

	buildCommand := app.Command("build", "Build version")

	cssCommand := app.Command("css", "Build CSS")
	devCommand := app.Command("dev", "Development")

	command, err := app.Parse(os.Args[1:])
	if err != nil {
		panic(err)
	}

	config := parseConfig(*configPath)

	switch command {
	case serverCommand.FullCommand():
		server.Start(*port, *developmentMode, config)
	case buildCommand.FullCommand():
		build()
	case cssCommand.FullCommand():
		compileCss()
	case devCommand.FullCommand():
		development(config)
	}
}

func build() {
	println("not implemented")
}

func development(conf map[string]string) {
	go developmentCSS()
	server.Start(8585, true, conf)
}

func parseConfig(path string) map[string]string {
	if path == "" {
		path = os.Getenv("HOME") + "/.lazne/config.json"
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	kv := make(map[string]string)

	err = json.Unmarshal(data, &kv)
	if err != nil {
		panic(err)
	}

	return kv
}

func compileCss() error {
	outfile, err := os.Create("public/compiled.css")
	if err != nil {
		return err
	}
	defer outfile.Close()

	return commandHelper(exec.Command("lessc", "public/css/index.less"), outfile)
}

func commandHelper(cmd *exec.Cmd, out io.Writer) error {
	var err error
	cmd.Stdout = out
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}
