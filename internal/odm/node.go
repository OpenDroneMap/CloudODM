package odm

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	odmio "github.com/OpenDroneMap/CloudODM/internal/io"
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

type OptionResponse struct {
	Domain interface{} `json:"domain"`
	Help   string      `json:"help"`
	Name   string      `json:"name"`
	Type   string      `json:"type"`
	Value  string      `json:"value"`
}

type AuthInfoResponse struct {
	Message     string `json:"message"`
	LoginUrl    string `json:"loginUrl"`
	RegisterUrl string `json:"registerUrl"`
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
		return nil, err
	}

	return body, nil
}

func (n Node) apiPOST(path string, values map[string]string) ([]byte, error) {
	url := n.URLFor(path)
	logger.Debug("POST: " + url)

	formData, _ := json.Marshal(values)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(formData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("Server returned status code: " + strconv.Itoa(resp.StatusCode))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
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

// Options GET: /options
func (n Node) Options() ([]OptionResponse, error) {
	options := []OptionResponse{}
	body, err := n.apiGET("/options")
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &options); err != nil {
		return nil, err
	}

	sort.Slice(options, func(i, j int) bool {
		return options[i].Name < options[j].Name
	})

	return options, nil
}

func (n Node) CheckAuthentication(err error) error {
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

type LoginResponse struct {
	Token string `json:"token"`
}

func (n Node) TryLogin(username string, password string) (token string, err error) {
	res := AuthInfoResponse{}
	body, err := n.apiGET("/auth/info")
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return "", err
	}

	if res.Message != "" {
		logger.Info("")
		logger.Info(res.Message)
		logger.Info("")
	}

	if res.LoginUrl != "" {
		if username == "" && password == "" {
			username, password = odmio.GetUsernamePassword()
		}

		logger.Debug("")
		logger.Debug("POST: " + res.LoginUrl)

		formData, _ := json.Marshal(map[string]string{"username": username, "password": password})
		resp, err := http.Post(res.LoginUrl, "application/json", bytes.NewBuffer(formData))
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return "", errors.New("Login URL returned status code: " + strconv.Itoa(resp.StatusCode))
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		res := LoginResponse{}
		if err := json.Unmarshal(body, &res); err != nil {
			return "", err
		}

		if res.Token == "" {
			return "", errors.New("Login failed")
		}

		return res.Token, nil
	}

	// TODO: support for res.RegisterUrl

	return "", errors.New("Cannot login")
}
