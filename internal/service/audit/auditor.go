package audit

import (
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type publisher interface {
	register()
	notify(evAudit *EventAudit)
}

type observer interface {
	updateAudit(evAudit *EventAudit, appLogger *zap.Logger)
}

type Event struct {
	observers []observer
	appLogger *zap.Logger
}

func (e *Event) Register(o observer) {
	e.observers = append(e.observers, o)
}

func (e *Event) notify(evAudit *EventAudit) {
	for _, observer := range e.observers {
		observer.updateAudit(evAudit, e.appLogger)
	}
}

func (e *Event) Update(evAudit *EventAudit) {
	e.notify(evAudit)
}

type EventAudit struct {
	Timestamp time.Time `json:"ts"`
	Action    string    `json:"action"`
	UserId    uuid.UUID `json:"user_id"`
	URL       string    `json:"url"`
}
