// Copyright © 2018 CloudODM Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"github.com/OpenDroneMap/CloudODM/internal/fs"
	"github.com/OpenDroneMap/CloudODM/internal/logger"
	"github.com/OpenDroneMap/CloudODM/internal/odm"

	homedir "github.com/mitchellh/go-homedir"
)

// NewConfiguration creates a new configuration from a specified file path
func NewConfiguration(filePath string) Configuration {
	conf := Configuration{}
	conf.Nodes = map[string]odm.Node{}
	conf.filePath = filePath
	return conf
}

// Save saves the configuration to file
func (c Configuration) Save() {
	saveToFile(c, c.filePath)
}

// Configuration is a collection of config values
type Configuration struct {
	Nodes map[string]odm.Node `json:"nodes"`

	filePath string
}

// Initialize the configuration
func Initialize() Configuration {
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cfgPath := filepath.Join(home, ".odm.json")
	user := Configuration{}

	if exists, _ := fs.FileExists(cfgPath); exists {
		// Read existing config
		user = loadFromFile(cfgPath)
	} else {
		// Download public nodes, choose a default
		nodes := GetPublicNodes()
		user = NewConfiguration(cfgPath)

		logger.Info("Found " + strconv.Itoa(len(nodes)) + " public nodes")

		if len(nodes) > 0 {
			logger.Info("Picking a random one...")

			randomNode := nodes[rand.Intn(len(nodes))]

			logger.Info("Setting default node to " + randomNode.String())
			user.AddNode("default", randomNode.Url)

			logger.Info("Initialized configuration at " + cfgPath)
		}
	}

	return user
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

	logger.Debug("Wrote configuration to " + filePath)
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

// AddNode adds a new node to the configuration
func (c Configuration) AddNode(name string, nodeURL string) error {
	if _, ok := c.Nodes[name]; ok {
		return errors.New("node" + name + " already exists. Remove it first.")
	}

	u, err := url.ParseRequestURI(nodeURL)
	if err != nil {
		return errors.New(nodeURL + " is not a valid URL. A valid URL looks like: http://hostname:port/?token=optional")
	}

	c.Nodes[name] = odm.Node{URL: u.Scheme + "://" + u.Host, Token: u.Query().Get("token")}
	c.Save()

	return nil
}

// RemoveNode removes a node from the configuration
func (c Configuration) RemoveNode(name string) bool {
	_, ok := c.Nodes[name]
	if ok {
		delete(c.Nodes, name)
		c.Save()
	}
	return ok
}

// GetNode gets a Node instance given its name
func (c Configuration) GetNode(name string) (*odm.Node, error) {
	if len(c.Nodes) == 0 {
		return nil, errors.New("No nodes. Add one with ./odm node")
	}

	node, ok := c.Nodes[name]
	if !ok {
		return nil, errors.New("node: " + name + " does not exist")
	}

	return &node, nil
}

func (c Configuration) UpdateNode(name string, node odm.Node) {
	c.Nodes[name] = node
	c.Save()
}
