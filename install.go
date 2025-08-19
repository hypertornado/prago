package prago

import (
	"bufio"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"
)

func (app *App) autoInstallDatabase() error {
	fmt.Printf("Would you like to install database for %s? (yes or no)\n", app.codeName)
	reader := bufio.NewReader(os.Stdin)
	confirm, _ := reader.ReadString('\n')
	if confirm != "yes\n" {
		return errors.New("Declined to install database")
	}

	fmt.Println("Root mysql password:")
	rootPassword, _ := reader.ReadString('\n')
	rootPassword = strings.TrimRight(rootPassword, "\n")

	db, err := connectMysql("root", rootPassword, "")
	must(err)
	defer db.Close()

	mysqlCodeName := app.codeName
	mysqlCodeName = strings.ReplaceAll(mysqlCodeName, ".", "")
	mysqlCodeName = strings.ReplaceAll(mysqlCodeName, "-", "")
	mysqlCodeName = strings.ReplaceAll(mysqlCodeName, "_", "")

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s CHARACTER SET utf8 DEFAULT COLLATE utf8_unicode_ci;", mysqlCodeName))
	if err != nil {
		return err
	}

	password, err := generateRandomPassword(12)
	must(err)

	_, err = db.Exec(fmt.Sprintf("CREATE USER '%s'@'localhost' IDENTIFIED BY '%s';", mysqlCodeName, password))
	if err != nil {
		return err
	}

	_, err = db.Exec(fmt.Sprintf("GRANT ALL ON %s.* TO '%s'@'localhost';", mysqlCodeName, mysqlCodeName))
	if err != nil {
		return err
	}

	_, err = db.Exec("FLUSH PRIVILEGES;")
	if err != nil {
		return err
	}

	config := dbConnectConfig{
		Name:     mysqlCodeName,
		User:     mysqlCodeName,
		Password: password,
	}

	configData, err := json.Marshal(config)
	if err != nil {
		return err
	}

	err = os.MkdirAll(getAppDotPath(app.codeName), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create folder: %v", err)
	}

	filePath := getDBConnectPath(app.codeName)

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating the file: %s", err)
	}
	defer file.Close()

	_, err = file.Write(configData)
	if err != nil {
		return fmt.Errorf("error writing to the file: %s", err)
	}

	fmt.Printf("Database '%s' created successfully\n", mysqlCodeName)
	return nil

}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func generateRandomPassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var password []byte

	for i := 0; i < length; i++ {
		randomInt, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		password = append(password, charset[randomInt.Int64()])
	}

	return string(password) + "a1!", nil
}
