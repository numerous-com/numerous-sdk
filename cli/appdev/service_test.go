package appdev

import (
	"context"
	"database/sql"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func convertID(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}

type AddElementTestCase struct {
	name                   string
	def                    AppDefinition
	added                  AppSessionElement
	expectedEvents         []AppSessionEvent
	expectedUpdatedSession AppSession
}

var addElementTestCases = []AddElementTestCase{
	{
		name:  "add number element",
		def:   AppDefinition{},
		added: AppSessionElement{Name: "elem", Type: "number", NumberValue: sql.NullFloat64{Valid: true, Float64: 10.0}},
		expectedEvents: []AppSessionEvent{
			{
				AppSessionID:   "0",
				SourceClientID: "addingClient",
				AddedElement: &AppSessionElement{
					Name:        "elem",
					Type:        "number",
					NumberValue: sql.NullFloat64{Valid: true, Float64: 10.0},
				},
			},
		},
		expectedUpdatedSession: AppSession{
			Elements: []AppSessionElement{
				{
					Model:       gorm.Model{ID: 0},
					Name:        "elem",
					Type:        "number",
					NumberValue: sql.NullFloat64{Valid: true, Float64: 10.0},
				},
			},
		},
	},
	{
		name: "add element to container",
		def: AppDefinition{
			Elements: []AppDefinitionElement{
				{
					Name: "container_element",
					Type: "container", Elements: []AppDefinitionElement{},
				},
			},
		},
		added: AppSessionElement{
			ParentID:    sql.NullString{Valid: true, String: "0"},
			Name:        "added_element",
			Type:        "number",
			NumberValue: sql.NullFloat64{Valid: true, Float64: 10.0},
		},
		expectedEvents: []AppSessionEvent{
			{
				AppSessionID:   "0",
				SourceClientID: "addingClient",
				AddedElement: &AppSessionElement{
					Model:       gorm.Model{ID: 1},
					ParentID:    sql.NullString{Valid: true, String: "0"},
					Name:        "added_element",
					Type:        "number",
					NumberValue: sql.NullFloat64{Valid: true, Float64: 10.0},
				},
			},
		},
		expectedUpdatedSession: AppSession{
			Elements: []AppSessionElement{
				{
					Model:    gorm.Model{ID: 0},
					Name:     "container_element",
					Type:     "container",
					Elements: []AppSessionElement{},
				},
				{
					Model:       gorm.Model{ID: 1},
					Name:        "added_element",
					Type:        "number",
					NumberValue: sql.NullFloat64{Valid: true, Float64: 10.0},
					ParentID:    sql.NullString{Valid: true, String: "0"},
				},
			},
		},
	},
	{
		name: "add container with child",
		def:  AppDefinition{Elements: []AppDefinitionElement{}},
		added: AppSessionElement{
			Name: "added_container",
			Type: "container",
			Elements: []AppSessionElement{
				{
					Model:       gorm.Model{ID: 1},
					Name:        "added_child",
					Type:        "number",
					NumberValue: sql.NullFloat64{Valid: true, Float64: 10.0},
				},
			},
		},
		expectedEvents: []AppSessionEvent{
			{
				AppSessionID:   "0",
				SourceClientID: "addingClient",
				AddedElement: &AppSessionElement{
					Model: gorm.Model{ID: 0},
					Name:  "added_container",
					Type:  "container",
					Elements: []AppSessionElement{
						{
							Model:       gorm.Model{ID: 1},
							Name:        "added_child",
							Type:        "number",
							NumberValue: sql.NullFloat64{Valid: true, Float64: 10.0},
							ParentID:    sql.NullString{Valid: true, String: "0"},
						},
					},
				},
			},
			{
				AppSessionID:   "0",
				SourceClientID: "addingClient",
				AddedElement: &AppSessionElement{
					Model:       gorm.Model{ID: 1},
					ParentID:    sql.NullString{Valid: true, String: "0"},
					Name:        "added_child",
					Type:        "number",
					NumberValue: sql.NullFloat64{Valid: true, Float64: 10.0},
				},
			},
		},
		expectedUpdatedSession: AppSession{
			Elements: []AppSessionElement{
				{
					Model: gorm.Model{ID: 0},
					Name:  "added_container",
					Type:  "container",
					Elements: []AppSessionElement{
						{
							Model:       gorm.Model{ID: 1},
							Name:        "added_child",
							Type:        "number",
							NumberValue: sql.NullFloat64{Valid: true, Float64: 10.0},
							ParentID:    sql.NullString{Valid: true, String: "0"},
						},
					},
				},
				{
					Model:       gorm.Model{ID: 1},
					Name:        "added_child",
					Type:        "number",
					NumberValue: sql.NullFloat64{Valid: true, Float64: 10.0},
					ParentID:    sql.NullString{Valid: true, String: "0"},
				},
			},
		},
	},
}

func TestAddElementReturnsSession(t *testing.T) {
	for _, testcase := range addElementTestCases {
		repo := NewMockAppSessionRepository()
		service := NewAppSessionService(repo)
		_, err := repo.Create(testcase.def)
		assert.NoError(t, err)

		updatedSession, err := service.AddElement("addingClient", testcase.added)
		assert.NoError(t, err)

		assert.Equal(t, testcase.expectedUpdatedSession, *updatedSession)
	}
}

func TestAddElementEmitsEvent(t *testing.T) {
	for _, testcase := range addElementTestCases {
		t.Run(testcase.name, func(t *testing.T) {
			repo := NewMockAppSessionRepository()
			service := NewAppSessionService(repo)
			session, err := repo.Create(testcase.def)
			assert.NoError(t, err)
			appSessionID := strconv.FormatUint(uint64(session.ID), 10)

			ctx, cancel := context.WithCancel(context.Background())
			subscription, err := service.Subscribe(ctx, appSessionID, "subscribingClient")
			assert.NoError(t, err)
			time.Sleep(time.Microsecond * 100) // wait for subscription to start listening

			_, err = service.AddElement("addingClient", testcase.added)
			assert.NoError(t, err)

			for _, expectedEvent := range testcase.expectedEvents {
				select {
				case ev := <-subscription:
					println(ev.AddedElement.Name)
					assert.Equal(t, expectedEvent, ev)
				case <-time.After(time.Millisecond * 100):
					t.Error("timed out waiting for subscription event")
				}
			}
			cancel()
		})
	}
}

type RemoveElementReturnsSessionTestCase struct {
	name                   string
	def                    AppDefinition
	removeElementPath      []string
	expectedUpdatedSession AppSession
	expectedRemovedElement AppSessionElement
}

var removeElementTestCases = []RemoveElementReturnsSessionTestCase{
	{
		name: "root element",
		def: AppDefinition{
			Elements: []AppDefinitionElement{
				{Name: "text_element", Type: "string", Default: "default"},
			},
		},
		removeElementPath:      []string{"text_element"},
		expectedUpdatedSession: AppSession{Elements: []AppSessionElement{}},
		expectedRemovedElement: AppSessionElement{
			Model:       gorm.Model{ID: 0},
			Name:        "text_element",
			Type:        "string",
			StringValue: sql.NullString{Valid: true, String: "default"},
		},
	},
	{
		name: "nested element",
		def: AppDefinition{
			Elements: []AppDefinitionElement{
				{
					Name: "container_element",
					Type: "container",
					Elements: []AppDefinitionElement{
						{Name: "text_element", Type: "string", Default: "default"},
					},
				},
			},
		},
		removeElementPath: []string{"container_element", "text_element"},
		expectedUpdatedSession: AppSession{Elements: []AppSessionElement{
			{
				Model:    gorm.Model{ID: 0},
				Name:     "container_element",
				Type:     "container",
				Elements: []AppSessionElement{},
			},
		}},
		expectedRemovedElement: AppSessionElement{
			Model:       gorm.Model{ID: 1},
			Name:        "text_element",
			Type:        "string",
			StringValue: sql.NullString{Valid: true, String: "default"},
			ParentID:    sql.NullString{Valid: true, String: "0"},
		},
	},
}

func TestRemoveElementReturnsSession(t *testing.T) {
	for _, testcase := range removeElementTestCases {
		t.Run(testcase.name, func(t *testing.T) {
			repo := NewMockAppSessionRepository()
			service := NewAppSessionService(repo)
			s, err := repo.Create(testcase.def)
			assert.NoError(t, err)
			toRemove, err := s.GetElementByPath(testcase.removeElementPath)
			assert.NoError(t, err)

			if assert.NotNil(t, toRemove) {
				updatedSession, err := service.RemoveElement("removingClient", *toRemove)
				assert.NoError(t, err)
				assert.Equal(t, &testcase.expectedUpdatedSession, updatedSession)
			}
		})
	}
}

func TestRemoveElementEmitsEvent(t *testing.T) {
	for _, testcase := range removeElementTestCases {
		t.Run(testcase.name, func(t *testing.T) {
			repo := NewMockAppSessionRepository()
			service := NewAppSessionService(repo)
			session, err := repo.Create(testcase.def)
			assert.NoError(t, err)
			toRemove, err := session.GetElementByPath(testcase.removeElementPath)
			assert.NoError(t, err)
			appSessionID := strconv.FormatUint(uint64(session.ID), 10)

			ctx, cancel := context.WithCancel(context.Background())
			subscription, err := service.Subscribe(ctx, appSessionID, "subscribingClient")
			assert.NoError(t, err)
			time.Sleep(time.Microsecond * 100) // wait for subscription to start listening

			if !assert.NotNil(t, toRemove) {
				cancel()
				return
			}

			_, err = service.RemoveElement("removingClient", *toRemove)
			assert.NoError(t, err)

			select {
			case ev := <-subscription:
				assert.Equal(t, testcase.expectedRemovedElement, *ev.RemovedElement)
				cancel()
			case <-time.After(time.Second):
				t.Error("timed out waiting for subscription event")
				cancel()
			}
		})
	}
}

type TestUpdateElementTestCase struct {
	name                  string
	definition            AppDefinition
	elementUpdate         AppSessionElementUpdate
	expectedResultElement AppSessionElement
	elementPath           []string
}

var (
	updateElementStringValue = "updated text"
	updateElementNumberValue = 22.22
	updateElementHTMLValue   = "<p>updated html value</p>"
	updateElementSliderValue = 33.44
	updateElementTestCases   = []TestUpdateElementTestCase{
		{
			name: "string element",
			definition: AppDefinition{
				Name: "App Name",
				Elements: []AppDefinitionElement{
					{Name: "Text", Type: "string", Default: "default text"},
				},
			},
			elementPath:   []string{"Text"},
			elementUpdate: AppSessionElementUpdate{StringValue: &updateElementStringValue},
			expectedResultElement: AppSessionElement{
				Name:        "Text",
				Type:        "string",
				StringValue: sql.NullString{Valid: true, String: updateElementStringValue},
			},
		},
		{
			name: "number element",
			definition: AppDefinition{
				Name: "App Name",
				Elements: []AppDefinitionElement{
					{Name: "Number", Type: "number", Default: 11.11},
				},
			},
			elementPath:   []string{"Number"},
			elementUpdate: AppSessionElementUpdate{NumberValue: &updateElementNumberValue},
			expectedResultElement: AppSessionElement{
				Name:        "Number",
				Type:        "number",
				NumberValue: sql.NullFloat64{Valid: true, Float64: updateElementNumberValue},
			},
		},
		{
			name: "html element",
			definition: AppDefinition{
				Name: "App Name",
				Elements: []AppDefinitionElement{
					{Name: "Html", Type: "html", Default: "<p>updated html value</p>"},
				},
			},
			elementPath:   []string{"Html"},
			elementUpdate: AppSessionElementUpdate{HTMLValue: &updateElementHTMLValue},
			expectedResultElement: AppSessionElement{
				Name:      "Html",
				Type:      "html",
				HTMLValue: sql.NullString{Valid: true, String: updateElementHTMLValue},
			},
		},
		{
			name: "slider element",
			definition: AppDefinition{
				Name: "App Name",
				Elements: []AppDefinitionElement{
					{Name: "Slider", Type: "slider", Default: 11.22, SliderMinValue: -10.0, SliderMaxValue: 100.0},
				},
			},
			elementPath:   []string{"Slider"},
			elementUpdate: AppSessionElementUpdate{SliderValue: &updateElementSliderValue},
			expectedResultElement: AppSessionElement{
				Name:           "Slider",
				Type:           "slider",
				SliderValue:    sql.NullFloat64{Valid: true, Float64: updateElementSliderValue},
				SliderMinValue: sql.NullFloat64{Valid: true, Float64: -10.0},
				SliderMaxValue: sql.NullFloat64{Valid: true, Float64: 100.0},
			},
		},
		{
			name: "element in container",
			definition: AppDefinition{
				Name: "App Name",
				Elements: []AppDefinitionElement{
					{Name: "Container", Type: "container", Elements: []AppDefinitionElement{
						{Name: "Text", Type: "string", Default: "default text"},
					}},
				},
			},
			elementPath:   []string{"Container", "Text"},
			elementUpdate: AppSessionElementUpdate{StringValue: &updateElementStringValue},
			expectedResultElement: AppSessionElement{
				Name:        "Text",
				Type:        "string",
				StringValue: sql.NullString{Valid: true, String: updateElementStringValue},
			},
		},
		{
			name: "element in nested container",
			definition: AppDefinition{
				Name: "App Name",
				Elements: []AppDefinitionElement{
					{Name: "Container1", Type: "container", Elements: []AppDefinitionElement{
						{Name: "Container2", Type: "container", Elements: []AppDefinitionElement{
							{Name: "Text", Type: "string", Default: "default text"},
						}},
					}},
				},
			},
			elementPath:   []string{"Container1", "Container2", "Text"},
			elementUpdate: AppSessionElementUpdate{StringValue: &updateElementStringValue},
			expectedResultElement: AppSessionElement{
				Name:        "Text",
				Type:        "string",
				StringValue: sql.NullString{Valid: true, String: updateElementStringValue},
			},
		},
	}
)

func TestUpdateElementReturnsResultElement(t *testing.T) {
	for _, testcase := range updateElementTestCases {
		t.Run(testcase.name, func(t *testing.T) {
			repo := NewMockAppSessionRepository()
			service := NewAppSessionService(repo)

			session, err := repo.Create(testcase.definition)
			assert.NoError(t, err)

			element, err := session.GetElementByPath(testcase.elementPath)
			assert.NoError(t, err)

			testcase.expectedResultElement.ID = element.ID
			testcase.expectedResultElement.ParentID = element.ParentID
			testcase.elementUpdate.ElementID = convertID(element.ID)
			result, err := service.UpdateElement(session.ID, "updatingClient", testcase.elementUpdate)
			assert.NoError(t, err)
			assert.Equal(t, testcase.expectedResultElement, *result.Element)
		})
	}
}

func TestUpdateElementSendsEvent(t *testing.T) {
	for _, testcase := range updateElementTestCases {
		t.Run(testcase.name, func(t *testing.T) {
			repo := NewMockAppSessionRepository()
			service := NewAppSessionService(repo)

			session, err := repo.Create(testcase.definition)
			assert.NoError(t, err)

			element, err := session.GetElementByPath(testcase.elementPath)
			assert.NoError(t, err)

			ctx, cancel := context.WithCancel(context.Background())
			subscription, err := service.Subscribe(ctx, convertID(element.AppSessionID), "subscribingClient")
			assert.NoError(t, err)
			time.Sleep(time.Millisecond)

			testcase.expectedResultElement.ID = element.ID
			testcase.expectedResultElement.ParentID = element.ParentID
			testcase.elementUpdate.ElementID = convertID(element.ID)
			_, err = service.UpdateElement(session.ID, "updatingClient", testcase.elementUpdate)
			assert.NoError(t, err)

			select {
			case ev := <-subscription:
				assert.Equal(t, testcase.expectedResultElement, *ev.UpdatedElement)
				cancel()
			case <-time.After(time.Second):
				t.Error("timed out waiting for subscription event")
				cancel()
			}
		})
	}
}
