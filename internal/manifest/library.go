package manifest

import (
	"errors"
	"fmt"
)

const (
	streamlitPort uint = 80
	plotyPort     uint = 8050
	marimoPort    uint = 8000
	numerousPort  uint = 7001
	panelPort     uint = 5006
)

var ErrUnsupportedLibrary = errors.New("unsupported library")

var (
	LibraryStreamlit  = Library{Name: "Streamlit", Key: "streamlit", Port: streamlitPort, Requirements: []string{"streamlit"}}
	LibraryPlotlyDash = Library{Name: "Plotly-dash", Key: "plotly", Port: plotyPort, Requirements: []string{"dash", "gunicorn"}}
	LibraryMarimo     = Library{Name: "Marimo", Key: "marimo", Port: marimoPort, Requirements: []string{"marimo"}}
	LibraryPanel      = Library{Name: "Panel", Key: "panel", Port: panelPort, Requirements: []string{"panel"}}
	LibraryNumerous   = Library{Name: "Numerous", Key: "numerous", Port: numerousPort, Requirements: []string{"numerous"}}
)

type Library struct {
	Name         string
	Key          string
	Port         uint
	Requirements []string
}

const numerousApp = `from numerous import action, app, slider


@app
class MyApp:
	count: float
	step: float = slider(min_value=0, max_value=10)

	@action
	def increment(self) -> None:
		self.count += self.step


appdef = MyApp
`

const panelApp = `import panel as pn


pn.template.MaterialTemplate(
    site="Panel",
    title="Hello world app",
    main=[pn.pane.Markdown("# Hello, world!")],
).servable()
`

func (l *Library) MarshalText() ([]byte, error) {
	return []byte(l.Key), nil
}

func (l *Library) UnmarshalText(text []byte) error {
	parsed, err := GetLibraryByKey(string(text))
	if err != nil {
		return err
	}

	l.Key = parsed.Key
	l.Name = parsed.Name
	l.Port = parsed.Port
	l.Requirements = parsed.Requirements

	return nil
}

func (l *Library) DefaultAppFile() string {
	switch l.Key {
	case "numerous":
		return numerousApp
	case "panel":
		return panelApp
	default:
		return ""
	}
}

var SupportedLibraries = []Library{LibraryStreamlit, LibraryPlotlyDash, LibraryMarimo, LibraryPanel, LibraryNumerous}

func GetLibraryByKey(key string) (Library, error) {
	for _, library := range SupportedLibraries {
		if library.Key == key {
			return library, nil
		}
	}

	return Library{}, ErrUnsupportedLibrary
}

func GetLibraryByName(name string) (Library, error) {
	for _, library := range SupportedLibraries {
		if library.Name == name {
			return library, nil
		}
	}

	return Library{}, ErrUnsupportedLibrary
}

func SupportedLibraryValuesList() string {
	supportedList := SupportedLibraries[0].Key
	lastIndex := len(SupportedLibraries[1:]) - 1
	for index, lib := range SupportedLibraries[1:] {
		if index == lastIndex {
			supportedList = fmt.Sprintf("%s, and %s", supportedList, lib.Key)
		} else {
			supportedList = fmt.Sprintf("%s, %s", supportedList, lib.Key)
		}
	}

	return supportedList
}
