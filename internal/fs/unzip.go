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

// Based on work published on https://golangcode.com/download-a-file-with-progress/

package fs

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/OpenDroneMap/CloudODM/internal/logger"
	"github.com/cheggaaa/pb"
)

// Unzip decompresses a zip archive
func Unzip(zipFilePath string, destDirectory string) ([]string, error) {
	var bar *pb.ProgressBar
	var filenames []string

	r, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	showProgress := !logger.QuietFlag && len(r.File) > 0
	if showProgress {
		bar = pb.New(len(r.File)).SetUnits(pb.U_NO).SetRefreshRate(time.Millisecond * 10)
		bar.Start()
		defer bar.Finish()
	}

	for _, f := range r.File {

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}
		defer rc.Close()

		if showProgress {
			bar.Prefix("[" + f.Name + "]")
		}

		// Store filename/path for returning and using later on
		fpath := filepath.Join(destDirectory, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(destDirectory)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {

			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)

		} else {

			// Make File
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return filenames, err
			}

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return filenames, err
			}

			_, err = io.Copy(outFile, rc)

			// Close the file without defer to close before next iteration of loop
			outFile.Close()

			if err != nil {
				return filenames, err
			}
		}

		if showProgress {
			bar.Increment()
		}
	}

	return filenames, nil
}
