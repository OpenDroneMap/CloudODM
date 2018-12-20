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
	"os"

	"github.com/OpenDroneMap/CloudODM/internal/config"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Process a dataset",
	Run: func(cmd *cobra.Command, args []string) {
		config.Initialize()

		fmt.Println("run called")
		fmt.Println(os.Args[2:])
		fmt.Println(args)
	},

	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
