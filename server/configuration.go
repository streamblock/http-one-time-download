package main

import (
	"os"

	"gopkg.in/yaml.v2"
)

type configuration struct {
	Server struct {
		Port 				string 	`yaml:"port"`
		Host 				string 	`yaml:"host"`
		MaxUploadSize 		int64 	`yaml:"max_size_upload_bytes"`
		UploadPath 			string 	`yaml:"upload_path"`

		DropPrivileges struct {
			Enable 			bool 	`yaml:"enable"`
			UserName 		string 	`yaml:"user_name"`
		} `yaml:"drop_privileges"`

		RegularUpload struct {
			Enable bool   `yaml:"enable"`
			URL    string `yaml:"url"`
		} `yaml:"regular_upload"`

		SecureUpload struct {
			Enable       bool   `yaml:"enable"`
			URL          string `yaml:"url"`
			SubDirLength int64  `yaml:"tmp_dir_length"`
		} `yaml:"secure_upload"`

		Download struct {
			URL string `yaml:"url"`
		} `yaml:"download"`

	} `yaml:"server"`

	Logging struct {
		LogEveryRequest 	bool 	`yaml:"log_every_request"`
	} `yaml:"logging"`
}

func newConfiguration(configPath string) (*configuration, error) {
	config := &configuration{}

	fConfig, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer fConfig.Close()

	decoder := yaml.NewDecoder(fConfig)
	err = decoder.Decode(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}