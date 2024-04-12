package appdev

import (
	"errors"
	"log/slog"
	"strconv"
)

// Represents the difference between an existing AppSession, and a new
// AppDefinition.
type AppSessionDifference struct {
	Added   []AppSessionElement
	Removed []AppSessionElement
	Updated []AppSessionElement
}

// Returns the difference between the elements in the provided existing app
// session, and the provided new app definition.
//
// Elements are matched by their paths within the app, where a path is a list of
// element names, ending with the elements own name, preceded by all the parent
// elements' names.
func GetAppSessionDifference(existing AppSession, newDef AppDefinition) AppSessionDifference {
	return AppSessionDifference{
		Removed: getRemovedElements(existing, newDef),
		Added:   getAddedElementsFromTool(existing, newDef),
		Updated: getUpdatedElementsFromTool(existing, newDef),
	}
}

func getRemovedElements(session AppSession, newDef AppDefinition) []AppSessionElement {
	var removed []AppSessionElement

	for _, sess := range session.Elements {
		path := session.GetPath(sess)
		_, err := newDef.GetElementByPath(path)
		if err != nil {
			removed = append(removed, sess)
		}
	}

	return removed
}

func getAddedElementsFromTool(session AppSession, newDef AppDefinition) []AppSessionElement {
	var added []AppSessionElement

	for _, newDef := range newDef.Elements {
		added = append(added, getAddedElementsFromElement(session, newDef)...)
	}

	return added
}

func getAddedElementsFromElement(session AppSession, newDef AppDefinitionElement) []AppSessionElement {
	var added []AppSessionElement

	p := newDef.GetPath()
	_, err := session.GetElementByPath(p)

	if errors.Is(err, ErrAppSessionElementNotFound) {
		newElem, err := newDef.CreateSessionElement()
		if err != nil {
			slog.Info("could not create new added element", slog.Any("error", err))
			return added
		}

		if newDef.Parent != nil {
			if parent, err := session.GetElementByPath(newDef.Parent.GetPath()); err == nil {
				newElem.ParentID.Valid = true
				newElem.ParentID.String = strconv.FormatUint(uint64(parent.ID), 10)
			}
		}

		newElem.AppSessionID = session.ID

		return append(added, *newElem)
	} else {
		for _, newChild := range newDef.Elements {
			added = append(added, getAddedElementsFromElement(session, newChild)...)
		}

		return added
	}
}

func getUpdatedElementsFromTool(session AppSession, newDefApp AppDefinition) []AppSessionElement {
	var updated []AppSessionElement

	for _, newDefElem := range newDefApp.Elements {
		updated = append(updated, getUpdatedElementsFromElement(session, newDefElem)...)
	}

	return updated
}

func getUpdatedElementsFromElement(session AppSession, newDefElem AppDefinitionElement) []AppSessionElement {
	var updated []AppSessionElement

	p := newDefElem.GetPath()
	sessionElement, err := session.GetElementByPath(p)

	if err == nil {
		if sessionElement.Label != newDefElem.Label {
			sessionElement.Label = newDefElem.Label
			updated = append(updated, *sessionElement)
		}
		for _, newChild := range newDefElem.Elements {
			updated = append(updated, getUpdatedElementsFromElement(session, newChild)...)
		}
	}

	return updated
}
