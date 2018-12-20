package fs

import "os"

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
