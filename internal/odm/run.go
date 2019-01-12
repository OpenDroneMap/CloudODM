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

package odm

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/OpenDroneMap/CloudODM/internal/fs"

	"github.com/OpenDroneMap/CloudODM/internal/logger"

	"github.com/cheggaaa/pb"
)

type TaskNewResponse struct {
	UUID  string `json:"uuid"`
	Error string `json:"error"`
}

// Run processes a dataset
func Run(files []string, options []Option, node Node, outputPath string) {
	var err error
	var bar *pb.ProgressBar

	var f *os.File
	var fi os.FileInfo

	var totalBytes int64

	showProgress := !logger.QuietFlag

	// Calculate total upload size
	for _, file := range files {
		if fi, err = os.Stat(file); err != nil {
			logger.Error(err)
		}
		totalBytes += fi.Size()
		f.Close()
	}

	if showProgress {
		bar = pb.New64(totalBytes).SetUnits(pb.U_BYTES).SetRefreshRate(time.Millisecond * 10)
		bar.Start()
	}

	// Convert options to JSON
	jsonOptions, err := json.Marshal(options)
	if err != nil {
		logger.Error(err)
	}

	// Setup pipe
	r, w := io.Pipe()
	mpw := multipart.NewWriter(w)

	// Pipe work, stream file contents
	go func() {
		var part io.Writer
		defer w.Close()
		defer f.Close()

		for _, file := range files {
			if f, err = os.Open(file); err != nil {
				logger.Error(err)
			}
			if fi, err = f.Stat(); err != nil {
				logger.Error(err)
			}

			if part, err = mpw.CreateFormFile("images", fi.Name()); err != nil {
				logger.Error(err)
			}

			if showProgress {
				bar.Prefix("[" + fi.Name() + "]")
			}

			part = io.MultiWriter(part, bar)

			if _, err = io.Copy(part, f); err != nil {
				logger.Error(err)
			}
		}

		mpw.WriteField("skipPostProcessing", "true")
		mpw.WriteField("options", string(jsonOptions))

		if err = mpw.Close(); err != nil {
			logger.Error(err)
		}
	}()

	resp, err := http.Post(node.URLFor("/task/new"), mpw.FormDataContentType(), r)
	if err != nil {
		logger.Error(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err)
	}

	res := TaskNewResponse{}
	if err := json.Unmarshal(body, &res); err != nil {
		logger.Error(err)
	}

	// Handle error response from API
	if res.Error != "" {
		logger.Error(res.Error)
	}

	if showProgress {
		bar.Finish()
	}

	// We should have a UUID
	uuid := res.UUID
	logger.Info("Task UUID: " + uuid)

	info, err := node.TaskInfo(uuid)
	if err != nil {
		logger.Error(err)
	}

	// Catch CTRL+C
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c

		logger.Info("Canceling task...")

		// Attempt to cancel task
		retryCount := 0
		retryLimit := 5

		for retryCount < retryLimit {
			if err := node.TaskCancel(uuid); err != nil {
				retryCount++
				logger.Info(err)
				time.Sleep(1 * time.Second)
			} else {
				break
			}
		}

		filesCount, err := fs.DirectoryFilesCount(outputPath)
		if err == nil && fs.IsDirectory(outputPath) && filesCount == 0 {
			os.Remove(outputPath)
		}

		os.Exit(1)
	}()

	// Start listening for output and task updates...
	status := info.Status.Code
	lineNum := 0

	for status == STATUS_QUEUED || status == STATUS_RUNNING {
		time.Sleep(3 * time.Second)

		info, err := node.TaskInfo(uuid)
		if err != nil {
			logger.Info(err)

			// Log error, try again later
			continue
		}

		status = info.Status.Code

		lines, err := node.TaskOutput(uuid, lineNum)
		if err != nil {
			logger.Info(err)
			continue
		}

		for _, line := range lines {
			logger.Info(line)
		}
		lineNum += len(lines)
	}

	if status == STATUS_CANCELED || status == STATUS_FAILED {
		logger.Error("Task failed or canceled")
	}

	if status == STATUS_COMPLETED {
		retryCount := 0
		retryLimit := 10

		archiveDst := path.Join(outputPath, "all.zip")
		logger.Info("Task completed! Downloading and extracting results...")
		logger.Info("")

		for {
			err := node.TaskDownload(uuid, "all.zip", archiveDst)
			if err == nil {
				break
			} else {
				logger.Info("Error downloading file (" + err.Error() + ") retrying in " + string(3*retryLimit) + " seconds...")
				time.Sleep(time.Duration(3*retryLimit) * time.Second)
				retryCount++
				if retryCount >= retryLimit {
					logger.Error("Download retries limit exceeded (" + string(retryLimit) + "), exiting...")
				}
			}
		}

		// Unzip
		_, err := fs.Unzip(archiveDst, outputPath)
		if err != nil {
			logger.Error(err)
		}

		// Remove
		if err := os.Remove(archiveDst); err != nil {
			logger.Info(err)
		}

		logger.Info("Done!")
	}
}
