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

package fs

import (
	"io/ioutil"
	"os"
)

// FileExists checks if a file path exists
func FileExists(filePath string) (bool, error) {
	exists := true
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			exists = false
		} else {
			return false, err
		}
	}

	return exists, nil
}

// IsDirectory checks whether a path is a directory
func IsDirectory(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

// IsFile checks whether a path is a directory
func IsFile(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !fileInfo.IsDir()
}

// DirectoryFilesCount returns the number of files in a directory. If the dir
// does not exists, it returns 0.
func DirectoryFilesCount(dirPath string) (int, error) {
	if _, err := os.Stat(dirPath); err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		} else {
			return -1, err
		}
	}

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return -1, err
	}

	return len(files), nil
}
