package configs

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type configuration struct {
	MS       MicroService `json:"microservice"`
	RootPath string
	DBPath   string `json:"dbpath"`
}

type MicroService struct {
	Title     string   `json:"title"`
	Domain    string   `json:"domain"`
	URL       []string `json:"url"`
	Addr      string   `json:"addr"`
	Timeout   string   `json:"timeout"`
	Heartbeat string   `json:"heartbeat"`
}

var Data = &configuration{}

func setRootPath() error {
	var root string
	var err error
	if strings.Contains(os.Args[0], ".test") {
		root = "../../" // for test dbmanager
	} else {
		root, err = os.Getwd()
		if err != nil {
			log.Printf("config Getwd: %#v", err)
		}
	}
	Data.RootPath = root
	return nil
}

func get() error {
	f, err := os.ReadFile(filepath.Join(Data.RootPath, "configs/configs.json"))
	if err != nil {
		return err
	}
	return json.Unmarshal(f, Data)
}
func init() {
	if err := setRootPath(); err != nil {
		log.Printf("config init error: %v", err)
	}
	if err := get(); err != nil {
		log.Printf("config get() error: %v", err)
	}
}

// Reset is for test to reset RootPath and invoke get()
func Reset(pwd string) error {
	Data.RootPath = pwd
	return get()
}
