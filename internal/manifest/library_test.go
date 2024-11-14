package manifest

import (
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
)

func TestSupportedLibraries(t *testing.T) {
	expected := []Library{LibraryStreamlit, LibraryPlotlyDash, LibraryMarimo, LibraryPanel}
	assert.Equal(t, expected, SupportedLibraries, "unexpected supported libraries - remember to update tests!")
}

func TestGetLibraryByKey(t *testing.T) {
	testCases := []struct {
		key      string
		expected Library
	}{
		{key: "streamlit", expected: LibraryStreamlit},
		{key: "marimo", expected: LibraryMarimo},
		{key: "plotly", expected: LibraryPlotlyDash},
		{key: "panel", expected: LibraryPanel},
	}

	for _, testCase := range testCases {
		t.Run(testCase.key, func(t *testing.T) {
			actual, err := GetLibraryByKey(testCase.key)
			if assert.NoError(t, err) {
				assert.Equal(t, testCase.expected, actual)
			}
		})
	}

	t.Run("unsupported library error", func(t *testing.T) {
		_, err := GetLibraryByKey("unsupported")
		assert.ErrorIs(t, err, ErrUnsupportedLibrary)
	})
}

func TestGetLibraryByName(t *testing.T) {
	testCases := []struct {
		name     string
		expected Library
	}{
		{name: "Streamlit", expected: LibraryStreamlit},
		{name: "Marimo", expected: LibraryMarimo},
		{name: "Plotly-dash", expected: LibraryPlotlyDash},
		{name: "Panel", expected: LibraryPanel},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual, err := GetLibraryByName(testCase.name)
			if assert.NoError(t, err) {
				assert.Equal(t, testCase.expected, actual)
			}
		})
	}

	t.Run("unsupported library error", func(t *testing.T) {
		_, err := GetLibraryByName("unsupported")
		assert.ErrorIs(t, err, ErrUnsupportedLibrary)
	})
}

func TestDefaultAppFile(t *testing.T) {
	testCases := []struct {
		library  Library
		expected string
	}{
		{library: LibraryPanel, expected: panelApp},
	}

	for _, testCase := range testCases {
		actual := testCase.library.DefaultAppFile()
		assert.Equal(t, testCase.expected, actual)
	}
}

func TestUnmarshalLibrary(t *testing.T) {
	testCases := SupportedLibraries

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			type Container struct {
				Library Library
			}

			var c Container
			data := []byte("library = \"" + tc.Key + "\"")
			err := toml.Unmarshal(data, &c)

			assert.NoError(t, err)
			assert.Equal(t, Container{Library: tc}, c)
		})
	}
}
