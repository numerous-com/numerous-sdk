package appdev

import (
	"errors"
	"fmt"
	"strconv"
)

type MockAppSessionRepository struct {
	toolSessions        map[uint]AppSession
	toolSessionElements map[uint]*AppSessionElement
	nextToolSessionID   uint
	nextElementID       uint
}

func NewMockAppSessionRepository() *MockAppSessionRepository {
	return &MockAppSessionRepository{
		toolSessions:        make(map[uint]AppSession),
		toolSessionElements: make(map[uint]*AppSessionElement),
	}
}

func (r *MockAppSessionRepository) Create(def AppDefinition) (*AppSession, error) {
	toolSession := def.CreateSession()
	toolSession.Model.ID = r.nextToolSessionID
	r.addElementIDs(toolSession.Elements)
	toolSession.Elements = append(toolSession.Elements, toolSession.GetAllChildren()...)
	r.toolSessions[r.nextToolSessionID] = toolSession
	r.nextToolSessionID++

	return &toolSession, nil
}

func (r *MockAppSessionRepository) addElementIDs(elements []AppSessionElement) {
	for i := 0; i < len(elements); i++ {
		element := &elements[i]
		r.assignElementID(element)
		r.toolSessionElements[element.ID] = element
		r.addElementIDs(element.Elements)
	}
}

func (r *MockAppSessionRepository) Delete(id uint) error {
	delete(r.toolSessions, id)
	return nil
}

func (r *MockAppSessionRepository) Read(id uint) (*AppSession, error) {
	if toolSession, ok := r.toolSessions[id]; ok {
		return &toolSession, nil
	} else {
		return nil, errors.New("tool session does not exist")
	}
}

func (r *MockAppSessionRepository) UpdateElement(element AppSessionElement) error {
	if elementReference, ok := r.toolSessionElements[element.ID]; ok {
		*elementReference = element
		return nil
	} else {
		return errors.New("tool session element does not exist")
	}
}

func (r *MockAppSessionRepository) AddElement(element AppSessionElement) (*AppSessionElement, error) {
	if session, ok := r.toolSessions[element.AppSessionID]; !ok {
		return nil, fmt.Errorf("cannot add element to tool session %d that does not exist", element.AppSessionID)
	} else {
		r.assignElementID(&element)
		session.Elements = append(session.Elements, element)
		session.Elements = append(session.Elements, r.addChildElements(&element)...)
		r.toolSessions[session.ID] = session

		return &element, nil
	}
}

func (r *MockAppSessionRepository) assignElementID(element *AppSessionElement) {
	element.Model.ID = r.nextElementID
	r.nextElementID++
}

func (r *MockAppSessionRepository) addChildElements(parent *AppSessionElement) []AppSessionElement {
	added := []AppSessionElement{}

	for i := 0; i < len(parent.Elements); i++ {
		e := &parent.Elements[i]
		e.ParentID.Valid = true
		e.ParentID.String = strconv.FormatUint(uint64(parent.ID), 10)
		r.assignElementID(e)
		added = append(added, *e)
		if e.Type == "container" {
			added = append(added, r.addChildElements(e)...)
		}
	}

	return added
}

func (r *MockAppSessionRepository) RemoveElement(element AppSessionElement) error {
	if session, ok := r.toolSessions[element.AppSessionID]; !ok {
		return fmt.Errorf("cannot remove element from tool session %d that does not exist", element.AppSessionID)
	} else {
		newElements, found := removeElement(element, session.Elements)

		if !found {
			return ErrRemoveNonExistingElement
		} else {
			session.Elements = newElements
			r.toolSessions[session.ID] = session

			return nil
		}
	}
}

func removeElement(removed AppSessionElement, elements []AppSessionElement) ([]AppSessionElement, bool) {
	newElements := make([]AppSessionElement, 0)
	found := false

	for _, v := range elements {
		childNewElements, childFound := removeElement(removed, v.Elements)
		v.Elements = childNewElements
		if childFound {
			found = true
		}

		if v.ID == removed.ID {
			found = true
		} else {
			newElements = append(newElements, v)
		}
	}

	return newElements, found
}
