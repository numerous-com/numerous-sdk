package app

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

type AppSourceUploadError struct {
	HTTPStatusCode int
	HTTPStatus     string
	UploadURL      string
	ResponseBody   []byte
}

func (e *AppSourceUploadError) Error() string {
	return fmt.Sprintf("http %d: %q uploading app source file to %q ", e.HTTPStatusCode, e.HTTPStatus, e.UploadURL)
}

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
		responseBody, _ := io.ReadAll(resp.Body)
		return &AppSourceUploadError{
			HTTPStatusCode: resp.StatusCode,
			HTTPStatus:     resp.Status,
			UploadURL:      uploadURL,
			ResponseBody:   responseBody,
		}
	}

	return nil
}
