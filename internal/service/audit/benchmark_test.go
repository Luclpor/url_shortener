package audit

import (
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type benchmarkObserver struct{}

func (benchmarkObserver) UpdateAudit(*EventAudit) error {
	return nil
}

func BenchmarkEventUpdate(b *testing.B) {
	publisher := NewEvent(zap.NewNop())
	publisher.Register(benchmarkObserver{})
	publisher.Register(benchmarkObserver{})
	event := &EventAudit{
		Timestamp: 12345678,
		Action:    "shorten",
		UserID:    uuid.New(),
		URL:       "https://example.com/original",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		publisher.Update(event)
	}
}
