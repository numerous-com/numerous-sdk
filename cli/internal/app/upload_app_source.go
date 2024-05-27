package app

import (
	"bytes"
	"errors"
	"io"
	"net/http"
)

var ErrAppSourceUpload = errors.New("error uploading app source")

func (s *Service) UploadAppSource(uploadURL string, archive io.Reader) error {
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, archive); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, uploadURL, &buf)
	if err != nil {
		return err
	}

	resp, err := s.uploadDoer.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return ErrAppSourceUpload
	}

	return nil
}
