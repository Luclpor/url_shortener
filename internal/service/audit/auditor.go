package audit

import (
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type observer interface {
	updateAudit(evAudit *EventAudit) error
}

type Event struct {
	observers []observer
	appLogger *zap.Logger
}

func NewEvent(appLogger *zap.Logger) *Event {
	return &Event{appLogger: appLogger}
}

func (e *Event) Register(o observer) {
	if o == nil {
		return
	}
	e.observers = append(e.observers, o)
}

func (e *Event) notify(evAudit *EventAudit) {
	for _, observer := range e.observers {
		if err := observer.updateAudit(evAudit); err != nil && e.appLogger != nil {
			e.appLogger.Error("Failed to send audit event", zap.Error(err))
		}
	}
}

func (e *Event) Update(evAudit *EventAudit) {
	e.notify(evAudit)
}

type EventAudit struct {
	Timestamp int64     `json:"ts"`
	Action    string    `json:"action"`
	UserId    uuid.UUID `json:"user_id"`
	URL       string    `json:"url"`
}
