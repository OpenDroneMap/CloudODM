package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/OpenDroneMap/CloudODM/internal/logger"
)

type PublicNode struct {
	Url        string `json:"url"`
	Maintainer string `json:"maintainer"`
	Company    string `json:"company"`
	Website    string `json:"website"`
}

func (n PublicNode) String() string {
	return fmt.Sprintf("%s", n.Url)
}

func GetPublicNodes() []PublicNode {
	logger.Debug("Retrieving public nodes...")
	nodes := []PublicNode{}

	resp, err := http.Get("https://raw.githubusercontent.com/OpenDroneMap/CloudODM/master/public_nodes.json")
	if err != nil {
		logger.Info(err)
		return nodes
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Info(err)
		return nodes
	}

	if err := json.Unmarshal(body, &nodes); err != nil {
		logger.Info("Invalid JSON content: " + string(body))
		return nodes
	}

	return nodes
}
