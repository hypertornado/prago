package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path"
	"strings"

	"github.com/hypertornado/prago/utils"

	_ "github.com/go-sql-driver/mysql" //use mysql
)

func main() {
	fmt.Println("Prago project installer")

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	fmt.Println("Installation directory:", wd)

	_, projectName := path.Split(wd)

	fmt.Println("Project name:", projectName)

	createSkeleton(wd, projectName)
	createConfigFiles(projectName)

}

func createDirectory(path string) {
	fmt.Println("Creating directory:", path)
	err := os.Mkdir(path, 0755)

	if err != nil {
		if os.IsExist(err) {
			fmt.Println("Directory already exists.")
		} else {
			panic(err)
		}
	}
}

func createConfigFiles(projectName string) {
	if !utils.ConsoleQuestion("Do you want to create app config files?") {
		return
	}

	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	dotPath := path.Join(user.HomeDir, "."+projectName)
	createDirectory(dotPath)

	configPath := path.Join(dotPath, "config.json")

	/*f, err := os.OpenFile(configPath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}*/

	conf := Config{
		StaticPaths: []string{"public"},
	}

	mysqlRootPassword := getValue("MySQL root password:", "")
	conf.DBUser = projectName
	conf.DBName = projectName
	conf.DBPassword = createDatabase(mysqlRootPassword, projectName)

	conf.SSH = getValue("SSH path:", "")

	conf.BaseURL = getValue("BaseURL:", "http://localhost:8585")
	conf.Random = randomPassword()

	conf.SendgridAPI = getValue("SendgridAPI:", "")
	conf.NoReplyEmail = getValue("NoReplyEmail:", "noreply@"+projectName+".com")

	conf.Google = getValue("Google API:", "")

	conf.CDNAccount = getValue("CDNAccount name:", projectName)
	conf.CDNPassword = getValue("CDNAccount password:", "")

	marshaledConf, err := json.MarshalIndent(conf, "", " ")
	if err != nil {
		panic(err)
	}
	createFile(configPath, string(marshaledConf))
}

type Config struct {
	StaticPaths []string `json:"staticPaths"`
	SSH         string   `json:"ssh"`

	DBUser     string `json:"dbUser"`
	DBName     string `json:"dbName"`
	DBPassword string `json:"dbPassword"`

	BaseURL string `json:"baseUrl"`
	Random  string `json:"random"`

	SendgridAPI  string `json:"sendgridApi"`
	NoReplyEmail string `json:"noReplyEmail"`

	Google string `json:"google"`

	CDNAccount  string `json:"cdnAccount"`
	CDNPassword string `json:"cdnPassword"`
}

func getValue(question, defaultValue string) string {
	q := question
	if defaultValue != "" {
		q += "(default: " + defaultValue + ")"
	}
	fmt.Println(q)
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	text = strings.ReplaceAll(text, "\n", "")

	if text == "" {
		return defaultValue
	}
	return text
}

func randomPassword() string {
	return utils.RandomString(16)
}

func createDatabase(rootPassword, projectName string) string {

	connectString := fmt.Sprintf("root:%s@/?charset=utf8&parseTime=True&loc=Local", rootPassword)
	fmt.Println("Connecting to MySQL database with string:", connectString)
	db, err := sql.Open("mysql", connectString)

	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Connected to MySQL database as root.")

	userPassword := randomPassword()

	fmt.Println("Droping user if exists.")
	_, err = db.Exec(fmt.Sprintf("drop user if exists '%s'@'localhost';", projectName))
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Droping database if exists.")
	_, err = db.Exec(fmt.Sprintf("drop database if exists %s;", projectName))
	if err != nil {
		fmt.Println(err)
	}

	queries := []string{
		fmt.Sprintf("CREATE USER '%s'@'localhost' IDENTIFIED BY '%s';", projectName, userPassword),
		fmt.Sprintf("CREATE DATABASE %s CHARACTER SET utf8 DEFAULT COLLATE utf8_unicode_ci;", projectName),
		fmt.Sprintf("GRANT ALL ON %s.* TO '%s'@'localhost';", projectName, projectName),
		fmt.Sprintf("FLUSH PRIVILEGES;"),
	}

	for _, q := range queries {
		fmt.Println("Executing:", q)
		_, err := db.Exec(q)
		if err != nil {
			panic(err)
		}
	}
	return userPassword
}
