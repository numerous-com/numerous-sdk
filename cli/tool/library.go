package tool

import (
	"fmt"
)

type Library struct {
	Name         string
	Key          string
	Port         uint
	Requirements []string
}

var numerousApp = `
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

func (l *Library) DefaultAppFile() string {
	switch l.Key {
	case "numerous":
		return numerousApp
	default:
		return ""
	}
}

var (
	streamlitPort uint = 80
	plotyPort     uint = 8050
	marimoPort    uint = 8000
	numerousPort  uint = 7001
)

var (
	LibraryStreamlit  = Library{Name: "Streamlit", Key: "streamlit", Port: streamlitPort, Requirements: []string{"streamlit"}}
	LibraryPlotlyDash = Library{Name: "Plotly-dash", Key: "plotly", Port: plotyPort, Requirements: []string{"dash", "gunicorn"}}
	LibraryMarimo     = Library{Name: "Marimo", Key: "marimo", Port: marimoPort, Requirements: []string{"marimo"}}
	LibraryNumerous   = Library{Name: "Numerous", Key: "numerous", Port: numerousPort, Requirements: []string{"numerous"}}
)
var SupportedLibraries = []Library{LibraryStreamlit, LibraryPlotlyDash, LibraryMarimo, LibraryNumerous}

func GetLibraryByKey(key string) (Library, error) {
	for _, library := range SupportedLibraries {
		if library.Key == key {
			return library, nil
		}
	}

	return Library{}, unsupportedLibraryError(key)
}

func GetLibraryByName(name string) (Library, error) {
	for _, library := range SupportedLibraries {
		if library.Name == name {
			return library, nil
		}
	}

	return Library{}, fmt.Errorf("no library named '%s'", name)
}

func unsupportedLibraryError(l string) error {
	supportedList := SupportedLibraries[0].Key
	lastIndex := len(SupportedLibraries[1:]) - 1
	for index, lib := range SupportedLibraries[1:] {
		if index == lastIndex {
			supportedList = fmt.Sprintf("%s, and %s", supportedList, lib.Key)
		} else {
			supportedList = fmt.Sprintf("%s, %s", supportedList, lib.Key)
		}
	}

	return fmt.Errorf("\"%s\" is not a valid app library. \nThe valid options are: %s", l, supportedList)
}
