package odm

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/OpenDroneMap/CloudODM/internal/logger"
)

// ErrUnauthorized means a response was not authorized
var ErrUnauthorized = errors.New("Unauthorized")

// ErrAuthRequired means authorization is required
var ErrAuthRequired = errors.New("Auth Required")

type InfoResponse struct {
	Version   string `json:"version"`
	MaxImages int    `json:"maxImages"`

	Error string `json:"error"`
}

// Node is a NodeODM processing node
type Node struct {
	URL   string `json:"url"`
	Token string `json:"token"`

	_debugUnauthorized bool
}

func (n Node) String() string {
	return n.URL
}

// URLFor builds a URL path
func (n Node) URLFor(path string) string {
	u, err := url.ParseRequestURI(n.URL + path)
	if err != nil {
		return ""
	}
	q := u.Query()
	if len(n.Token) > 0 {
		q.Add("token", n.Token)
	}
	if n._debugUnauthorized {
		q.Add("_debugUnauthorized", "1")
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func (n Node) apiGET(path string) ([]byte, error) {
	url := n.URLFor(path)
	logger.Debug("GET: " + url)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return nil, ErrUnauthorized
	}
	if resp.StatusCode != 200 {
		return nil, errors.New("Server returned status code: " + strconv.Itoa(resp.StatusCode))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Info(err)
		return nil, err
	}

	return body, nil
}

// Info GET: /info
func (n Node) Info() (*InfoResponse, error) {
	res := InfoResponse{}
	body, err := n.apiGET("/info")
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	if res.Error != "" {
		if strings.HasPrefix(res.Error, "Invalid authentication token") {
			return nil, ErrUnauthorized
		}
		return nil, errors.New(res.Error)
	}

	if res.MaxImages == 0 {
		res.MaxImages = math.MaxInt32
	}

	return &res, nil
}

func (n Node) CheckAuthorization(err error) error {
	if err != nil {
		if err == ErrUnauthorized {
			// Is there a token?
			if n.Token == "" {
				return ErrAuthRequired
			} else {
				return errors.New("Cannot authenticate with the node (invalid token).")
			}
		} else {
			return err
		}
	}

	return nil
}
