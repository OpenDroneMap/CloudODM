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

package cmd

import (
	"fmt"

	"github.com/OpenDroneMap/CloudODM/internal/config"
	"github.com/OpenDroneMap/CloudODM/internal/logger"
	"github.com/spf13/cobra"
)

var argsCmd = &cobra.Command{
	Use:     "args",
	Aliases: []string{"arguments"},
	Short:   "View arguments",
	Run: func(cmd *cobra.Command, args []string) {
		user := config.Initialize()

		user.CheckLogin(nodeName, "", "")

		node, err := user.GetNode(nodeName)
		if err != nil {
			logger.Error(err)
		}

		options, err := node.Options()
		if err != nil {
			logger.Error(err)
		}

		logger.Info("Args:")
		logger.Info("")

		for _, option := range options {
			domain := ""
			switch option.Domain.(type) {
			case string:
				domain = option.Domain.(string)
			case []interface{}:
				for i, v := range option.Domain.([]interface{}) {
					if i > 0 {
						domain += ","
					}
					if v == "" {
						v = "\"\""
					}
					domain += fmt.Sprint(v)
				}
			default:
				domain = "?"
			}

			if domain != "" {
				domain = "<" + domain + ">"
			}

			logger.Info("--" + option.Name + " " + domain)
			logger.Info(option.Help)
			logger.Info("")
		}
	},
}

func init() {
	argsCmd.Flags().StringVarP(&nodeName, "node", "n", "default", "Processing node to use")

	rootCmd.AddCommand(argsCmd)
}
