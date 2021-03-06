package main

import (
	"io/ioutil"
	"net/http"
	"path"
	"strings"
)

const maxMemory int64 = 1024 * 1024 * 64

func readFormPayload(r *http.Request) ([]byte, error) {
	file, _, err := r.FormFile("file")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buf, err := ioutil.ReadAll(file)
	if len(buf) == 0 {
		err = ErrEmptyPayload
	}

	return buf, err
}

func readBodyType(r *http.Request) ([]byte, error) {
	contentType := r.Header.Get("Content-Type")

	if strings.HasPrefix(contentType, "multipart/") {
		err := r.ParseMultipartForm(maxMemory)
		if err != nil {
			return nil, err
		}
		return readFormPayload(r)
	}

	return ioutil.ReadAll(r.Body)
}

func readPayload(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	if r.Method != "POST" {
		return nil, ErrorReply(w, ErrMethodNotAllowed)
	}

	buf, err := readBodyType(r)
	if err != nil {
		return nil, ErrorReply(w, NewError("Cannot read payload: "+err.Error(), BAD_REQUEST))
	}

	return buf, nil
}

func readLocalImage(w http.ResponseWriter, r *http.Request, mountPath string) ([]byte, error) {
	file := r.URL.Query().Get("file")
	if file == "" {
		return nil, ErrorReply(w, ErrMissingParamFile)
	}

	file = path.Clean(path.Join(mountPath, file))
	if strings.HasPrefix(file, mountPath) == false {
		return nil, ErrorReply(w, ErrInvalidFilePath)
	}

	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, ErrorReply(w, ErrInvalidFilePath)
	}

	return buf, nil
}
