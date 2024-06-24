package assets

import (
	"embed"
	"io"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
)

//nolint:typecheck
//go:embed spa/**
var spaFS embed.FS

//go:embed spa/index.html
var spaIndex []byte

type IndexHandler struct{}

func (i *IndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(spaIndex); err != nil {
		slog.Info("error serving SPA index", slog.Any("error", err))
	}
}

func SPAMRegister(mux *http.ServeMux) {
	subFS, subErr := fs.Sub(spaFS, "spa")
	if subErr != nil {
		log.Fatalf("Could create sub filesystem: %s", subFS)
	}

	httpFs := http.FS(subFS)
	fileServer := http.FileServer(httpFs)
	mux.Handle("/assets/", fileServer)
	mux.Handle("/vite.svg", fileServer)
	mux.Handle("/", &IndexHandler{})
}

//go:embed images/placeholder_tool_cover.png
var image embed.FS

func CopyToolPlaceholderCover(destPath string) error {
	srcFile, err := image.Open("images/placeholder_tool_cover.png")
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}

	return nil
}
