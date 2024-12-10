package app

import (
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

type UploadArchive struct {
	Reader io.Reader
	Size   int64
}

func (s *Service) UploadAppSource(uploadURL string, archive UploadArchive) error {
	req, err := http.NewRequest(http.MethodPut, uploadURL, archive.Reader)
	if err != nil {
		return err
	}

	req.ContentLength = archive.Size

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
