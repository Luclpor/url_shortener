package audit_test

import (
	"testing"

	"github.com/Luclpor/url_shortener.git/internal/service/audit"
	"go.uber.org/zap"
)

type externalObserver struct {
	received *audit.EventAudit
}

func (o *externalObserver) UpdateAudit(evAudit *audit.EventAudit) error {
	o.received = evAudit
	return nil
}

func TestObserverCanBeImplementedOutsideAuditPackage(t *testing.T) {
	var _ audit.Observer = (*externalObserver)(nil)

	observer := &externalObserver{}
	event := &audit.EventAudit{Action: "shorten", URL: "https://example.com"}
	publisher := audit.NewEvent(zap.NewNop())

	publisher.Register(observer)
	publisher.Update(event)

	if observer.received != event {
		t.Fatalf("observer received %v, want %v", observer.received, event)
	}
}
