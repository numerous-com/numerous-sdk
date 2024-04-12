package appdev

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
)

const appSessionEventChanBufferSize = 100

type AppSessionEvent struct {
	AppSessionID           string
	SourceClientID         string
	UpdatedElement         *AppSessionElement
	TriggeredActionElement *AppSessionElement
	AddedElement           *AppSessionElement
	RemovedElement         *AppSessionElement
}

func (a *AppSessionEvent) Type() string {
	if a.UpdatedElement != nil {
		return "UpdatedElement"
	}

	if a.TriggeredActionElement != nil {
		return "TriggeredActionElement"
	}

	if a.AddedElement != nil {
		return "AddedElement"
	}

	if a.RemovedElement != nil {
		return "RemovedElement"
	}

	return "Unknown"
}

type AppSessionSubscription struct {
	ID       uint
	Channel  chan AppSessionEvent
	ClientID string
}

type AppSessionService struct {
	subscriptions      map[uint]AppSessionSubscription
	nextSubscriptionID uint
	appSessions        AppSessionRepository
}

type AppSessionElementUpdate struct {
	ElementID   string
	StringValue *string
	NumberValue *float64
	HTMLValue   *string
	SliderValue *float64
}

type AppSessionElementResult struct {
	Session *AppSession
	Element *AppSessionElement
}

func NewAppSessionService(appSessions AppSessionRepository) AppSessionService {
	return AppSessionService{
		appSessions:   appSessions,
		subscriptions: make(map[uint]AppSessionSubscription),
	}
}

// Update the identified app session with the given update and send updates to
// to all subscribers whose client IDs are different from the specified client ID.
func (s *AppSessionService) UpdateElement(appSessionID uint, clientID string, update AppSessionElementUpdate) (*AppSessionElementResult, error) {
	session, err := s.appSessions.Read(appSessionID)
	if err != nil {
		return nil, err
	}

	convertedAppSessionID := strconv.FormatUint(uint64(appSessionID), 10)
	if e, err := session.GetElementByID(update.ElementID); err != nil {
		slog.Debug("update element not found", slog.Any("id", update.ElementID))
		return nil, err
	} else if err := s.updateElement(convertedAppSessionID, clientID, e, update); err != nil {
		slog.Debug("update updating error", slog.Any("id", e.ID), slog.String("name", e.Name), slog.String("error", err.Error()))
		return nil, err
	} else {
		slog.Debug("update element complete", slog.Any("id", e.ID), slog.String("name", e.Name))
		return &AppSessionElementResult{
			Session: session,
			Element: e,
		}, nil
	}
}

func (s *AppSessionService) updateElement(appSessionID string, clientID string, elem *AppSessionElement, update AppSessionElementUpdate) error {
	switch elem.Type {
	case "string":
		if update.StringValue == nil {
			return fmt.Errorf("string element %d update missing value", elem.ID)
		}
		elem.StringValue.String = *update.StringValue
		elem.StringValue.Valid = true
	case "number":
		if update.NumberValue == nil {
			return fmt.Errorf("number element %d update missing value", elem.ID)
		}
		elem.NumberValue.Float64 = *update.NumberValue
		elem.NumberValue.Valid = true
	case "html":
		if update.HTMLValue == nil {
			return fmt.Errorf("html element %d update missing value", elem.ID)
		}
		elem.HTMLValue.Valid = true
		elem.HTMLValue.String = *update.HTMLValue
	case "slider":
		if update.SliderValue == nil {
			return fmt.Errorf("slider element %d update missing value", elem.ID)
		}
		elem.SliderValue.Float64 = *update.SliderValue
		elem.SliderValue.Valid = true
	default:
		return fmt.Errorf("updating %s element %d is unsupported", elem.Type, elem.ID)
	}

	if err := s.appSessions.UpdateElement(*elem); err != nil {
		return err
	}

	event := AppSessionEvent{
		AppSessionID:   appSessionID,
		SourceClientID: clientID,
		UpdatedElement: elem,
	}
	go s.sendEvent(clientID, event) // TODO: this is only needed for tests, fix

	return nil
}

// Triggers the specified action in the specified app session, sending a trigger
// event to all subscribers whose client IDs are different from the specified
// client ID.
func (s *AppSessionService) TriggerAction(appSessionID uint, clientID string, actionElementID string) (*AppSessionElementResult, error) {
	session, err := s.appSessions.Read(appSessionID)
	if err != nil {
		return nil, err
	}

	convertedAppSessionID := strconv.FormatUint(uint64(appSessionID), 10)
	if actionElement, err := session.GetElementByID(actionElementID); err != nil {
		return nil, err
	} else if actionElement.Type != "action" {
		return nil, fmt.Errorf("cannot trigger action for element \"%s\" of type \"%s\"", actionElementID, actionElement.Type)
	} else {
		slog.Debug("triggering action", slog.Any("triggered action element", actionElementID))

		event := AppSessionEvent{
			AppSessionID:           convertedAppSessionID,
			SourceClientID:         clientID,
			TriggeredActionElement: actionElement,
		}
		s.sendEvent(clientID, event)

		return &AppSessionElementResult{
			Session: session,
			Element: actionElement,
		}, nil
	}
}

func (s *AppSessionService) sendEvent(sourceClientID string, event AppSessionEvent) {
	sent := false

	for _, s := range s.subscriptions {
		if s.ClientID != sourceClientID {
			slog.Info("sending event", slog.String("type", event.Type()), slog.String("sourceClientID", sourceClientID), slog.Any("subscriberClientID", s.ClientID))
			s.Channel <- event
			sent = true
		}
	}

	if !sent {
		slog.Info("no subscribers found for event", slog.String("type", event.Type()), slog.String("sourceClientID", sourceClientID))
	}
}

// Subscribe for updates to an app session, for events not sent by the specified
// client ID.
func (s *AppSessionService) Subscribe(ctx context.Context, appSessionID string, clientID string) (chan AppSessionEvent, error) {
	subscription := make(chan AppSessionEvent, appSessionEventChanBufferSize)
	subID := s.nextSubscriptionID
	s.nextSubscriptionID++
	s.addSubscription(clientID, subID, subscription)

	go func() {
		<-ctx.Done()
		s.removeSubscription(clientID, subID)
	}()

	return subscription, nil
}

func (s *AppSessionService) addSubscription(clientID string, subscriptionID uint, subscription chan AppSessionEvent) {
	slog.Info("adding subscription", slog.String("clientID", clientID), slog.Uint64("subscriptionID", uint64(subscriptionID)))
	s.subscriptions[subscriptionID] = AppSessionSubscription{
		ID:       subscriptionID,
		Channel:  subscription,
		ClientID: clientID,
	}
}

func (s *AppSessionService) removeSubscription(clientID string, subscriptionID uint) {
	slog.Info("removing subscription", slog.String("clientID", clientID), slog.Uint64("subscriptionID", uint64(subscriptionID)))
	delete(s.subscriptions, subscriptionID)
}

// Add an element to an existing app session, sending an element added event to
// all subscribers whose client IDs are different from the specified client ID.
func (s *AppSessionService) AddElement(clientID string, element AppSessionElement) (*AppSession, error) {
	slog.Debug("client adding element", slog.String("clientID", clientID), slog.Any("element", element))
	if element, err := s.appSessions.AddElement(element); err != nil {
		slog.Info("error adding element", slog.Any("element", element), slog.Any("error", err))
		return nil, err
	} else if session, err := s.appSessions.Read(element.AppSessionID); err != nil {
		slog.Info("error adding element", slog.Any("element", element), slog.Any("error", err))
		return nil, err
	} else {
		event := AppSessionEvent{
			SourceClientID: clientID,
			AppSessionID:   strconv.FormatUint(uint64(element.AppSessionID), 10),
			AddedElement:   element,
		}
		s.sendEvent(clientID, event) // TODO: this is only needed for tests, fix

		for _, addedChild := range element.GetAllChildren() {
			event := AppSessionEvent{
				SourceClientID: clientID,
				AppSessionID:   strconv.FormatUint(uint64(element.AppSessionID), 10),
				AddedElement:   &addedChild,
			}
			s.sendEvent(clientID, event)
		}

		return session, nil
	}
}

// Remove an element from  an existing app session, sending an element removed
// event to all subscribers whose client IDs are different from the specified
// client ID.
func (s *AppSessionService) RemoveElement(clientID string, element AppSessionElement) (*AppSession, error) {
	slog.Debug("client removing element", slog.String("clientID", clientID), slog.Any("element", element))
	if err := s.appSessions.RemoveElement(element); err != nil {
		slog.Info("error removing element", slog.Any("element", element), slog.Any("error", err))
		return nil, err
	} else if session, err := s.appSessions.Read(element.AppSessionID); err != nil {
		slog.Info("error removing element", slog.Any("element", element), slog.Any("error", err))
		return nil, err
	} else {
		event := AppSessionEvent{
			SourceClientID: clientID,
			AppSessionID:   strconv.FormatUint(uint64(element.AppSessionID), 10),
			RemovedElement: &element,
		}
		go s.sendEvent(clientID, event) // TODO: consider why this is necessary

		return session, nil
	}
}

// Update an element label in an existing app session, sending an element updated
// event to all subscribers whose client IDs are different from the specified
// client ID.
func (s *AppSessionService) UpdateElementLabel(clientID string, element AppSessionElement) (*AppSession, error) {
	slog.Debug("client updating element label", slog.String("clientID", clientID), slog.Any("element", element))
	if err := s.appSessions.UpdateElement(element); err != nil {
		slog.Info("error updating element label", slog.Any("element", element), slog.Any("error", err))
		return nil, err
	} else if session, err := s.appSessions.Read(element.AppSessionID); err != nil {
		slog.Info("error updating element label", slog.Any("element", element), slog.Any("error", err))
		return nil, err
	} else {
		event := AppSessionEvent{
			SourceClientID: clientID,
			AppSessionID:   strconv.FormatUint(uint64(element.AppSessionID), 10),
			UpdatedElement: &element,
		}
		s.sendEvent(clientID, event) // TODO: consider if this should be a Goroutine

		return session, nil
	}
}
