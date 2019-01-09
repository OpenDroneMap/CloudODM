package odm

import (
	"testing"
)

func TestAPI(t *testing.T) {
	node := Node{URL: "http://localhost:3000"}
	offlineNode := Node{URL: "http://unknownhost:3000"}

	resp, err := node.Info()
	if err != nil {
		t.Error("Cannot retrieve /info")
	}

	if resp.Version == "" {
		t.Error("Version is not set")
	}

	resp, err = offlineNode.Info()
	if err == nil {
		t.Error("No error when retrieving an offline node /info")
	}

	node._debugUnauthorized = true
	resp, err = node.Info()
	t.Log(resp, err)
	if err != ErrUnauthorized {
		t.Error("Error should have been ErrUnauthorized")
	}
}
