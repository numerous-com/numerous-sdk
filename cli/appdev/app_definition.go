package appdev

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
)

type AppDefinitionElement struct {
	Name           string                 `json:"name"`
	Label          string                 `json:"label"`
	Type           string                 `json:"type"`
	Default        any                    `json:"default"`
	SliderMinValue float64                `json:"slider_min_value"`
	SliderMaxValue float64                `json:"slider_max_value"`
	Elements       []AppDefinitionElement `json:"elements,omitempty"`
	Parent         *AppDefinitionElement  `json:"-"` // ignore in JSON
}

func (a AppDefinitionElement) setAsParentOnChildren() {
	for i := 0; i < len(a.Elements); i++ {
		e := &a.Elements[i]
		e.Parent = &a
		e.setAsParentOnChildren()
	}
}

func (a AppDefinitionElement) GetPath() []string {
	if a.Parent == nil {
		return []string{a.Name}
	} else {
		return append(a.Parent.GetPath(), a.Name)
	}
}

func (a AppDefinitionElement) String() string {
	elementsField := ""
	if a.Elements != nil {
		elements := make([]string, 0)
		for _, e := range a.Elements {
			elements = append(elements, e.String())
		}
		elementsField = fmt.Sprintf(", Elements: {%s}", strings.Join(elements, ", "))
	}

	return fmt.Sprintf("AppElementDefinition{Name: `%s`, Type: `%s`, Default: \"%v\"%s}", a.Name, a.Type, a.Default, elementsField)
}

func (a AppDefinitionElement) CreateSessionElement() (*AppSessionElement, error) {
	sessionElement := AppSessionElement{
		Name:  a.Name,
		Label: a.Label,
		Type:  a.Type,
	}

	switch a.Type {
	case "string":
		switch d := a.Default.(type) {
		case string:
			sessionElement.StringValue.Valid = true
			sessionElement.StringValue.String = d
		default:
			return nil, fmt.Errorf("parameter %s of type %s has invalid default \"%v\"", a.Name, a.Type, d)
		}
	case "number":
		switch d := a.Default.(type) {
		case float64:
			sessionElement.NumberValue.Valid = true
			sessionElement.NumberValue.Float64 = d
		default:
			return nil, fmt.Errorf("parameter %s of type %s has invalid default \"%v\"", a.Name, a.Type, d)
		}
	case "container":
		sessionElement.Elements = createSessionElements(a.Elements)
	case "action":
	case "html":
		switch d := a.Default.(type) {
		case string:
			sessionElement.HTMLValue.Valid = true
			sessionElement.HTMLValue.String = d
		default:
			return nil, fmt.Errorf("parameter %s of type %s has invalid default \"%v\"", a.Name, a.Type, d)
		}
	case "slider":
		switch d := a.Default.(type) {
		case float64:
			sessionElement.SliderValue.Valid = true
			sessionElement.SliderValue.Float64 = d
			sessionElement.SliderMinValue.Valid = true
			sessionElement.SliderMinValue.Float64 = a.SliderMinValue
			sessionElement.SliderMaxValue.Valid = true
			sessionElement.SliderMaxValue.Float64 = a.SliderMaxValue
		default:
			return nil, fmt.Errorf("parameter %s of type %s has invalid default \"%v\"", a.Name, a.Type, d)
		}
	default:
		return nil, fmt.Errorf("unexpected element type \"%s\"", a.Type)
	}

	return &sessionElement, nil
}

type AppDefinition struct {
	Title    string                 `json:"title"`
	Name     string                 `json:"name"`
	Elements []AppDefinitionElement `json:"elements"`
}

var ErrAppDefinitionElementNotFound = errors.New("element definition not found")

func (ad *AppDefinition) GetElementByPath(path []string) (*AppDefinitionElement, error) {
	var element *AppDefinitionElement
	elements := ad.Elements

	for _, name := range path {
		found := false
		for _, e := range elements {
			if name != e.Name {
				continue
			}

			found = true
			element = &e
			if e.Type == "container" {
				elements = e.Elements
			}
		}

		if !found {
			return nil, ErrAppDefinitionElementNotFound
		}
	}

	if element == nil {
		return nil, ErrAppDefinitionElementNotFound
	} else {
		return element, nil
	}
}

func (ad *AppDefinition) SetElementParents() {
	for _, e := range ad.Elements {
		e.setAsParentOnChildren()
	}
}

func (ad AppDefinition) CreateSession() AppSession {
	return AppSession{
		Title:    ad.Title,
		Name:     ad.Name,
		Elements: ad.CreateAppSessionElements(),
	}
}

func (ad AppDefinition) CreateAppSessionElements() []AppSessionElement {
	return createSessionElements(ad.Elements)
}

func createSessionElements(definitionElements []AppDefinitionElement) []AppSessionElement {
	elements := []AppSessionElement{}

	for _, def := range definitionElements {
		if e, err := def.CreateSessionElement(); err != nil {
			slog.Warn("could not create session element", slog.Any("error", err))
		} else {
			elements = append(elements, *e)
		}
	}

	return elements
}
