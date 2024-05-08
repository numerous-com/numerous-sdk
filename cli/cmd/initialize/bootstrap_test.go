package initialize

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"numerous/cli/manifest"
	"numerous/cli/test"
	"numerous/cli/tool"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type filesNotReadError struct {
	files []string
}

func (f *filesNotReadError) Error() string {
	return "Could not read the following files:" + "\n[\n" + strings.Join(f.files, ",\n") + "\n]\n"
}

func allFilesExist(fileNames []string) error {
	filesNotRead := []string{}

	for _, fileName := range fileNames {
		_, err := os.Stat(fileName)
		if err != nil {
			filesNotRead = append(filesNotRead, fileName)
		}
	}

	if len(filesNotRead) == 0 {
		return nil
	}

	return &filesNotReadError{files: filesNotRead}
}

func TestBootstrapAllFiles(t *testing.T) {
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))
	lib, err := tool.GetLibraryByKey("streamlit")
	require.NoError(t, err)
	testTool := tool.Tool{
		Library:          lib,
		AppFile:          "app.py",
		RequirementsFile: "requirements.txt",
		CoverImage:       "cover_image.png",
	}
	expectedFiles := []string{
		".gitignore",
		manifest.ManifestFileName,
		testTool.AppFile,
		testTool.RequirementsFile,
		testTool.CoverImage,
	}

	err = bootstrapFiles(testTool, "some-id", tempDir)

	if assert.NoError(t, err) {
		err = allFilesExist(expectedFiles)
		assert.NoError(t, err)
	}
}

func TestBootstrapRequirementsFile(t *testing.T) {
	var (
		dummyRequirementsWithoutNewLine = strings.Join([]string{"some", "different", "dependencies=2.0.0"}, "\n")
		dummyRequirementsWithNewLine    = dummyRequirementsWithoutNewLine + "\n"
	)

	testCases := []struct {
		name                 string
		library              tool.Library
		initialRequirements  string
		expectedRequirements string
	}{
		{
			name:                 "plotly-dash without initial requirements",
			library:              tool.LibraryPlotlyDash,
			initialRequirements:  "",
			expectedRequirements: "dash\ngunicorn\n",
		},
		{
			name:                 "streamlit without initial requirements",
			library:              tool.LibraryStreamlit,
			initialRequirements:  "",
			expectedRequirements: "streamlit\n",
		},
		{
			name:                 "marimo without initial requirements",
			library:              tool.LibraryMarimo,
			initialRequirements:  "",
			expectedRequirements: "marimo\n",
		},
		{
			name:                 "numerous without initial requirements",
			library:              tool.LibraryNumerous,
			initialRequirements:  "",
			expectedRequirements: "numerous\n",
		},
		{
			name:                 "marimo with initial requirements with newline appends at end",
			library:              tool.LibraryMarimo,
			initialRequirements:  dummyRequirementsWithNewLine,
			expectedRequirements: dummyRequirementsWithNewLine + "marimo\n",
		},
		{
			name:                 "marimo with initial requirements without newline appends at end",
			library:              tool.LibraryMarimo,
			initialRequirements:  dummyRequirementsWithoutNewLine,
			expectedRequirements: dummyRequirementsWithNewLine + "marimo\n",
		},
		{
			name:                 "marimo with initial requirements and library is part of requirements, nothing changes",
			library:              tool.LibraryMarimo,
			initialRequirements:  "marimo\n" + dummyRequirementsWithNewLine,
			expectedRequirements: "marimo\n" + dummyRequirementsWithNewLine,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			tempDir := t.TempDir()
			require.NoError(t, os.Chdir(tempDir))
			testTool := tool.Tool{
				Library:          testCase.library,
				AppFile:          "app.py",
				RequirementsFile: "requirements.txt",
				CoverImage:       "cover_image.png",
			}
			if testCase.initialRequirements != "" {
				err := os.WriteFile(testTool.RequirementsFile, []byte(testCase.initialRequirements), 0o644)
				require.NoError(t, err)
			}

			err := bootstrapFiles(testTool, "some-id", tempDir)

			require.NoError(t, err)
			actualRequirements, err := os.ReadFile(testTool.RequirementsFile)
			require.NoError(t, err)
			assert.Equal(t, testCase.expectedRequirements, string(actualRequirements))
		})
	}
}

var expectedNumerousApp string = `
from numerous import action, app, slider


@app
class MyApp:
	count: float
	step: float = slider(min_value=0, max_value=10)

	@action
	def increment(self) -> None:
		self.count += self.step


appdef = MyApp
`

func TestBootstrapFiles(t *testing.T) {
	t.Run("bootstraps app file", func(t *testing.T) {
		testCases := []struct {
			name            string
			library         tool.Library
			expectedAppFile string
		}{
			{
				name:            "numerous",
				library:         tool.LibraryNumerous,
				expectedAppFile: expectedNumerousApp,
			},
			{
				name:            "streamlit",
				library:         tool.LibraryStreamlit,
				expectedAppFile: "",
			},
			{
				name:            "dash",
				library:         tool.LibraryPlotlyDash,
				expectedAppFile: "",
			},
			{
				name:            "marimo",
				library:         tool.LibraryMarimo,
				expectedAppFile: "",
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				require.NoError(t, os.Chdir(t.TempDir()))
				testTool := tool.Tool{
					Library:          testCase.library,
					AppFile:          "app.py",
					RequirementsFile: "requirements.txt",
					CoverImage:       "cover_image.png",
				}
				tempDir, err := os.Getwd()
				require.NoError(t, err)

				err = bootstrapFiles(testTool, "tool id", tempDir)

				require.NoError(t, err)
				appContent, err := os.ReadFile("app.py")
				require.NoError(t, err)
				assert.Equal(t, testCase.expectedAppFile, string(appContent))
			})
		}
	})
	t.Run("adds expected lines to existing .gitignore", func(t *testing.T) {
		tmpDir := t.TempDir()
		toolID := "tool-id"
		tool := tool.Tool{
			RequirementsFile: "requirements.txt",
			AppFile:          "app.py",
			CoverImage:       "conver_img.png",
		}
		initialGitIgnoreContent := "some/ignore/pattern\nanother-ignore-pattern"
		expectedGitIgnoreContent := initialGitIgnoreContent + "\n# added by numerous init\n\n.app_id.txt\n.env"
		gitignoreFilePath := filepath.Join(tmpDir, ".gitignore")
		test.WriteFile(t, gitignoreFilePath, []byte(initialGitIgnoreContent))

		err := bootstrapFiles(tool, toolID, tmpDir)

		assert.NoError(t, err)
		actualGitIgnoreContent, err := os.ReadFile(gitignoreFilePath)
		if assert.NoError(t, err) {
			assert.Equal(t, expectedGitIgnoreContent, string(actualGitIgnoreContent))
		}
	})

	t.Run("writes manifest with expected excludes", func(t *testing.T) {
		tmpDir := t.TempDir()
		toolID := "tool-id"
		tool := tool.Tool{
			RequirementsFile: "requirements.txt",
			AppFile:          "app.py",
			CoverImage:       "conver_img.png",
		}

		bootErr := bootstrapFiles(tool, toolID, tmpDir)
		manifest, manifestErr := manifest.LoadManifest(tmpDir + "/" + manifest.ManifestFileName)

		assert.NoError(t, bootErr)
		assert.NoError(t, manifestErr)
		expectedExclude := []string{"*venv", "venv*", ".git", ".env"}
		assert.Equal(t, expectedExclude, manifest.Exclude)
	})
}
