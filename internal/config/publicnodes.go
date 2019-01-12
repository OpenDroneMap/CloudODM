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
