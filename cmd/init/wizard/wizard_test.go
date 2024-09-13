package wizard

import (
	"testing"

	"numerous.com/cli/internal/manifest"

	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/stretchr/testify/assert"
)

func TestGetQuestions(t *testing.T) {
	t.Run("given empty manifest gets all questions", func(t *testing.T) {
		qs := getQuestions(&manifest.Manifest{
			Python: &manifest.Python{},
		})

		assert.Len(t, qs, 5)
	})

	t.Run("given full manifest gets no questions", func(t *testing.T) {
		qs := getQuestions(&manifest.Manifest{
			App: manifest.App{
				Name:        "Some name",
				Description: "Some description",
			},
			Python: &manifest.Python{
				Library:          manifest.LibraryNumerous,
				AppFile:          "app.py",
				RequirementsFile: "requirements.txt",
			},
		})

		assert.Empty(t, qs)
	})
}

func TestRunInitAppWizard(t *testing.T) {
	t.Run("it updates the manifest with expected values", func(t *testing.T) {
		path := t.TempDir()

		for _, tc := range []struct {
			name      string
			input     manifest.Manifest
			expected  manifest.Manifest
			stubAsker StubAsker
		}{
			{
				stubAsker: StubAsker{
					useFolderQuestion(path): true,
					"Name":                  "App Name",
					"Description":           "App description",
					"LibraryName":           "Streamlit",
					"AppFile":               "app.py",
					"RequirementsFile":      "requirements.txt",
				},
				input: manifest.Manifest{Python: &manifest.Python{}},
				expected: manifest.Manifest{
					App: manifest.App{
						Name:        "App Name",
						Description: "App description",
					},
					Python: &manifest.Python{
						Library:          manifest.LibraryStreamlit,
						AppFile:          "app.py",
						RequirementsFile: "requirements.txt",
						Port:             manifest.LibraryStreamlit.Port,
					},
				},
			},
			{
				stubAsker: StubAsker{
					useFolderQuestion(path): true,
					"Name":                  "Not Used App Name",
					"Description":           "Not Used App description",
					"LibraryName":           "Not Used Library",
					"AppFile":               "Not Used App File.py",
					"RequirementsFile":      "Not Used Requirements.txt",
				},
				input: manifest.Manifest{
					App: manifest.App{
						Name:        "App Name From Input",
						Description: "App description from input",
					},
					Python: &manifest.Python{
						Library:          manifest.LibraryMarimo,
						AppFile:          "app file from input.py",
						RequirementsFile: "requirements from input.txt",
						Port:             manifest.LibraryMarimo.Port,
					},
				},
				expected: manifest.Manifest{
					App: manifest.App{
						Name:        "App Name From Input",
						Description: "App description from input",
					},
					Python: &manifest.Python{
						Library:          manifest.LibraryMarimo,
						AppFile:          "app file from input.py",
						RequirementsFile: "requirements from input.txt",
						Port:             manifest.LibraryMarimo.Port,
					},
				},
			},
		} {
			actual := tc.input
			cont, err := RunInitAppWizard(&tc.stubAsker, path, &actual)

			assert.True(t, cont)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		}
	})

	t.Run("interrupt returns no error and no continue", func(t *testing.T) {
		path := t.TempDir()
		for _, tc := range []struct {
			questionToInterrupt string
		}{
			{questionToInterrupt: "Name"},
			{questionToInterrupt: "Description"},
			{questionToInterrupt: "AppFile"},
			{questionToInterrupt: "RequirementsFile"},
			{questionToInterrupt: useFolderQuestion(path)},
		} {
			t.Run(tc.questionToInterrupt, func(t *testing.T) {
				stubAsker := StubAsker{
					useFolderQuestion(path): true,
					"Name":                  "App Name",
					"Description":           "App description",
					"LibraryName":           "Streamlit",
					"AppFile":               "app.py",
					"RequirementsFile":      "requirements.txt",
				}
				stubAsker[tc.questionToInterrupt] = terminal.InterruptErr

				actual := manifest.Manifest{Python: &manifest.Python{}}
				cont, err := RunInitAppWizard(&stubAsker, path, &actual)

				assert.False(t, cont)
				assert.NoError(t, err)
			})
		}
	})
}
