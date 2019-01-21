package handler

import (
	"bytes"
	"github.com/ameliaikeda/sakura"
	"github.com/monzo/typhon"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"sync"
)

type response struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

type config struct {
	name string
	file io.Reader
	process func(body io.Reader) ([]byte, error)
	upload func(name string, buf []byte) error
}

const (
	errFormNotPresent = "multipart form expected, none given"
	errImageMissing   = `key "image" is missing from form files.`
	errInvalidFiles   = `there should be exactly one file given under the "image" key`
	errOpeningFile    = "error opening uploaded file"
	errReadingFile    = "error reading uploaded file"
	errNameMissing    = `key "name" is missing from form values`
	errNameEmpty      = `key "name" must not be empty`
)

func Service(g sakura.Generator, u sakura.Uploader) typhon.Service {
	return func(req typhon.Request) typhon.Response {
		form := req.MultipartForm
		if form == nil {
			res := req.Response(&response{Success: false, Error: errFormNotPresent})
			res.StatusCode = http.StatusUnprocessableEntity

			return res
		}

		// we need to free all temporary files after processing.
		defer form.RemoveAll()

		name, err := filename(form)
		if err != nil {
			return badRequest(req, err)
		}

		file, err := file(form)
		if err != nil {
			return badRequest(req, err)
		}

		// generate both main image and thumbnail simultaneously? potential RAM situation as resizing is heavy.
		var wg sync.WaitGroup

		wg.Add(2)

		go processAndUpload(wg, config{
			name: name,
			file: file,
			process: g.ProcessThumbnail,
			upload: u.UploadThumbnail,
		})

		go processAndUpload(wg, config{
			name: name,
			file: file,
			process: g.ProcessImage,
			upload: u.UploadImage,
		})

		wg.Wait()

		return req.Response(&response{Success: true})
	}
}

func badRequest(req typhon.Request, err error) typhon.Response {
	res := req.Response(&response{Success: false, Error: err.Error()})
	res.StatusCode = http.StatusUnprocessableEntity
	res.Error = err

	return res
}

func file(form *multipart.Form) (io.Reader, error) {
	h, ok := form.File["image"]
	if !ok {
		return nil, errors.New(errImageMissing)
	}

	if len(h) != 1 {
		return nil, errors.New(errInvalidFiles)
	}

	header := h[0]

	f, err := header.Open()
	if err != nil {
		return nil, errors.New(errOpeningFile)
	}

	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.New(errReadingFile)
	}

	return bytes.NewBuffer(b), nil
}

func filename(form *multipart.Form) (string, error) {
	name, ok := form.Value["name"]
	if !ok {
		return "", errors.New(errNameMissing)
	}

	n := name[0]

	if len(n) == 0 {
		return "", errors.New(errNameEmpty)
	}

	return n, nil
}

func processAndUpload(wg sync.WaitGroup, cfg config) {
	defer wg.Done()

	// TODO: func body
}
