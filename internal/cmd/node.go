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
	"github.com/OpenDroneMap/CloudODM/internal/logger"

	"github.com/OpenDroneMap/CloudODM/internal/config"
	"github.com/spf13/cobra"
)

// nodeCmd represents the node command
var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Manage processing nodes",
	Run: func(cmd *cobra.Command, args []string) {
		config.Initialize()

		for k, n := range config.User.Nodes {
			if logger.Verbose {
				logger.Info(k + " - " + n.String())
			} else {
				logger.Info(k)
			}
		}
	},
}

var addCmd = &cobra.Command{
	Use:   "add <name> <url>",
	Short: "Add a new processing node",
	Args:  cobra.ExactValidArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		config.Initialize()

		if err := config.User.AddNode(args[0], args[1]); err != nil {
			logger.Error(err)
		}
	},
}

var removeCmd = &cobra.Command{
	Use:     "remove <name>",
	Short:   "Remove a processing node",
	Aliases: []string{"delete", "rm", "del"},
	Args:    cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		config.Initialize()

		if !config.User.RemoveNode(args[0]) {
			logger.Error("Cannot remove node " + args[0] + " (does it exist?)")
		}
	},
}

func init() {
	nodeCmd.AddCommand(addCmd)
	nodeCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(nodeCmd)
}
