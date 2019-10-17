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
	"mime/multipart"
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

type TaskNewResponse struct {
	UUID  string `json:"uuid"`
	Error string `json:"error"`
}

// Node is a NodeODM processing node
type Node struct {
	URL   string `json:"url"`
	Token string `json:"token"`

	_debugUnauthorized bool
}

type downloadChunk struct {
	partNum    int
	bytesStart int64
	bytesEnd   int64
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

func downloadWorker(chunksToProcess <-chan downloadChunk, chunksProcessed chan<- downloadChunk) {
	for c := range chunksToProcess {
		chunksProcessed <- c
	}
}

// TaskDownload GET: /task/<uuid>/download/<asset>
func (n Node) TaskDownload(uuid string, asset string, outputFile string, parallelConnections int) error {
	out, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer out.Close()

	//resp, err := http.Get(n.URLFor("/task/" + uuid + "/download/" + asset))
	resp, err := http.Get("https://wln2.fra1.digitaloceanspaces.com/afdc2b64-a484-4624-a428-ed65e9c7a1bf/all.zip")

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	totalBytes, err := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		totalBytes = 0
		logger.Debug("Warning: Content-length not set")
	}

	const ChunkSize int64 = 10.0 * 1024.0 * 1024.0 // MB
	acceptRanges := strings.ToLower(resp.Header.Get("Accept-Ranges"))
	showProgress := !logger.QuietFlag && totalBytes > 0

	// Parallel?
	if acceptRanges == "bytes" && totalBytes > 0 && totalBytes > ChunkSize && parallelConnections > 1 {
		numChunks := int(math.Ceil(float64(totalBytes) / float64(ChunkSize)))

		chunksToProcess := make(chan downloadChunk, numChunks)
		chunksProcessed := make(chan downloadChunk, numChunks)

		for w := 1; w <= parallelConnections; w++ {
			go downloadWorker(chunksToProcess, chunksProcessed)
		}

	} else {
		var bar *pb.ProgressBar

		if showProgress {
			bar = pb.New64(totalBytes).SetUnits(pb.U_BYTES).SetRefreshRate(time.Millisecond * 10)
			bar.Start()
			bar.Prefix("[" + asset + "]")
		}

		var writer io.Writer
		if bar != nil {
			writer = io.MultiWriter(out, bar)
		} else {
			writer = out
		}

		written, err := io.Copy(writer, resp.Body)
		if written == 0 {
			return errors.New("Download returned 0 bytes")
		}

		if showProgress {
			bar.Finish()
		}
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

// TaskNewInit POST: /task/new/init
func (n Node) TaskNewInit(jsonOptions []byte) TaskNewResponse {
	var err error
	reqBody := &bytes.Buffer{}
	mpw := multipart.NewWriter(reqBody)
	mpw.WriteField("skipPostProcessing", "true")
	mpw.WriteField("options", string(jsonOptions))
	if err = mpw.Close(); err != nil {
		return TaskNewResponse{"", err.Error()}
	}

	resp, err := http.Post(n.URLFor("/task/new/init"), mpw.FormDataContentType(), reqBody)
	defer resp.Body.Close()
	if err != nil {
		return TaskNewResponse{"", err.Error()}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return TaskNewResponse{"", err.Error()}
	}

	var res TaskNewResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return TaskNewResponse{"", err.Error()}
	}

	return res
}

// TaskNewUpload POST: /task/new/upload/{uuid}
func (n Node) TaskNewUpload(file string, uuid string, bar *pb.ProgressBar) error {
	var f *os.File
	var fi os.FileInfo
	var err error
	r, w := io.Pipe()
	mpw := multipart.NewWriter(w)

	go func() {
		var part io.Writer
		defer w.Close()
		defer f.Close()

		if f, err = os.Open(file); err != nil {
			return
		}
		if fi, err = f.Stat(); err != nil {
			return
		}

		if bar != nil {
			bar.SetTotal64(fi.Size())
			bar.Set64(0)
			bar.Prefix("[" + fi.Name() + "]")
		}

		if part, err = mpw.CreateFormFile("images", fi.Name()); err != nil {
			return
		}

		if bar != nil {
			part = io.MultiWriter(part, bar)
		}

		if _, err = io.Copy(part, f); err != nil {
			return
		}

		if err = mpw.Close(); err != nil {
			return
		}
	}()

	resp, err := http.Post(n.URLFor("/task/new/upload/"+uuid), mpw.FormDataContentType(), r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var res ApiActionResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return err
	}

	if res.Error != "" {
		return errors.New(res.Error)
	}

	if !res.Success {
		return errors.New("Cannot complete upload. /task/new/upload failed with success: false")
	}

	return nil
}

// TaskNewCommit POST: /task/new/commit/{uuid}
func (n Node) TaskNewCommit(uuid string) TaskNewResponse {
	var res TaskNewResponse

	body, err := n.apiPOST("/task/new/commit/"+uuid, map[string]string{})
	if err != nil {
		return TaskNewResponse{"", err.Error()}
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return TaskNewResponse{"", err.Error()}
	}

	return res
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
