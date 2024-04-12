package appdev

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAppDefinitionReturnsResultWithApp(t *testing.T) {
	containerDef := AppDefinitionElement{
		Name: "Container",
		Type: "container",
	}
	containerDef.Elements = []AppDefinitionElement{{Name: "Text", Type: "string", Default: "default text", Parent: &containerDef}}

	testCases := []struct {
		json     string
		expected ParseAppDefinitionResult
	}{
		{
			json:     `{"app": {"name": "App Name", "elements": []}}`,
			expected: ParseAppDefinitionResult{App: &AppDefinition{Name: "App Name", Elements: []AppDefinitionElement{}}},
		},
		{
			json: `{"app": {"name": "App Name", "elements": [{"name": "Text", "type": "string", "default": "default text"}]}}`,
			expected: ParseAppDefinitionResult{App: &AppDefinition{Name: "App Name", Elements: []AppDefinitionElement{
				{Name: "Text", Type: "string", Default: "default text"},
			}}},
		},
		{
			json: `{"app": {"name": "App Name", "elements": [{"name": "Number", "type": "number", "default": 12.3}]}}`,
			expected: ParseAppDefinitionResult{App: &AppDefinition{Name: "App Name", Elements: []AppDefinitionElement{
				{Name: "Number", Type: "number", Default: 12.3},
			}}},
		},
		{
			json: `{"app": {"name": "App Name", "elements": [{"name": "Slider", "type": "slider", "default": 1.2, "slider_min_value": 3.4, "slider_max_value": 5.6}]}}`,
			expected: ParseAppDefinitionResult{App: &AppDefinition{Name: "App Name", Elements: []AppDefinitionElement{
				{Name: "Slider", Type: "slider", Default: 1.2, SliderMinValue: 3.4, SliderMaxValue: 5.6},
			}}},
		},
		{
			json: `{"app": {"name": "App Name", "elements": [{"name": "Html", "type": "html", "default": "<p>default html</p>"}]}}`,
			expected: ParseAppDefinitionResult{App: &AppDefinition{Name: "App Name", Elements: []AppDefinitionElement{
				{Name: "Html", Type: "html", Default: "<p>default html</p>"},
			}}},
		},
		{
			json:     `{"app": {"name": "App Name", "elements": [{"name": "Container", "type": "container", "elements": [{"name": "Text", "type": "string", "default": "default text"}]}]}}`,
			expected: ParseAppDefinitionResult{App: &AppDefinition{Name: "App Name", Elements: []AppDefinitionElement{containerDef}}},
		},
		{
			json:     `{"app": {"name": "App Name", "elements": [{"name": "Container", "type": "container", "elements": [{"name": "Text", "type": "string", "default": "default text"}]}, {"name": "Sibling", "type": "action"}]}}`,
			expected: ParseAppDefinitionResult{App: &AppDefinition{Name: "App Name", Elements: []AppDefinitionElement{containerDef, {Name: "Sibling", Type: "action"}}}},
		},
	}

	for _, testCase := range testCases {
		result, err := ParseAppDefinition([]byte(testCase.json))
		assert.NoError(t, err)
		assert.Equal(t, testCase.expected, result)
	}
}

func TestParseAppDefinitionReturnsResultWithError(t *testing.T) {
	containerDef := AppDefinitionElement{
		Name: "Container",
		Type: "container",
	}
	containerDef.Elements = []AppDefinitionElement{{Name: "Text", Type: "string", Default: "default text", Parent: &containerDef}}

	testCases := []struct {
		json     string
		expected ParseAppDefinitionResult
	}{
		{
			json: `
				{
					"error": {
						"appnotfound": {
							"app": "MyApp",
							"found_apps": ["MyOtherApp"]
						}
					}
				}`,
			expected: ParseAppDefinitionResult{Error: &ParseAppDefinitionError{
				AppNotFound: &AppNotFoundError{
					App:       "MyApp",
					FoundApps: []string{"MyOtherApp"},
				},
			}},
		},
		{
			json: `
				{
					"error": {
						"modulenotfound": {
							"module": "somenotfoundmodule"
						}
					}
				}`,
			expected: ParseAppDefinitionResult{Error: &ParseAppDefinitionError{
				ModuleNotFound: &AppModuleNotFoundError{
					Module: "somenotfoundmodule",
				},
			}},
		},
		{
			json: `
				{
					"error": {
						"appsyntax": {
							"context": "def bla(\n      ^",
							"msg": "error description",
							"pos": {"line": 5, "offset": 7}
						}
					}
				}`,
			expected: ParseAppDefinitionResult{Error: &ParseAppDefinitionError{
				Syntax: &AppSyntaxError{
					Msg:     "error description",
					Context: "def bla(\n      ^",
					Pos:     AppCodeCoordinate{Line: 5, Offset: 7},
				},
			}},
		},
		{
			json: `
				{
					"error": {
						"unknown": {
							"typename": "SomeError",
							"traceback": "Traceback:\nFile: blabla\nFile: blabla\n"
						}
					}
				}`,
			expected: ParseAppDefinitionResult{
				Error: &ParseAppDefinitionError{
					Unknown: &AppUnknownError{
						Typename:  "SomeError",
						Traceback: "Traceback:\nFile: blabla\nFile: blabla\n",
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		result, err := ParseAppDefinition([]byte(testCase.json))
		assert.NoError(t, err)
		assert.Equal(t, testCase.expected, result)
	}
}

func TestParseAppDefinitionReturnsError(t *testing.T) {
	t.Run("invalid json returns error", func(t *testing.T) {
		result, err := ParseAppDefinition([]byte(`{jens: 123}`))
		assert.Equal(t, ParseAppDefinitionResult{}, result)
		assert.Error(t, err)
	})

	t.Run("unended json returns error", func(t *testing.T) {
		result, err := ParseAppDefinition([]byte(`{"name": "App", "elements": []`))
		assert.Equal(t, ParseAppDefinitionResult{}, result)
		assert.Error(t, err)
	})
}
