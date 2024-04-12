package appdev

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type CreateSessionElementTestCase struct {
	name     string
	def      AppDefinitionElement
	expected AppSessionElement
}

var createSessionElementTestCases = []CreateSessionElementTestCase{
	{
		name: "string field",
		def: AppDefinitionElement{
			Label:   "String Label",
			Type:    "string",
			Default: "default string value",
		},
		expected: AppSessionElement{
			Label:       "String Label",
			Type:        "string",
			StringValue: sql.NullString{Valid: true, String: "default string value"},
		},
	},
	{
		name: "number field",
		def: AppDefinitionElement{
			Label:   "Number Label",
			Type:    "number",
			Default: 10.0,
		},
		expected: AppSessionElement{
			Label:       "Number Label",
			Type:        "number",
			NumberValue: sql.NullFloat64{Valid: true, Float64: 10.0},
		},
	},
	{
		name: "slider field",
		def: AppDefinitionElement{
			Label:          "Slider label",
			Type:           "slider",
			Default:        10.0,
			SliderMinValue: 123.0,
			SliderMaxValue: 456.0,
		},
		expected: AppSessionElement{
			Label:          "Slider label",
			Type:           "slider",
			SliderValue:    sql.NullFloat64{Valid: true, Float64: 10.0},
			SliderMinValue: sql.NullFloat64{Valid: true, Float64: 123.0},
			SliderMaxValue: sql.NullFloat64{Valid: true, Float64: 456.0},
		},
	},
	{
		name: "html field",
		def: AppDefinitionElement{
			Label:   "HTML label",
			Type:    "html",
			Default: "<p>default html value</p>",
		},
		expected: AppSessionElement{
			Label:     "HTML label",
			Type:      "html",
			HTMLValue: sql.NullString{Valid: true, String: "<p>default html value</p>"},
		},
	},
	{
		name: "container with 1 child",
		def: AppDefinitionElement{
			Label: "Container Label",
			Type:  "container",
			Elements: []AppDefinitionElement{
				{Label: "string label", Type: "string", Default: "default string"},
			},
		},
		expected: AppSessionElement{
			Label: "Container Label",
			Type:  "container",
			Elements: []AppSessionElement{
				{Label: "string label", Type: "string", StringValue: sql.NullString{Valid: true, String: "default string"}},
			},
		},
	},
	{
		name: "nested container with 1 child",
		def: AppDefinitionElement{
			Label: "Container Label",
			Type:  "container",
			Elements: []AppDefinitionElement{
				{
					Label: "Nested container Label",
					Type:  "container",
					Elements: []AppDefinitionElement{
						{Type: "string", Default: "default string"},
					},
				},
			},
		},
		expected: AppSessionElement{
			Label: "Container Label",
			Type:  "container",
			Elements: []AppSessionElement{
				{
					Label: "Nested container Label",
					Type:  "container",
					Elements: []AppSessionElement{
						{Type: "string", StringValue: sql.NullString{Valid: true, String: "default string"}},
					},
				},
			},
		},
	},
}

func TestCreateSessionElement(t *testing.T) {
	for _, testcase := range createSessionElementTestCases {
		t.Run(testcase.name, func(t *testing.T) {
			sess, err := testcase.def.CreateSessionElement()

			assert.NoError(t, err)
			assert.Equal(t, testcase.expected, *sess)
		})
	}
}

func TestCreateSessionElementError(t *testing.T) {
	createErrorTestTestCases := []AppDefinitionElement{
		{Type: "string", Default: 123.45},
		{Type: "number", Default: "some string value"},
	}

	for _, definition := range createErrorTestTestCases {
		definition.Name = "element_name"
		testName := fmt.Sprintf("%s element definition with non-%s default returns error", definition.Type, definition.Type)
		t.Run(testName, func(t *testing.T) {
			sess, err := definition.CreateSessionElement()

			expectedError := fmt.Sprintf("parameter element_name of type %s has invalid default \"%v\"", definition.Type, definition.Default)
			assert.Nil(t, sess)
			assert.EqualError(t, err, expectedError)
		})
	}
}

type CreateAppSessionTestCase struct {
	name     string
	def      AppDefinition
	expected AppSession
}

var createAppSessionTestCases = []CreateAppSessionTestCase{
	{
		name: "empty app session",
		def: AppDefinition{
			Elements: []AppDefinitionElement{},
		},
		expected: AppSession{
			Elements: []AppSessionElement{},
		},
	},
	{
		name: "container with multiple children",
		def: AppDefinition{
			Name: "App",
			Elements: []AppDefinitionElement{
				{
					Name: "container_element",
					Type: "container",
					Elements: []AppDefinitionElement{
						{Name: "string_element", Type: "string", Default: "default value"},
						{Name: "number_element", Type: "number", Default: 123.45},
					},
				},
			},
		},
		expected: AppSession{
			Name: "App",
			Elements: []AppSessionElement{
				{
					Name: "container_element",
					Type: "container",
					Elements: []AppSessionElement{
						{
							Name:        "string_element",
							Type:        "string",
							StringValue: sql.NullString{Valid: true, String: "default value"},
						},
						{
							Name:        "number_element",
							Type:        "number",
							NumberValue: sql.NullFloat64{Valid: true, Float64: 123.45},
						},
					},
				},
			},
		},
	},
	{
		name: "sibling containers",
		def: AppDefinition{
			Elements: []AppDefinitionElement{
				{Name: "c1", Type: "container", Elements: []AppDefinitionElement{{Type: "string", Default: "default"}}},
				{Name: "c2", Type: "container", Elements: []AppDefinitionElement{{Type: "string", Default: "default"}}},
			},
		},
		expected: AppSession{
			Elements: []AppSessionElement{
				{
					Name: "c1",
					Type: "container",
					Elements: []AppSessionElement{
						{Type: "string", StringValue: sql.NullString{Valid: true, String: "default"}},
					},
				},
				{
					Name: "c2",
					Type: "container",
					Elements: []AppSessionElement{
						{Type: "string", StringValue: sql.NullString{Valid: true, String: "default"}},
					},
				},
			},
		},
	},
}

func TestCreateAppSessionReturnsExpected(t *testing.T) {
	for _, testcase := range createAppSessionTestCases {
		t.Run(testcase.name, func(t *testing.T) {
			s := testcase.def.CreateSession()
			assert.Equal(t, testcase.expected, s)
		})
	}
}

type GetElementByPathTestCase struct {
	name     string
	def      AppDefinition
	path     []string
	expected *AppDefinitionElement
}

func TestAppDefinitionGetElementByPath(t *testing.T) {
	testCases := []GetElementByPathTestCase{
		{
			name: "returns root element",
			def: AppDefinition{
				Elements: []AppDefinitionElement{
					{
						Name:    "elem",
						Type:    "string",
						Default: "default string",
					},
				},
			},
			path: []string{"elem"},
			expected: &AppDefinitionElement{
				Name:    "elem",
				Type:    "string",
				Default: "default string",
			},
		},
		{
			name: "returns nested element",
			def: AppDefinition{
				Elements: []AppDefinitionElement{
					{
						Name: "cont",
						Type: "container",
						Elements: []AppDefinitionElement{
							{
								Name:    "child",
								Type:    "number",
								Default: 1.2,
							},
						},
					},
				},
			},
			path: []string{"cont", "child"},
			expected: &AppDefinitionElement{
				Name:    "child",
				Type:    "number",
				Default: 1.2,
			},
		},
	}

	for _, testcase := range testCases {
		elem, err := testcase.def.GetElementByPath(testcase.path)
		assert.NoError(t, err)
		assert.Equal(t, testcase.expected, elem)
	}
}

type GetElementByPathErrorTestCase struct {
	name string
	def  AppDefinition
	path []string
	err  error
}

var errorTestCases = []GetElementByPathErrorTestCase{
	{
		name: "returns error for non-existing root element",
		def: AppDefinition{
			Elements: []AppDefinitionElement{
				{
					Name:    "elem",
					Type:    "number",
					Default: 1.2,
				},
			},
		},
		path: []string{"cont", "child"},
		err:  ErrAppDefinitionElementNotFound,
	},
	{
		name: "returns error for non-existing nested element",
		def: AppDefinition{
			Elements: []AppDefinitionElement{
				{
					Name: "cont",
					Type: "container",
					Elements: []AppDefinitionElement{
						{
							Name:    "child",
							Type:    "number",
							Default: 1.2,
						},
					},
				},
			},
		},
		path: []string{"cont", "non-existing"},
		err:  ErrAppDefinitionElementNotFound,
	},
}

func TestAppDefinitionGetElementByPathError(t *testing.T) {
	for _, testcase := range errorTestCases {
		t.Run(testcase.name, func(t *testing.T) {
			elem, err := testcase.def.GetElementByPath(testcase.path)
			assert.ErrorIs(t, err, ErrAppDefinitionElementNotFound)
			assert.Nil(t, elem)
		})
	}
}
