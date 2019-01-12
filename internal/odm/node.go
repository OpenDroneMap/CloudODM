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
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	odmio "github.com/OpenDroneMap/CloudODM/internal/io"
	"github.com/OpenDroneMap/CloudODM/internal/logger"
	"github.com/cheggaaa/pb"
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

type StatusCode struct {
	Code int `json:"code"`
}

type TaskInfoResponse struct {
	ProcessingTime int        `json:"processingTime"`
	Status         StatusCode `json:"status"`
}

type ApiActionResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
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
	targetURL := n.URLFor(path)
	logger.Debug("POST: " + targetURL)

	formData := url.Values{}
	for k, v := range values {
		formData.Set(k, v)
		logger.Debug(k + ": " + v)
	}

	resp, err := http.Post(targetURL, "application/x-www-form-urlencoded", strings.NewReader(formData.Encode()))
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

// TaskInfo GET: /task/<uuid>/info
func (n Node) TaskInfo(uuid string) (*TaskInfoResponse, error) {
	res := TaskInfoResponse{}
	body, err := n.apiGET("/task/" + uuid + "/info")
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// TaskOutput GET: /task/<uuid>/output
func (n Node) TaskOutput(uuid string, line int) ([]string, error) {
	res := []string{}
	body, err := n.apiGET("/task/" + uuid + "/output?line=" + strconv.Itoa(line))
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return res, nil
}

// TaskDownload GET: /task/<uuid>/download/<asset>
func (n Node) TaskDownload(uuid string, asset string, outputFile string) error {
	var bar *pb.ProgressBar

	out, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(n.URLFor("/task/" + uuid + "/download/" + asset))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	totalBytes, err := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		totalBytes = 0
		logger.Debug("Warning: Content-length not set")
	}

	showProgress := !logger.QuietFlag && totalBytes > 0
	if showProgress {
		bar = pb.New64(totalBytes).SetUnits(pb.U_BYTES).SetRefreshRate(time.Millisecond * 10)
		bar.Start()
		bar.Prefix("[" + out.Name() + "]")
	}

	writer := io.MultiWriter(out, bar)

	written, err := io.Copy(writer, resp.Body)
	if written == 0 {
		return errors.New("Download returned 0 bytes")
	}

	if showProgress {
		bar.Finish()
	}

	return nil
}

// TaskCancel POST: /task/cancel
func (n Node) TaskCancel(uuid string) error {
	res := ApiActionResponse{}

	body, err := n.apiPOST("/task/cancel", map[string]string{"uuid": uuid})
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return err
	}

	if res.Error != "" {
		return errors.New(res.Error)
	}

	return nil
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
