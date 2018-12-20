package config

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type PublicNode struct {
	Url        string `json:"url"`
	Maintainer string `json:"maintainer"`
	Company    string `json:"company"`
	Website    string `json:"website"`
}

func GetPublicNodes() []PublicNode {
	resp, err := http.Get("https://raw.githubusercontent.com/OpenDroneMap/CloudODM/master/public_nodes.json")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	nodes := []PublicNode{}
	json.Unmarshal(body, &nodes)

	return nil
}
