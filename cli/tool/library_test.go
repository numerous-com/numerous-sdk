package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSupportedLibraries(t *testing.T) {
	assert.Equal(t, []Library{
		LibraryStreamlit, LibraryPlotlyDash, LibraryMarimo, LibraryNumerous,
	}, SupportedLibraries, "unexpected supported libraries - remember to update tests!")
}

func TestGetLibraryByKey(t *testing.T) {
	testCases := []struct {
		key      string
		expected Library
	}{
		{key: "streamlit", expected: LibraryStreamlit},
		{key: "numerous", expected: LibraryNumerous},
		{key: "marimo", expected: LibraryMarimo},
		{key: "plotly", expected: LibraryPlotlyDash},
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
		if assert.Error(t, err) {
			assert.Equal(t, "\"unsupported\" is not a valid app library. \nThe valid options are: streamlit, plotly, marimo, and numerous", err.Error())
		}
	})
}

func TestGetLibraryByName(t *testing.T) {
	testCases := []struct {
		name     string
		expected Library
	}{
		{name: "Streamlit", expected: LibraryStreamlit},
		{name: "Numerous", expected: LibraryNumerous},
		{name: "Marimo", expected: LibraryMarimo},
		{name: "Plotly-dash", expected: LibraryPlotlyDash},
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
		if assert.Error(t, err) {
			assert.Equal(t, "no library named 'unsupported'", err.Error())
		}
	})
}

func TestDefaultAppFile(t *testing.T) {
	testCases := []struct {
		library  Library
		expected string
	}{
		{library: LibraryNumerous, expected: numerousApp},
	}

	for _, testCase := range testCases {
		actual := testCase.library.DefaultAppFile()
		assert.Equal(t, testCase.expected, actual)
	}
}
