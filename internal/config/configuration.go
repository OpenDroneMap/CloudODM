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

// User contains the user's configuration
var User Configuration

// NewConfiguration creates a new configuration from a specified file path
func NewConfiguration(filePath string) Configuration {
	conf := Configuration{}
	conf.Nodes = map[string]Node{}
	conf.filePath = filePath
	return conf
}

func (c Configuration) Save() {
	saveToFile(c, c.filePath)
}

type Configuration struct {
	Nodes map[string]Node `json:"nodes"`

	filePath string
}

type Node struct {
	Url   string `json:"url"`
	Token string `json:"token"`
}

func (n Node) String() string {
	return n.Url
}

// Initialize the configuration
func Initialize() {
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cfgPath := filepath.Join(home, ".odm.json")

	if exists, _ := fs.FileExists(cfgPath); exists {
		// Read existing config
		User = loadFromFile(cfgPath)
	} else {
		// Download public nodes, choose a default
		nodes := GetPublicNodes()
		User := NewConfiguration(cfgPath)

		logger.Info("Found " + strconv.Itoa(len(nodes)) + " public nodes")

		if len(nodes) > 0 {
			logger.Info("Picking a random one...")

			randomNode := nodes[rand.Intn(len(nodes))]

			logger.Info("Setting default node to " + randomNode.String())
			User.AddNode("default", randomNode.Url)

			logger.Info("Initialized configuration at " + cfgPath)
		}
	}
}

func saveToFile(conf Configuration, filePath string) {
	jsonData, err := json.MarshalIndent(conf, "", " ")
	if err != nil {
		logger.Error(err)
	}

	jsonFile, err := os.Create(filePath)
	if err != nil {
		logger.Error(err)
	}
	defer jsonFile.Close()

	jsonFile.Write(jsonData)

	logger.Debug("Wrote default configuration in " + filePath)
}

func loadFromFile(filePath string) Configuration {
	jsonFile, err := os.Open(filePath)
	if err != nil {
		logger.Error(err)
	}
	logger.Debug("Loaded configuration from " + filePath)

	defer jsonFile.Close()

	jsonData, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		logger.Error("Cannot read configuration file: " + filePath)
	}

	conf := NewConfiguration(filePath)
	err = json.Unmarshal([]byte(jsonData), &conf)
	if err != nil {
		logger.Error("Cannot parse configuration file: " + err.Error())
	}

	return conf
}
