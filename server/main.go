package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

var maxUploadSize int64
var uploadPath string
var uploadSubDirLength int64
var downloadURLPattern string


func isValidPath(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		return err
	}
	if s.IsDir() {
		return fmt.Errorf("'%s' is a directory", path)
	}
	return nil
}

func parseFlags() (string, error) {
	// -config
	var configPath string
	flag.StringVar(&configPath, "config", "./config.yml", "path to config file")

	// Parse CLI
	flag.Parse()

	if err := isValidPath(configPath); err != nil {
		return "", err
	}
	return configPath, nil
}

func main() {
	configPath, err := parseFlags()
	if err != nil {
		log.Fatal(err)
	}
	configuration, err := newConfiguration(configPath)
	if err != nil {
		log.Fatal(err)
	}

	// set global variable
	maxUploadSize = configuration.Server.MaxUploadSize
	uploadPath = configuration.Server.UploadPath
	uploadSubDirLength = configuration.Server.SecureUpload.SubDirLength
	downloadURLPattern = configuration.Server.Download.URL

	// Drop privileges?
	if (configuration.Server.DropPrivileges.Enable) {
		err := chowmPath(uploadPath, configuration.Server.DropPrivileges.UserName)
		if err != nil {
			log.Fatal(err)
		}

		err = dropPrivileges(configuration.Server.DropPrivileges.UserName)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println("[+] server address        = " + configuration.Server.Host)
	log.Println("[+] server port           = " + configuration.Server.Port)
	log.Println("[+] server upload dir     = " + configuration.Server.UploadPath)
	log.Println("")
	log.Println("[+] server download URL   = " + configuration.Server.Download.URL)
	log.Println("")
	log.Println("[+] server upload         = " + strconv.FormatBool(configuration.Server.RegularUpload.Enable))
	log.Println("[+] server upload URL     = " + configuration.Server.RegularUpload.URL)
	log.Println("")
	log.Println("[+] server sec upload     = " + strconv.FormatBool(configuration.Server.SecureUpload.Enable))
	log.Println("[+] server sec upload URL = " + configuration.Server.SecureUpload.URL)

	mux := newMux(configuration)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
