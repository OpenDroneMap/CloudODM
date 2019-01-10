package odm

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/OpenDroneMap/CloudODM/internal/logger"

	"github.com/cheggaaa/pb"
)

type TaskNewResponse struct {
	UUID  string `json:"uuid"`
	Error string `json:"error"`
}

// Run processes a dataset
func Run(files []string, options []Option, node Node) {
	var err error
	var bar *pb.ProgressBar

	var f *os.File
	var fi os.FileInfo

	var totalBytes int64

	// Calculate total upload size
	for _, file := range files {
		if fi, err = os.Stat(file); err != nil {
			logger.Error(err)
		}
		totalBytes += fi.Size()
		f.Close()
	}

	bar = pb.New64(totalBytes).SetUnits(pb.U_BYTES).SetRefreshRate(time.Millisecond * 10)
	bar.Start()

	// Convert options to JSON
	jsonOptions, err := json.Marshal(options)
	if err != nil {
		logger.Error(err)
	}

	// Setup pipe
	r, w := io.Pipe()
	mpw := multipart.NewWriter(w)

	// Pipe work, stream file contents
	go func() {
		var part io.Writer
		defer w.Close()
		defer f.Close()

		for _, file := range files {
			if f, err = os.Open(file); err != nil {
				logger.Error(err)
			}
			if fi, err = f.Stat(); err != nil {
				logger.Error(err)
			}

			if part, err = mpw.CreateFormFile("images", fi.Name()); err != nil {
				logger.Error(err)
			}
			bar.Prefix("[" + fi.Name() + "]")
			part = io.MultiWriter(part, bar)

			if _, err = io.Copy(part, f); err != nil {
				logger.Error(err)
			}
		}

		mpw.WriteField("skipPostProcessing", "true")
		mpw.WriteField("options", string(jsonOptions))

		if err = mpw.Close(); err != nil {
			logger.Error(err)
		}
	}()

	resp, err := http.Post(node.URLFor("/task/new"), mpw.FormDataContentType(), r)
	if err != nil {
		logger.Error(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err)
	}

	res := TaskNewResponse{}
	if err := json.Unmarshal(body, &res); err != nil {
		logger.Error(err)
	}

	// Handle error response from API
	if res.Error != "" {
		logger.Error(res.Error)
	}

	// We should have a UUID
	// Start listening for output
	// TODO
}
