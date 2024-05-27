package app

import (
	"errors"
	"io"
	"net/http"
)

var ErrAppSourceUpload = errors.New("error uploading app source")

func (s *Service) UploadAppSource(uploadURL string, archive io.Reader) error {
	req, err := http.NewRequest(http.MethodPut, uploadURL, archive)
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
