package appdev

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func compareToolSession(t *testing.T, s *AppSession) {
	t.Helper()
	assert.Equal(t, "Tool", s.Name)

	expectedTextElement := AppSessionElement{
		Model:       gorm.Model{ID: 0},
		Name:        "text_element",
		Type:        "string",
		StringValue: sql.NullString{Valid: true, String: "default"},
	}

	expectedNumberElement := AppSessionElement{
		Model:       gorm.Model{ID: 1},
		Name:        "number_element",
		Type:        "number",
		NumberValue: sql.NullFloat64{Valid: true, Float64: 1.0},
	}

	assert.Equal(t, []AppSessionElement{expectedTextElement, expectedNumberElement}, s.Elements)
}

func compareSecondToolSession(t *testing.T, s *AppSession) {
	t.Helper()
	if s.Name != "SecondTool" {
		t.Fatalf("got tool name %s, want %s", s.Name, "Tool")
	}

	expectedTextElement := AppSessionElement{
		Model:       gorm.Model{ID: 2},
		Name:        "second_text_element",
		Type:        "string",
		StringValue: sql.NullString{Valid: true, String: "second default"},
	}
	expectedNumberElement := AppSessionElement{
		Model:       gorm.Model{ID: 3},
		Name:        "second_number_element",
		Type:        "number",
		NumberValue: sql.NullFloat64{Valid: true, Float64: 10.0},
	}

	assert.Equal(t, []AppSessionElement{expectedTextElement, expectedNumberElement}, s.Elements)
}

func TestToolSessionRepository(t *testing.T) {
	def := AppDefinition{
		Name: "Tool",
		Elements: []AppDefinitionElement{
			{Name: "text_element", Type: "string", Default: "default"},
			{Name: "number_element", Type: "number", Default: 1.0},
		},
	}

	secondDef := AppDefinition{
		Name: "SecondTool",
		Elements: []AppDefinitionElement{
			{Name: "second_text_element", Type: "string", Default: "second default"},
			{Name: "second_number_element", Type: "number", Default: 10.0},
		},
	}

	t.Run("Create returns new ToolSession", func(t *testing.T) {
		r := InMemoryAppSessionRepository{}

		s, err := r.Create(def)
		require.NoError(t, err)

		compareToolSession(t, s)
	})

	t.Run("Read returns error before Create", func(t *testing.T) {
		r := InMemoryAppSessionRepository{}

		if s, err := r.Read(0); err == nil {
			t.Fatalf("r.Read(0) = (%#v, nil), want (nil, error)", s)
		} else {
			require.EqualError(t, err, ErrSessionNotCreated.Error())
		}
	})

	t.Run("Read returns ToolSession after Create", func(t *testing.T) {
		r := InMemoryAppSessionRepository{}

		_, err := r.Create(def)
		require.NoError(t, err)

		s, err := r.Read(0)
		require.NoError(t, err)

		compareToolSession(t, s)
	})

	t.Run("Read returns same ToolSession no matter the ID", func(t *testing.T) {
		r := InMemoryAppSessionRepository{}

		_, err := r.Create(def)
		require.NoError(t, err)

		for i := uint(0); i < 1000; i += 100 {
			s, err := r.Read(i)
			require.NoError(t, err)
			compareToolSession(t, s)
		}
	})

	t.Run("Create overrides existing global ToolSession", func(t *testing.T) {
		r := InMemoryAppSessionRepository{}
		_, err := r.Create(def)
		require.NoError(t, err)

		_, err = r.Create(secondDef)
		require.NoError(t, err)
		s, err := r.Read(0)
		require.NoError(t, err)

		compareSecondToolSession(t, s)
	})

	t.Run("Delete panics", func(t *testing.T) {
		r := InMemoryAppSessionRepository{}

		require.Panics(t, func() {
			//nolint:errcheck
			r.Delete(0)
		})
	})

	t.Run("assigns ids to nested elements", func(t *testing.T) {
		r := InMemoryAppSessionRepository{}

		def := AppDefinition{
			Elements: []AppDefinitionElement{
				{
					Name: "root",
					Type: "container",
					Elements: []AppDefinitionElement{
						{
							Name:     "middle",
							Type:     "container",
							Elements: []AppDefinitionElement{{Name: "leaf", Type: "string", Default: "default"}},
						},
					},
				},
			},
		}
		expected := AppSession{
			Model: gorm.Model{ID: 0},
			Elements: []AppSessionElement{
				{
					Model: gorm.Model{ID: 0},
					Name:  "root",
					Type:  "container",
					Elements: []AppSessionElement{
						{
							Model: gorm.Model{ID: 1},
							Name:  "middle",
							Type:  "container",
							Elements: []AppSessionElement{
								{
									Model:       gorm.Model{ID: 2},
									Name:        "leaf",
									Type:        "string",
									StringValue: sql.NullString{Valid: true, String: "default"},
								},
							},
						},
					},
				},
				{
					Model: gorm.Model{ID: 1},
					Name:  "middle",
					Type:  "container",
					Elements: []AppSessionElement{
						{
							Model:       gorm.Model{ID: 2},
							Name:        "leaf",
							Type:        "string",
							StringValue: sql.NullString{Valid: true, String: "default"},
						},
					},
					ParentID: sql.NullString{Valid: true, String: "0"},
				},
				{
					Model:       gorm.Model{ID: 2},
					Name:        "leaf",
					Type:        "string",
					StringValue: sql.NullString{Valid: true, String: "default"},
					ParentID:    sql.NullString{Valid: true, String: "1"},
				},
			},
		}

		actual, err := r.Create(def)
		require.NoError(t, err)
		assert.Equal(t, expected, *actual)
	})
}

type addElementTestCase struct {
	name     string
	def      AppDefinition
	added    AppSessionElement
	expected AppSessionElement
}

var addTestCases = []addElementTestCase{
	{
		name: "element added to empty session gets id 0",
		def:  AppDefinition{},
		added: AppSessionElement{
			Name:        "text",
			Type:        "string",
			StringValue: sql.NullString{Valid: true, String: "value"},
		},
		expected: AppSessionElement{
			Model:       gorm.Model{ID: 0},
			Name:        "text",
			Type:        "string",
			StringValue: sql.NullString{Valid: true, String: "value"},
		},
	},
	{
		name: "element added to non-empty session gets expected id",
		def: AppDefinition{
			Elements: []AppDefinitionElement{
				{Name: "field1", Type: "string", Default: ""},
				{Name: "field2", Type: "string", Default: ""},
				{Name: "field3", Type: "string", Default: ""},
			},
		},
		added: AppSessionElement{
			Name:        "field4",
			Type:        "string",
			StringValue: sql.NullString{Valid: true, String: "value"},
		},
		expected: AppSessionElement{
			Model:       gorm.Model{ID: 3},
			Name:        "field4",
			Type:        "string",
			StringValue: sql.NullString{Valid: true, String: "value"},
		},
	},
	{
		name: "child element expected id",
		def: AppDefinition{
			Elements: []AppDefinitionElement{
				{Name: "container", Type: "container", Elements: []AppDefinitionElement{}},
			},
		},
		added: AppSessionElement{
			Name:        "child",
			Type:        "string",
			StringValue: sql.NullString{Valid: true, String: "value"},
			ParentID:    sql.NullString{Valid: true, String: "0"},
		},
		expected: AppSessionElement{
			Model:       gorm.Model{ID: 1},
			Name:        "child",
			Type:        "string",
			StringValue: sql.NullString{Valid: true, String: "value"},
			ParentID:    sql.NullString{Valid: true, String: "0"},
		},
	},
	{
		name: "added container and child have expected ids",
		def:  AppDefinition{Elements: []AppDefinitionElement{}},
		added: AppSessionElement{
			Name: "container",
			Type: "container",
			Elements: []AppSessionElement{{
				Name:        "child",
				Type:        "string",
				StringValue: sql.NullString{Valid: true, String: "value"},
			}},
		},
		expected: AppSessionElement{
			Model: gorm.Model{ID: 0},
			Name:  "container",
			Type:  "container",
			Elements: []AppSessionElement{{
				Model:       gorm.Model{ID: 1},
				Name:        "child",
				Type:        "string",
				StringValue: sql.NullString{Valid: true, String: "value"},
				ParentID:    sql.NullString{Valid: true, String: "0"},
			}},
		},
	},
}

func TestAddElement(t *testing.T) {
	for _, testcase := range addTestCases {
		r := InMemoryAppSessionRepository{}
		_, err := r.Create(testcase.def)
		require.NoError(t, err)

		actual, err := r.AddElement(testcase.added)

		require.NoError(t, err)
		assert.Equal(t, testcase.expected, *actual)
	}
}
