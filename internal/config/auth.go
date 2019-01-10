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
