package config

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/OpenDroneMap/CloudODM/internal/logger"
)

type PublicNode struct {
	Url        string `json:"url"`
	Maintainer string `json:"maintainer"`
	Company    string `json:"company"`
	Website    string `json:"website"`
}

func GetPublicNodes() []PublicNode {
	logger.Debug("Retrieving public nodes...")

	resp, err := http.Get("https://raw.githubusercontent.com/OpenDroneMap/CloudODM/master/public_nodes.json")
	if err != nil {
		logger.Info(err)
		return nil
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Info(err)
		return nil
	}

	nodes := []PublicNode{}
	if err := json.Unmarshal(body, &nodes); err != nil {
		logger.Info("Invalid JSON content: " + string(body))
		return nil
	}

	logger.Info("Loaded " + strconv.Itoa(len(nodes)) + " public nodes")
	return nodes
}
