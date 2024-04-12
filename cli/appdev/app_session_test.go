package appdev

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestGetElementByID(t *testing.T) {
	session := AppSession{
		Elements: []AppSessionElement{
			{Model: gorm.Model{ID: 1}, Name: "Elem 1"},
			{Model: gorm.Model{ID: 2}, Name: "Elem 2"},
			{Model: gorm.Model{ID: 3}, Name: "Elem 3"},
			{Model: gorm.Model{ID: 4}, Name: "Elem 4"},
		},
	}

	t.Run("returns expected element", func(t *testing.T) {
		wantName := "Elem 2"

		gotElem, gotErr := session.GetElementByID("2")
		assert.NoError(t, gotErr)

		assert.NoError(t, gotErr)
		assert.Equal(t, wantName, gotElem.Name)
	})

	t.Run("returns error element does not exist", func(t *testing.T) {
		_, gotErr := session.GetElementByID("10")
		assert.EqualError(t, gotErr, "no element with ID \"10\"")
	})
}

func TestGetElementByPath(t *testing.T) {
	session := AppSession{
		Elements: []AppSessionElement{
			{Model: gorm.Model{ID: 1}, Name: "elem_1"},
			{Model: gorm.Model{ID: 2}, Name: "elem_2"},
			{Model: gorm.Model{ID: 3}, Name: "elem_3", ParentID: sql.NullString{String: "2", Valid: true}},
			{Model: gorm.Model{ID: 4}, Name: "elem_4", ParentID: sql.NullString{String: "3", Valid: true}},
		},
	}

	t.Run("get expected root level element", func(t *testing.T) {
		var wantID uint = 1
		gotElem, gotErr := session.GetElementByPath([]string{"elem_1"})

		assert.NoError(t, gotErr)
		if assert.NotNil(t, gotElem) {
			assert.Equal(t, wantID, gotElem.ID)
		}
	})

	t.Run("cannot get nested element at root", func(t *testing.T) {
		_, gotErr := session.GetElementByPath([]string{"elem_3"})
		assert.ErrorIs(t, gotErr, ErrAppSessionElementNotFound)
	})

	t.Run("returns expected child element", func(t *testing.T) {
		var wantID uint = 3
		gotElem, gotErr := session.GetElementByPath([]string{"elem_2", "elem_3"})

		assert.NoError(t, gotErr)
		assert.Equal(t, wantID, gotElem.ID)
	})

	t.Run("returns expected nested element", func(t *testing.T) {
		var wantID uint = 4
		gotElem, gotErr := session.GetElementByPath([]string{"elem_2", "elem_3", "elem_4"})

		assert.NoError(t, gotErr)
		assert.Equal(t, wantID, gotElem.ID)
	})

	t.Run("does not return child that does not exist", func(t *testing.T) {
		gotElem, gotErr := session.GetElementByPath([]string{"elem_2", "elem_3", "elem_4", "non_existing"})
		assert.ErrorIs(t, gotErr, ErrAppSessionElementNotFound)
		assert.Nil(t, gotElem)
	})
}
