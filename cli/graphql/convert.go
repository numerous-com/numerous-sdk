package graphql

import (
	"fmt"
	"log/slog"
	"strconv"

	"numerous/cli/appdev"
	"numerous/cli/graphql/model"

	"github.com/google/uuid"
)

func AppSessionFromDomain(session appdev.AppSession) *model.ToolSession {
	s := &model.ToolSession{
		ID:          convertID(session.ID),
		Title:       session.Title,
		Name:        session.Name,
		AllElements: AppSessionElementsFromDomain(session),
		IsActive:    true,
		ClientID:    uuid.NewString(),
	}

	return s
}

func AppSessionElementsFromDomain(session appdev.AppSession) []model.Element {
	var elements []model.Element
	for _, domainElement := range session.Elements {
		elem := AppSessionElementFromDomain(&session, domainElement)
		elements = append(elements, elem)
	}

	return elements
}

func AppSessionElementFromDomain(session *appdev.AppSession, elem appdev.AppSessionElement) model.Element {
	context := getGraphContext(session, elem)
	if element, err := getElementFromDomain(elem, &context); err != nil {
		slog.Error("Cannot retrieve element added from domain", slog.String("error", err.Error()))
		return nil
	} else {
		return element
	}
}

func getGraphContext(session *appdev.AppSession, domainElement appdev.AppSessionElement) model.ElementGraphContext {
	if !domainElement.ParentID.Valid {
		return model.ElementGraphContext{}
	}

	parentElement, err := session.GetElementByID(domainElement.ParentID.String)
	if err != nil {
		slog.Info("could not get graph parent, returning empty context", slog.Any("element", domainElement))
		return model.ElementGraphContext{}
	}

	if parent, err := getGraphParent(session, *parentElement); err != nil {
		slog.Info("error getting graph parent, returning empty context", slog.Any("element", domainElement), slog.Any("error", err))
		return model.ElementGraphContext{}
	} else {
		return model.ElementGraphContext{Parent: parent}
	}
}

func getGraphParent(session *appdev.AppSession, parentElement appdev.AppSessionElement) (model.ElementGraphParent, error) {
	switch parentElement.Type {
	case "container":
		switch c := AppSessionElementFromDomain(session, parentElement).(type) {
		case model.Container:
			return &c, nil
		default:
			return nil, fmt.Errorf("invalid parent %#v", c)
		}
	default:
		slog.Debug("invalid parent type for graph parent", slog.Any("session", session))
		return nil, fmt.Errorf("invalid parent type %s", parentElement.Type)
	}
}

func AppSessionEventFromDomain(session *appdev.AppSession, event appdev.AppSessionEvent) model.ToolSessionEvent {
	switch {
	case event.UpdatedElement != nil:
		return getElementUpdateFromDomain(session, *event.UpdatedElement)
	case event.TriggeredActionElement != nil:
		return getElementActionTriggerFromDomain(session, event.TriggeredActionElement)
	case event.AddedElement != nil:
		return getElementAddedFromDomain(session, event.AddedElement)
	case event.RemovedElement != nil:
		return getElementRemovedFromDomain(session, event.RemovedElement)
	default:
		slog.Warn("Unsupported tool session event", slog.Any("event", event))
		return nil
	}
}

func getElementFromDomain(domainElement appdev.AppSessionElement, context *model.ElementGraphContext) (model.Element, error) {
	switch domainElement.Type {
	case "string":
		return model.TextField{
			ID:           convertID(domainElement.ID),
			Name:         domainElement.Name,
			Label:        domainElement.Label,
			Value:        domainElement.StringValue.String,
			GraphContext: context,
		}, nil
	case "number":
		return model.NumberField{
			ID:           convertID(domainElement.ID),
			Name:         domainElement.Name,
			Label:        domainElement.Label,
			Value:        domainElement.NumberValue.Float64,
			GraphContext: context,
		}, nil
	case "action":
		return model.Button{
			ID:           convertID(domainElement.ID),
			Name:         domainElement.Name,
			Label:        domainElement.Label,
			GraphContext: context,
		}, nil
	case "container":
		return model.Container{
			ID:           convertID(domainElement.ID),
			Name:         domainElement.Name,
			Label:        domainElement.Label,
			GraphContext: context,
		}, nil
	case "slider":
		return model.SliderElement{
			ID:           convertID(domainElement.ID),
			Name:         domainElement.Name,
			Label:        domainElement.Label,
			GraphContext: context,
			Value:        domainElement.SliderValue.Float64,
			MinValue:     domainElement.SliderMinValue.Float64,
			MaxValue:     domainElement.SliderMaxValue.Float64,
		}, nil
	case "html":
		return model.HTMLElement{
			ID:           convertID(domainElement.ID),
			Name:         domainElement.Name,
			Label:        domainElement.Label,
			GraphContext: context,
			HTML:         domainElement.HTMLValue.String,
		}, nil
	default:
		slog.Debug("Invalid element type", slog.Any("type", domainElement.Type))
		return nil, fmt.Errorf("invalid element type %s", domainElement.Type)
	}
}

func getElementActionTriggerFromDomain(session *appdev.AppSession, actionTrigger *appdev.AppSessionElement) model.ToolSessionEvent {
	context := getGraphContext(session, *actionTrigger)
	return model.ToolSessionActionTriggered{
		Element: &model.Button{
			ID:           convertID(actionTrigger.ID),
			Label:        actionTrigger.Label,
			Name:         actionTrigger.Name,
			GraphContext: &context,
		},
	}
}

func getElementAddedFromDomain(session *appdev.AppSession, added *appdev.AppSessionElement) model.ToolSessionElementAdded {
	context := getGraphContext(session, *added)
	if element, err := getElementFromDomain(*added, &context); err != nil {
		slog.Error("Cannot retrieve element added from domain", slog.String("error", err.Error()))
		return model.ToolSessionElementAdded{
			Element: nil,
		}
	} else {
		return model.ToolSessionElementAdded{
			Element: element,
		}
	}
}

func getElementRemovedFromDomain(session *appdev.AppSession, removed *appdev.AppSessionElement) model.ToolSessionElementRemoved {
	context := getGraphContext(session, *removed)
	if element, err := getElementFromDomain(*removed, &context); err != nil {
		slog.Error("Cannot retrieve element removed from domain", slog.String("error", err.Error()))
		return model.ToolSessionElementRemoved{
			Element: nil,
		}
	} else {
		return model.ToolSessionElementRemoved{
			Element: element,
		}
	}
}

func getElementUpdateFromDomain(session *appdev.AppSession, update appdev.AppSessionElement) model.ToolSessionEvent {
	context := getGraphContext(session, update)
	if element, err := getElementFromDomain(update, &context); err != nil {
		slog.Error("Cannot retrieve element update from domain", slog.String("error", err.Error()))
		return model.ToolSessionElementUpdated{
			Element: nil,
		}
	} else {
		return model.ToolSessionElementUpdated{
			Element: element,
		}
	}
}

func ElementInputToDomain(elementInput model.ElementInput) appdev.AppSessionElementUpdate {
	return appdev.AppSessionElementUpdate{
		ElementID:   elementInput.ElementID,
		StringValue: elementInput.TextValue,
		NumberValue: elementInput.NumberValue,
		HTMLValue:   elementInput.HTMLValue,
		SliderValue: elementInput.SliderValue,
	}
}

func convertID(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}
