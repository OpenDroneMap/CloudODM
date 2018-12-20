package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"

	"github.com/OpenDroneMap/CloudODM/internal/fs"
	"github.com/OpenDroneMap/CloudODM/internal/logger"

	homedir "github.com/mitchellh/go-homedir"
)

var cfgPath string

func NewConfiguration() Configuration {
	conf := Configuration{}
	conf.Nodes = map[string]Node{}
	return conf
}

type Configuration struct {
	Nodes map[string]Node `json:"nodes"`
}

type Node struct {
	Url   string `json:"url"`
	Token string `json:"token"`
}

// Initialize the configuration
func Initialize() {
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cfgPath = filepath.Join(home, ".odm.json")
	if exists, _ := fs.FileExists(cfgPath); exists {
		// Read existing config
		loadFromFile()
	} else {
		// Download public nodes, choose a default
		nodes := GetPublicNodes()
		defaultConfig := NewConfiguration()

		logger.Info("Found " + strconv.Itoa(len(nodes)) + " public nodes")

		if len(nodes) > 0 {
			logger.Info("Picking a random one...")

			randomNode := nodes[rand.Intn(len(nodes))]

			logger.Info("Setting default node to " + randomNode.String())
			defaultConfig.Nodes["default"] = Node{Url: randomNode.Url, Token: ""}
		}

		saveToFile(defaultConfig)
	}
}

func saveToFile(conf Configuration) {
	jsonData, err := json.MarshalIndent(conf, "", " ")
	if err != nil {
		logger.Error(err)
	}

	jsonFile, err := os.Create(cfgPath)
	if err != nil {
		logger.Error(err)
	}
	defer jsonFile.Close()

	jsonFile.Write(jsonData)

	logger.Info("Wrote default configuration in " + cfgPath)
}

func loadFromFile() Configuration {
	jsonFile, err := os.Open(cfgPath)
	if err != nil {
		logger.Error(err)
	}
	logger.Debug("Loaded configuration from " + cfgPath)

	defer jsonFile.Close()

	jsonData, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		logger.Error("Cannot read configuration file: " + cfgPath)
	}

	conf := Configuration{}
	err = json.Unmarshal([]byte(jsonData), &conf)
	if err != nil {
		logger.Error("Cannot parse configuration file: " + err.Error())
	}

	return conf
}
