package appdev

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type differenceTestCase struct {
	name         string
	session      AppSession
	newDef       AppDefinition
	expectedDiff AppSessionDifference
}

func TestGetToolSessionDifference(t *testing.T) {
	testCases := []differenceTestCase{
		{
			name: "element removed",
			session: AppSession{
				Model: gorm.Model{ID: 0},
				Elements: []AppSessionElement{
					{
						Model:       gorm.Model{ID: 0},
						Name:        "number_element",
						Type:        "number",
						NumberValue: sql.NullFloat64{Valid: true, Float64: 10.0},
					},
					{
						Model:       gorm.Model{ID: 1},
						Name:        "string_element",
						Type:        "string",
						StringValue: sql.NullString{Valid: true, String: "some string value"},
					},
				},
			},
			newDef: AppDefinition{
				Elements: []AppDefinitionElement{
					{Name: "number_element", Type: "number", Default: 123.0},
				},
			},
			expectedDiff: AppSessionDifference{
				Removed: []AppSessionElement{
					{
						Model:       gorm.Model{ID: 1},
						Name:        "string_element",
						Type:        "string",
						StringValue: sql.NullString{Valid: true, String: "some string value"},
					},
				},
			},
		},
		{
			name: "element added",
			session: AppSession{
				Model: gorm.Model{ID: 0},
				Elements: []AppSessionElement{
					{
						Model:       gorm.Model{ID: 0},
						Name:        "number_element",
						Type:        "number",
						NumberValue: sql.NullFloat64{Valid: true, Float64: 10.0},
					},
				},
			},
			newDef: AppDefinition{
				Elements: []AppDefinitionElement{
					{Name: "number_element", Type: "number", Default: 123.0},
					{Name: "string_element", Type: "string", Default: "default string value"},
				},
			},
			expectedDiff: AppSessionDifference{
				Added: []AppSessionElement{
					{
						Name:        "string_element",
						Type:        "string",
						StringValue: sql.NullString{Valid: true, String: "default string value"},
					},
				},
			},
		},
		{
			name: "element label updated",
			session: AppSession{
				Model: gorm.Model{ID: 0},
				Elements: []AppSessionElement{
					{
						Model:       gorm.Model{ID: 0},
						Name:        "number_element",
						Label:       "Number Label",
						Type:        "number",
						NumberValue: sql.NullFloat64{Valid: true, Float64: 10.0},
					},
					{
						Model:       gorm.Model{ID: 1},
						Name:        "string_element",
						Label:       "String Label",
						Type:        "string",
						StringValue: sql.NullString{Valid: true, String: "default string value"},
					},
				},
			},
			newDef: AppDefinition{
				Elements: []AppDefinitionElement{
					{Name: "number_element", Type: "number", Label: "Number Label", Default: 10.0},
					{Name: "string_element", Type: "string", Label: "Updated String Label", Default: "default string value"},
				},
			},
			expectedDiff: AppSessionDifference{
				Updated: []AppSessionElement{
					{
						Model:       gorm.Model{ID: 1},
						Name:        "string_element",
						Type:        "string",
						Label:       "Updated String Label",
						StringValue: sql.NullString{Valid: true, String: "default string value"},
					},
				},
			},
		},
		{
			name: "nested element added",
			session: AppSession{
				Model: gorm.Model{ID: 0},
				Elements: []AppSessionElement{
					{
						Model: gorm.Model{ID: 5},
						Name:  "container_element",
						Type:  "container",
					},
					{
						Model:       gorm.Model{ID: 6},
						Name:        "number_element",
						Type:        "number",
						NumberValue: sql.NullFloat64{Valid: true, Float64: 10.0},
						ParentID:    sql.NullString{Valid: true, String: "5"},
					},
				},
			},
			newDef: AppDefinition{
				Elements: []AppDefinitionElement{
					{Name: "container_element", Type: "container", Elements: []AppDefinitionElement{
						{Name: "number_element", Type: "number", Default: 123.0},
						{Name: "string_element", Type: "string", Default: "default string value"},
					}},
				},
			},
			expectedDiff: AppSessionDifference{
				Added: []AppSessionElement{
					{
						Name:        "string_element",
						Type:        "string",
						StringValue: sql.NullString{Valid: true, String: "default string value"},
						ParentID:    sql.NullString{Valid: true, String: "5"},
					},
				},
			},
		},
		{
			name: "nested element label updated",
			session: AppSession{
				Model: gorm.Model{ID: 0},
				Elements: []AppSessionElement{
					{
						Model: gorm.Model{ID: 1},
						Name:  "container_element",
						Type:  "container",
					},
					{
						Model:    gorm.Model{ID: 2},
						Name:     "nested_container_element",
						Type:     "container",
						ParentID: sql.NullString{Valid: true, String: "1"},
					},
					{
						Model:       gorm.Model{ID: 3},
						Name:        "number_element",
						Label:       "Number Label",
						Type:        "number",
						NumberValue: sql.NullFloat64{Valid: true, Float64: 10.0},
						ParentID:    sql.NullString{Valid: true, String: "1"},
					},
					{
						Model:       gorm.Model{ID: 4},
						Name:        "string_element",
						Label:       "String Label",
						Type:        "string",
						StringValue: sql.NullString{Valid: true, String: "default string value"},
						ParentID:    sql.NullString{Valid: true, String: "2"},
					},
				},
			},
			newDef: AppDefinition{
				Elements: []AppDefinitionElement{
					{
						Name: "container_element", Type: "container", Elements: []AppDefinitionElement{
							{Name: "number_element", Label: "Updated Number Label", Type: "number", Default: 123.0},
							{
								Name: "nested_container_element", Type: "container", Elements: []AppDefinitionElement{
									{Name: "string_element", Label: "Updated String Label", Type: "string", Default: "default string value"},
								},
							},
						},
					},
				},
			},
			expectedDiff: AppSessionDifference{
				Updated: []AppSessionElement{
					{
						Model:       gorm.Model{ID: 3},
						Name:        "number_element",
						Label:       "Updated Number Label",
						Type:        "number",
						NumberValue: sql.NullFloat64{Valid: true, Float64: 10.0},
						ParentID:    sql.NullString{Valid: true, String: "1"},
					},
					{
						Model:       gorm.Model{ID: 4},
						Name:        "string_element",
						Label:       "Updated String Label",
						Type:        "string",
						StringValue: sql.NullString{Valid: true, String: "default string value"},
						ParentID:    sql.NullString{Valid: true, String: "2"},
					},
				},
			},
		},
		{
			name: "unchanged session with doubled child has empty diff",
			session: AppSession{
				Model: gorm.Model{
					ID: 0,
				},
				Name: "ContainerTool",
				Elements: []AppSessionElement{
					{
						Model: gorm.Model{ID: 0},
						Name:  "my_container",
						Type:  "container",
						Elements: []AppSessionElement{
							{
								Model:        gorm.Model{ID: 0},
								AppSessionID: 0,
								Name:         "child",
								Type:         "string",
								StringValue:  sql.NullString{String: "", Valid: true},
							},
						},
					},
					{
						Model:        gorm.Model{ID: 1},
						AppSessionID: 0,
						Name:         "print_child",
						Type:         "action",
					},
					{
						Model:        gorm.Model{ID: 2},
						AppSessionID: 0,
						ParentID:     sql.NullString{String: "0", Valid: true},
						Name:         "child",
						Type:         "string",
						StringValue:  sql.NullString{String: "", Valid: true},
					},
				},
			},
			newDef: AppDefinition{
				Name: "ContainerTool",
				Elements: []AppDefinitionElement{
					{
						Name: "my_container",
						Type: "container",
						Elements: []AppDefinitionElement{
							{
								Name:    "child",
								Type:    "string",
								Default: "",
							},
						},
					},
					{
						Name: "print_child",
						Type: "action",
					},
				},
			},
			expectedDiff: AppSessionDifference{},
		},
		{
			name: "container added diff is nested",
			session: AppSession{
				Model: gorm.Model{
					ID: 0,
				},
				Elements: []AppSessionElement{
					{
						Model:        gorm.Model{ID: 0},
						Name:         "action",
						Type:         "action",
						AppSessionID: 0,
					},
				},
			},
			newDef: AppDefinition{
				Elements: []AppDefinitionElement{
					{
						Name: "container",
						Type: "container",
						Elements: []AppDefinitionElement{
							{Name: "child", Type: "string", Default: "default"},
						},
					},
					{
						Name: "action",
						Type: "action",
					},
				},
			},
			expectedDiff: AppSessionDifference{
				Added: []AppSessionElement{
					{
						Name: "container",
						Type: "container",
						Elements: []AppSessionElement{
							{
								Name:        "child",
								Type:        "string",
								StringValue: sql.NullString{String: "default", Valid: true},
							},
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		name := testCase.name
		t.Run(name, func(t *testing.T) {
			testCase.newDef.SetElementParents()
			diff := GetAppSessionDifference(testCase.session, testCase.newDef)
			if !assert.Equal(t, testCase.expectedDiff, diff) {
				println()
			}
		})
	}
}
