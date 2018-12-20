package config

import (
	"fmt"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
)

type Configuration struct {
	nodes map[string]Node
}

type Node struct {
	url   string
	token string
}

// Initialize the configuration
func Initialize() {
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cfgPath := filepath.Join(home, ".odm.json")
	GetPublicNodes()
	fmt.Println(cfgPath)
}
