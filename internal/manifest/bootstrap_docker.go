package manifest

import (
	"errors"
	"fmt"
	"path/filepath"
)

const DockerfileLibraryKey = "dockerfile"

var ErrNoBootstrapDockerfileExists = errors.New("cannot bootstrap with pre-existing dockerfile")

const dockerExampleDockerfile = `FROM python:3.11-slim
EXPOSE %d

COPY requirements.txt /app/requirements.txt
RUN pip install -r /app/requirements.txt

COPY app.py /app/app.py

CMD ["streamlit", "run", "/app/app.py", "--server.port", "%d"]`

const dockerExampleAppPy = `import streamlit as st


st.header("Your app built from a Dockerfile is running!")

st.markdown(
    "The Dockerfile in your app folder can be modified to build your app."
    "\n\n"
    "The Numerous CLI has initialized this app as an example of how to build "
    "your app with a Dockerfile."
    "\n\n"
    "See https://docs.docker.com/build/concepts/dockerfile/ for information "
    "about how to write your Dockerfile."
)
`

const dockerExampleRequirementsTxt = "streamlit\n"

func (d Docker) bootstrapFiles(basePath string, port uint) error {
	dockerfilePath := filepath.Join(basePath, d.Dockerfile)
	appPath := filepath.Join(basePath, "app.py")
	requirementsPath := filepath.Join(basePath, "requirements.txt")

	if exists, err := fileExists(dockerfilePath); err != nil {
		return err
	} else if exists {
		return ErrNoBootstrapDockerfileExists
	}

	if err := createAndWriteIfFileNotExist(dockerfilePath, fmt.Sprintf(dockerExampleDockerfile, port, port)); err != nil {
		return err
	}

	if err := createAndWriteIfFileNotExist(appPath, dockerExampleAppPy); err != nil {
		return err
	}

	if err := createAndWriteIfFileNotExist(requirementsPath, dockerExampleRequirementsTxt); err != nil {
		return err
	}

	return nil
}
