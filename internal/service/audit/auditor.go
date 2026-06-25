package audit

import (
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Observer receives audit events from Event.
type Observer interface {
	UpdateAudit(evAudit *EventAudit) error
}

// Event publishes audit events to registered observers.
type Event struct {
	observers []Observer
	appLogger *zap.Logger
}

// NewEvent creates an audit event publisher.
func NewEvent(appLogger *zap.Logger) *Event {
	return &Event{appLogger: appLogger}
}

// Register subscribes an audit observer to future events.
func (e *Event) Register(o Observer) {
	if o == nil {
		return
	}
	e.observers = append(e.observers, o)
}

func (e *Event) notify(evAudit *EventAudit) {
	for _, observer := range e.observers {
		if err := observer.UpdateAudit(evAudit); err != nil && e.appLogger != nil {
			e.appLogger.Error("Failed to send audit event", zap.Error(err))
		}
	}
}

// Update sends an audit event to all registered observers.
func (e *Event) Update(evAudit *EventAudit) {
	e.notify(evAudit)
}

// EventAudit describes a URL shortener audit event.
type EventAudit struct {
	// Timestamp is a Unix timestamp in seconds.
	Timestamp int64 `json:"ts"`
	// Action names the operation that produced the event.
	Action string `json:"action"`
	// UserID identifies the user associated with the event when available.
	UserID uuid.UUID `json:"user_id"`
	// URL is the original URL involved in the event.
	URL string `json:"url"`
}
