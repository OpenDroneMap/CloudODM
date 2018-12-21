package config

import (
	"errors"
	"net/url"
)

// AddNode adds a new node to the configuration
func (c Configuration) AddNode(name string, nodeURL string) error {
	if _, ok := c.Nodes[name]; ok {
		return errors.New("node" + name + " already exists. Remove it first.")
	}

	u, err := url.ParseRequestURI(nodeURL)
	if err != nil {
		return errors.New(nodeURL + " is not a valid URL. A valid URL looks like: http://hostname:port/?token=optional")
	}

	c.Nodes[name] = Node{Url: u.Scheme + "://" + u.Host, Token: u.Query().Get("token")}
	c.Save()

	return nil
}

// RemoveNode removes a node from the configuration
func (c Configuration) RemoveNode(name string) bool {
	_, ok := c.Nodes[name]
	if ok {
		delete(c.Nodes, name)
		c.Save()
	}
	return ok
}
