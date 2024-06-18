package appdev

import (
	"errors"
	"log/slog"
	"strconv"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrRemoveNonExistingElement error = errors.New("cannot remove element that does not exist")

type AppSessionRepository interface {
	Create(definition AppDefinition) (*AppSession, error)
	Delete(id uint) error
	Read(id uint) (*AppSession, error)
	UpdateElement(element AppSessionElement) error
	AddElement(element AppSessionElement) (*AppSessionElement, error)
	RemoveElement(element AppSessionElement) error
}

type DBAppSessionRepository struct {
	db *gorm.DB
}

func NewDBAppSessionRepository(db *gorm.DB) *DBAppSessionRepository {
	return &DBAppSessionRepository{db: db}
}

func (r *DBAppSessionRepository) Create(def AppDefinition) (*AppSession, error) {
	appSession := def.CreateSession()
	if err := r.db.Create(&appSession).First(&appSession).Error; err != nil {
		slog.Error(err.Error())
		return nil, errors.New("error creating app session")
	}
	slog.Debug("Created app session", slog.Any("app session", appSession))

	return &appSession, nil
}

func (r *DBAppSessionRepository) Delete(id uint) error {
	if err := r.db.Delete(&AppSession{}, id).Error; err != nil {
		slog.Error(err.Error())
		return err
	}

	return nil
}

func (r *DBAppSessionRepository) Read(id uint) (*AppSession, error) {
	appSession := AppSession{Model: gorm.Model{ID: id}}
	if err := r.db.Preload(clause.Associations).First(&appSession).Error; err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	slog.Debug("Read app session", slog.Any("appSession", appSession))

	return &appSession, nil
}

func (r *DBAppSessionRepository) UpdateElement(element AppSessionElement) error {
	if err := r.db.Save(&element).Error; err != nil {
		slog.Error(err.Error())
		return err
	}

	slog.Debug("Update app session element", slog.Any("elementID", element.ID))

	return nil
}

func (r *DBAppSessionRepository) AddElement(element AppSessionElement) (*AppSessionElement, error) {
	panic("not implemented")
}

func (r *DBAppSessionRepository) RemoveElement(element AppSessionElement) error {
	panic("not implemented")
}

type InMemoryAppSessionRepository struct {
	session       *AppSession
	nextElementID uint
}

func (r *InMemoryAppSessionRepository) getNewElementID() uint {
	newID := r.nextElementID
	r.nextElementID++

	return newID
}

func (r *InMemoryAppSessionRepository) Create(definition AppDefinition) (*AppSession, error) {
	session := definition.CreateSession()
	r.setSessionIDs(session.Elements)
	childElements := session.GetAllChildren()
	session.Elements = append(session.Elements, childElements...)
	r.session = &session

	return &session, nil
}

func (r *InMemoryAppSessionRepository) setSessionIDs(elements []AppSessionElement) {
	for i := 0; i < len(elements); i++ {
		e := &elements[i]
		e.Model.ID = r.getNewElementID()
		if e.Type == "container" {
			r.setSessionIDs(e.Elements)
		}
	}
}

var ErrSessionNotCreated error = errors.New("session not created yet")

func (r *InMemoryAppSessionRepository) Delete(id uint) error {
	panic("not implemented")
}

func (r *InMemoryAppSessionRepository) Read(id uint) (*AppSession, error) {
	if r.session == nil {
		return nil, ErrSessionNotCreated
	} else {
		return r.session, nil
	}
}

func (r *InMemoryAppSessionRepository) UpdateElement(element AppSessionElement) error {
	found := false
	var index int
	for i, v := range r.session.Elements {
		if v.ID == element.ID {
			index = i
			found = true
		}
	}

	if found {
		r.session.Elements[index] = element
		return nil
	} else {
		return errors.New("tool session element does not exist")
	}
}

func (r *InMemoryAppSessionRepository) AddElement(element AppSessionElement) (*AppSessionElement, error) {
	r.assignElementIds(&element)
	r.session.Elements = append(r.session.Elements, element)
	r.session.Elements = append(r.session.Elements, element.GetAllChildren()...)

	return &element, nil
}

func (r *InMemoryAppSessionRepository) assignElementIds(element *AppSessionElement) {
	element.Model.ID = r.getNewElementID()
	for i := 0; i < len(element.Elements); i++ {
		e := &element.Elements[i]
		e.ParentID.Valid = true
		e.ParentID.String = strconv.FormatUint(uint64(element.ID), 10)
		r.assignElementIds(e)
	}
}

func (r *InMemoryAppSessionRepository) RemoveElement(element AppSessionElement) error {
	newElements := make([]AppSessionElement, 0)
	found := false

	for _, v := range r.session.Elements {
		if v.ID == element.ID {
			found = true
		} else {
			newElements = append(newElements, v)
		}
	}

	if !found {
		return ErrRemoveNonExistingElement
	} else {
		r.session.Elements = newElements
		return nil
	}
}
