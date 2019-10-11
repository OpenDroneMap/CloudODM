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
	"mime"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/OpenDroneMap/CloudODM/internal/config"
	"github.com/OpenDroneMap/CloudODM/internal/fs"
	"github.com/OpenDroneMap/CloudODM/internal/logger"
	"github.com/OpenDroneMap/CloudODM/internal/odm"

	"github.com/spf13/cobra"
)

var outputPath string
var nodeName string
var force bool

var rootCmd = &cobra.Command{
	Use:     "odm [flags] <images> [<gcp>] [args]",
	Short:   "A command line tool to process aerial imagery in the cloud",
	Version: "1.1.0",
	Run: func(cmd *cobra.Command, args []string) {
		user := config.Initialize()
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}

		// Check output directory
		filesCount, err := fs.DirectoryFilesCount(outputPath)
		if err != nil {
			logger.Error(err)
		}
		if filesCount > 0 && !force {
			logger.Error(outputPath + " already exists (pass --force to override directory contents)")
		}

		inputFiles, options := parseArgs(args)
		inputFiles = filterImagesAndText(inputFiles)

		logger.Verbose("Input Files (" + strconv.Itoa(len(inputFiles)) + ")")
		for _, file := range inputFiles {
			logger.Debug(" * " + file)
		}

		logger.Debug("Options: " + strings.Join(options, " "))

		info := user.CheckLogin(nodeName, "", "")

		// Check max images
		if len(inputFiles) > info.MaxImages {
			logger.Error("Cannot process", len(inputFiles), "files with this node, the node has a limit of", info.MaxImages)
		}

		logger.Debug("NodeODM version: " + info.Version)

		node, err := user.GetNode(nodeName)
		if err != nil {
			logger.Error(err)
		}

		nodeOptions, err := node.Options()
		if err != nil {
			logger.Error(err)
		}

		// Create output directory
		if !fs.IsDirectory(outputPath) {
			err = os.MkdirAll(outputPath, 0755)
			if err != nil {
				logger.Error(err)
			}
		}

		odm.Run(inputFiles, parseOptions(options, nodeOptions), *node, outputPath)
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
	rootCmd.PersistentFlags().BoolVarP(&logger.QuietFlag, "quiet", "q", false, "suppress output")

	rootCmd.Flags().BoolVarP(&force, "force", "f", false, "replace the contents of the output directory if it already exists")
	rootCmd.Flags().StringVarP(&outputPath, "output", "o", "./output", "directory where to store processing results")
	rootCmd.Flags().StringVarP(&nodeName, "node", "n", "default", "Processing node to use")

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
			inputFiles = append(inputFiles, arg)
		} else {
			options = append(options, arg)
		}
	}

	return inputFiles, options
}

func filterImagesAndText(files []string) []string {
	var result []string

	for _, file := range files {
		mimeType := mime.TypeByExtension(filepath.Ext(file))
		if strings.HasPrefix(mimeType, "image") || strings.HasPrefix(mimeType, "text") {
			result = append(result, file)
		}
	}

	return result
}

func invalidArg(arg string) {
	logger.Error("Invalid argument " + arg + ". See ./odm args for a list of valid arguments.")
}

func parseOptions(options []string, nodeOptions []odm.OptionResponse) []odm.Option {
	result := []odm.Option{}

	for i := 0; i < len(options); i++ {
		o := options[i]

		if strings.HasPrefix(o, "--") || strings.HasPrefix(o, "-") {
			currentOption := odm.Option{}

			// Key
			o = strings.TrimPrefix(o, "--")
			o = strings.TrimPrefix(o, "-")

			found := false
			optType := "string"
			for _, no := range nodeOptions {
				if no.Name == o {
					found = true
					optType = no.Type
					break
				}
			}

			if !found {
				invalidArg(o)
			}

			// TODO: domain checks

			currentOption.Name = o
			if optType == "bool" {
				currentOption.Value = "true"
			} else {
				if i < len(options)-1 {
					currentOption.Value = options[i+1]
					i++
				} else {
					invalidArg(o)
				}
			}

			result = append(result, currentOption)
		} else {
			invalidArg(o)
		}
	}

	return result
}
