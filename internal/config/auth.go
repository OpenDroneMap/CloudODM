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
	"github.com/OpenDroneMap/CloudODM/internal/logger"
	"github.com/OpenDroneMap/CloudODM/internal/odm"
)

// CheckLogin checks if the node needs login
// if it does, it attempts to login
// it it doesn't, returns node.Info()
// on error, it prints a message and exits
func CheckLogin(nodeName string, username string, password string) *odm.InfoResponse {
	node, err := User.GetNode(nodeName)
	if err != nil {
		logger.Error(err)
	}

	info, err := node.Info()
	err = node.CheckAuthentication(err)
	if err != nil {
		if err == odm.ErrAuthRequired {
			token, err := node.TryLogin(username, password)
			if err != nil {
				logger.Error(err)
			}

			// Validate token
			node.Token = token
			info, err = node.Info()
			err = node.CheckAuthentication(err)
			if err != nil {
				logger.Error(err)
			}

			User.UpdateNode(nodeName, *node)
		} else {
			logger.Error(err)
		}
	}

	return info
}
