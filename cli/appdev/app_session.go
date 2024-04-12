package appdev

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type AppSession struct {
	gorm.Model
	Title     string
	Name      string
	CreatedAt time.Time
	Elements  []AppSessionElement
	ClientID  string
}

var ErrAppSessionElementNotFound = errors.New("element not found")

func (s AppSession) GetElementByID(elementID string) (*AppSessionElement, error) {
	for _, e := range s.Elements {
		if strconv.FormatUint(uint64(e.ID), 10) == elementID {
			return &e, nil
		}
	}

	return nil, fmt.Errorf("no element with ID \"%s\"", elementID)
}

func (s AppSession) GetElementByPath(elementPath []string) (*AppSessionElement, error) {
	var element *AppSessionElement = nil
	for _, name := range elementPath {
		if e := s.getElementByNameAndParent(name, element); e == nil {
			return nil, ErrAppSessionElementNotFound
		} else {
			element = e
		}
	}

	if element != nil {
		return element, nil
	} else {
		return nil, ErrAppSessionElementNotFound
	}
}

func (s AppSession) getElementByNameAndParent(name string, parent *AppSessionElement) *AppSessionElement {
	for _, e := range s.Elements {
		if e.Name != name {
			continue
		}

		noParent := !e.ParentID.Valid && parent == nil
		if noParent {
			return &e
		}

		matchingParent := e.ParentID.Valid && parent != nil && e.ParentID.String == strconv.FormatUint(uint64(parent.ID), 10)
		if matchingParent {
			return &e
		}
	}

	return nil
}

type AppSessionElement struct {
	gorm.Model
	AppSessionID   uint
	ParentID       sql.NullString
	Name           string
	Label          string
	Type           string
	NumberValue    sql.NullFloat64
	StringValue    sql.NullString
	SliderValue    sql.NullFloat64
	HTMLValue      sql.NullString
	Elements       []AppSessionElement `gorm:"foreignKey:ParentID"`
	SliderMinValue sql.NullFloat64
	SliderMaxValue sql.NullFloat64
}

func (s AppSession) GetAllChildren() []AppSessionElement {
	children := make([]AppSessionElement, 0)

	for _, e := range s.Elements {
		children = append(children, e.GetAllChildren()...)
	}

	return children
}

func (s AppSessionElement) GetAllChildren() []AppSessionElement {
	children := make([]AppSessionElement, 0)

	for _, e := range s.Elements {
		e.ParentID = sql.NullString{Valid: true, String: strconv.FormatUint(uint64(s.ID), 10)}
		children = append(children, e)
		children = append(children, e.GetAllChildren()...)
	}

	return children
}

func (s AppSession) GetParentOf(e *AppSessionElement) *AppSessionElement {
	if !e.ParentID.Valid {
		return nil
	}

	for _, p := range s.Elements {
		if strconv.FormatUint(uint64(p.ID), 10) == e.ParentID.String {
			return &p
		}
	}

	return nil
}

func (s AppSession) GetPath(e AppSessionElement) []string {
	var path []string

	elem := &e
	for elem != nil {
		path = append([]string{elem.Name}, path...)
		elem = s.GetParentOf(elem)
	}

	return path
}
