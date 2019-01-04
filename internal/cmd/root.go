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
	"path/filepath"
	"strconv"
	"strings"

	"github.com/OpenDroneMap/CloudODM/internal/config"
	"github.com/OpenDroneMap/CloudODM/internal/fs"
	"github.com/OpenDroneMap/CloudODM/internal/logger"

	"github.com/spf13/cobra"
)

var outputPath string

var rootCmd = &cobra.Command{
	Use:   "odm [flags] <images> [<gcp>] [parameters]",
	Short: "A command line tool to process aerial imagery in the cloud",

	Run: func(cmd *cobra.Command, args []string) {
		config.Initialize()
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}

		inputFiles, options := parseArgs(args)

		logger.Verbose("Input Files (" + strconv.Itoa(len(inputFiles)) + ")")
		for _, file := range inputFiles {
			logger.Debug(" * " + file)
		}

		logger.Debug("Options: " + strings.Join(options, " "))
	},

	TraverseChildren: true,

	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.Error(err)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&logger.VerboseFlag, "verbose", "v", false, "show verbose output")
	rootCmd.PersistentFlags().BoolVarP(&logger.DebugFlag, "debug", "d", false, "show debug output")
	rootCmd.Flags().StringVarP(&outputPath, "output", "o", "./output", "directory where to store processing results")
	rootCmd.Flags().SetInterspersed(false)
}

func parseArgs(args []string) ([]string, []string) {
	var inputFiles []string
	var options []string

	for _, arg := range args {
		if fs.IsDirectory(arg) {
			// Add everything from directory
			globPaths, err := filepath.Glob(arg + "/*")
			if err != nil {
				logger.Error(err)
			}

			for _, globPath := range globPaths {
				if fs.IsFile(globPath) {
					inputFiles = append(inputFiles, globPath)
				}
			}
		} else if fs.IsFile(arg) {
			fmt.Printf(arg)
			inputFiles = append(inputFiles, arg)
		} else {
			options = append(options, arg)
		}
	}

	return inputFiles, options
}
