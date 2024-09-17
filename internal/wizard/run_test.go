package wizard

import (
	"testing"

	"numerous.com/cli/internal/manifest"

	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	expectedExclude := []string{"*venv", "venv*", ".git", ".env"}
	expectedCover := "app_cover.jpg"

	t.Run("it updates the manifest with expected values", func(t *testing.T) {
		path := t.TempDir()

		for _, tc := range []struct {
			name      string
			params    RunWizardParams
			expected  manifest.Manifest
			stubAsker StubAsker
		}{
			{
				name: "streamlit wizard answers",
				stubAsker: StubAsker{
					UseFolderQuestion(path): true,
					"Name":                  "App Name",
					"Description":           "App description",
					"LibraryName":           "Streamlit",
					"AppFile":               "app.py",
					"RequirementsFile":      "requirements.txt",
				},
				expected: manifest.Manifest{
					App:    manifest.App{Name: "App Name", Description: "App description", CoverImage: expectedCover, Exclude: expectedExclude},
					Python: &manifest.Python{Library: manifest.LibraryStreamlit, AppFile: "app.py", RequirementsFile: "requirements.txt", Port: manifest.LibraryStreamlit.Port},
				},
			},
			{
				name: "marimo from input",
				stubAsker: StubAsker{
					UseFolderQuestion(path): true,
					"Name":                  "Not Used App Name",
					"Description":           "Not Used App description",
					"LibraryName":           "Not Used Library",
					"AppFile":               "Not Used App File.py",
					"RequirementsFile":      "Not Used Requirements.txt",
				},
				params: RunWizardParams{
					App:    AppAnswers{Name: "App Name From Input", Description: "App description from input"},
					Python: PythonAnswers{Library: manifest.LibraryMarimo, AppFile: "app file from input.py", RequirementsFile: "requirements from input.txt"},
				},
				expected: manifest.Manifest{
					App:    manifest.App{Name: "App Name From Input", Description: "App description from input", CoverImage: expectedCover, Exclude: expectedExclude},
					Python: &manifest.Python{Library: manifest.LibraryMarimo, AppFile: "app file from input.py", RequirementsFile: "requirements from input.txt", Port: manifest.LibraryMarimo.Port},
				},
			},
			{
				name: "app name given as input",
				stubAsker: StubAsker{
					UseFolderQuestion(path): true,
					"Description":           "App description",
					"LibraryName":           "Streamlit",
					"AppFile":               "app.py",
					"RequirementsFile":      "requirements.txt",
				},
				params: RunWizardParams{App: AppAnswers{Name: "App Name from Input"}},
				expected: manifest.Manifest{
					App:    manifest.App{Name: "App Name from Input", Description: "App description", CoverImage: expectedCover, Exclude: expectedExclude},
					Python: &manifest.Python{Library: manifest.LibraryStreamlit, AppFile: "app.py", RequirementsFile: "requirements.txt", Port: manifest.LibraryStreamlit.Port},
				},
			},
			{
				name: "app description given as input",
				stubAsker: StubAsker{
					UseFolderQuestion(path): true,
					"Name":                  "App Name",
					"LibraryName":           "Streamlit",
					"AppFile":               "app.py",
					"RequirementsFile":      "requirements.txt",
				},
				params: RunWizardParams{App: AppAnswers{Description: "App description from input"}},
				expected: manifest.Manifest{
					App:    manifest.App{Name: "App Name", Description: "App description from input", CoverImage: expectedCover, Exclude: expectedExclude},
					Python: &manifest.Python{Library: manifest.LibraryStreamlit, AppFile: "app.py", RequirementsFile: "requirements.txt", Port: manifest.LibraryStreamlit.Port},
				},
			},
			{
				name: "app python library given as input",
				stubAsker: StubAsker{
					UseFolderQuestion(path): true,
					"Name":                  "App Name",
					"Description":           "App description",
					"AppFile":               "app.py",
					"RequirementsFile":      "requirements.txt",
				},
				params: RunWizardParams{App: AppAnswers{LibraryName: manifest.LibraryPlotlyDash.Name, LibraryKey: manifest.LibraryPlotlyDash.Key}, Python: PythonAnswers{Library: manifest.LibraryPlotlyDash}},
				expected: manifest.Manifest{
					App:    manifest.App{Name: "App Name", Description: "App description", CoverImage: expectedCover, Exclude: expectedExclude},
					Python: &manifest.Python{Library: manifest.LibraryPlotlyDash, AppFile: "app.py", RequirementsFile: "requirements.txt", Port: manifest.LibraryPlotlyDash.Port},
				},
			},
			{
				name: "app python requirements file given as input",
				stubAsker: StubAsker{
					UseFolderQuestion(path): true,
					"Name":                  "App Name",
					"LibraryName":           "Streamlit",
					"Description":           "App description",
					"AppFile":               "app.py",
				},
				params: RunWizardParams{Python: PythonAnswers{RequirementsFile: "requirements from input.txt"}},
				expected: manifest.Manifest{
					App:    manifest.App{Name: "App Name", Description: "App description", CoverImage: expectedCover, Exclude: expectedExclude},
					Python: &manifest.Python{Library: manifest.LibraryStreamlit, AppFile: "app.py", RequirementsFile: "requirements from input.txt", Port: manifest.LibraryStreamlit.Port},
				},
			},
			{
				name: "app python app file given as input",
				stubAsker: StubAsker{
					UseFolderQuestion(path): true,
					"Name":                  "App Name",
					"LibraryName":           "Streamlit",
					"Description":           "App description",
					"RequirementsFile":      "requirements.txt",
				},
				params: RunWizardParams{Python: PythonAnswers{AppFile: "app from input.py"}},
				expected: manifest.Manifest{
					App:    manifest.App{Name: "App Name", Description: "App description", CoverImage: expectedCover, Exclude: expectedExclude},
					Python: &manifest.Python{Library: manifest.LibraryStreamlit, AppFile: "app from input.py", RequirementsFile: "requirements.txt", Port: manifest.LibraryStreamlit.Port},
				},
			},
			{
				name: "docker app from questions",
				stubAsker: StubAsker{
					UseFolderQuestion(path): true,
					"Name":                  "App Name",
					"Description":           "App description",
					"LibraryName":           "Dockerfile",
					"Dockerfile":            "Dockerfile from question",
					"Context":               "Docker context from question",
				},
				expected: manifest.Manifest{
					App:    manifest.App{Name: "App Name", Description: "App description", CoverImage: expectedCover, Exclude: expectedExclude},
					Docker: &manifest.Docker{Dockerfile: "Dockerfile from question", Context: "Docker context from question"},
				},
			},
			{
				name: "docker app from input",
				stubAsker: StubAsker{
					UseFolderQuestion(path): true,
					"Name":                  "App Name",
					"Description":           "App description",
				},
				params: RunWizardParams{
					App:    AppAnswers{LibraryKey: manifest.DockerfileLibraryKey},
					Docker: DockerAnswers{Dockerfile: "Dockerfile from input", Context: "Docker context from input"},
				},
				expected: manifest.Manifest{
					App:    manifest.App{Name: "App Name", Description: "App description", CoverImage: expectedCover, Exclude: expectedExclude},
					Docker: &manifest.Docker{Dockerfile: "Dockerfile from input", Context: "Docker context from input"},
				},
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				tc.params.ProjectFolderPath = path
				m, _ := Run(&tc.stubAsker, tc.params)

				if assert.NotNil(t, m) {
					assert.Equal(t, tc.expected, *m)
				}
			})
		}
	})

	t.Run("interrupt returns no error and no continue", func(t *testing.T) {
		path := t.TempDir()
		for _, tc := range []struct {
			questionToInterrupt string
			libraryName         string
		}{
			{questionToInterrupt: "Name", libraryName: "Streamlit"},
			{questionToInterrupt: "Description", libraryName: "Streamlit"},
			{questionToInterrupt: "AppFile", libraryName: "Streamlit"},
			{questionToInterrupt: "LibraryName", libraryName: "Streamlit"},
			{questionToInterrupt: "RequirementsFile", libraryName: "Streamlit"},
			{questionToInterrupt: "Dockerfile", libraryName: "Dockerfile"},
			{questionToInterrupt: "Context", libraryName: "Dockerfile"},
			{questionToInterrupt: UseFolderQuestion(path)},
		} {
			t.Run(tc.questionToInterrupt, func(t *testing.T) {
				stubAsker := StubAsker{
					UseFolderQuestion(path): true,
					"Name":                  "App Name",
					"Description":           "App description",
					"LibraryName":           tc.libraryName,
					"AppFile":               "app.py",
					"RequirementsFile":      "requirements.txt",
					"Dockerfile":            "Dockerfile",
					"Context":               "Docker context",
				}
				// override answer of question to interrupt with interrupt error
				stubAsker[tc.questionToInterrupt] = terminal.InterruptErr

				_, err := Run(&stubAsker, RunWizardParams{ProjectFolderPath: path})

				assert.ErrorIs(t, err, ErrStopInit)
			})
		}
	})
}
