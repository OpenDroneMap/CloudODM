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

package config

import (
	"os"
	"testing"
)

func TestNodes(t *testing.T) {
	os.Remove("test.json")
	c := NewConfiguration("test.json")

	c.AddNode("default", "http://localhost")
	if value, ok := c.Nodes["default"]; ok {
		if value.Token != "" {
			t.Error("Token should not have been set")
		}
	} else {
		t.Error("Default has not been added")
	}

	if c.AddNode("default", "http://test123") == nil {
		t.Error("Should not be able to add an existing node")
	}

	c.AddNode("test1", "https://localhost:8080/?token=123")
	value, _ := c.Nodes["test1"]
	if value.Token != "123" {
		t.Error("Token should be 123")
	}
	if value.URL != "https://localhost:8080" {
		t.Error("URL is not properly set")
	}

	if c.AddNode("test2", "invalid-url") == nil {
		t.Error("We shouldn't be able to add invalid URLs")
	}

	if !c.RemoveNode("test1") {
		t.Error("Function should have returned true")
	}

	if _, ok := c.Nodes["test1"]; ok {
		t.Error("test1 should have been removed")
	}

	if c.RemoveNode("test1") {
		t.Error("Function should have returned false (already deleted)")
	}

	if _, err := c.GetNode("default"); err != nil {
		t.Error("Cannot get default node")
	}

	if _, err := c.GetNode("nonexistant"); err == nil {
		t.Error("Can get nonexistant node")
	}
}
