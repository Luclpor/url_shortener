package audit

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type testObserver struct {
	events []*EventAudit
	err    error
}

func (o *testObserver) UpdateAudit(evAudit *EventAudit) error {
	o.events = append(o.events, evAudit)
	return o.err
}

func TestEventNotifiesRegisteredObservers(t *testing.T) {
	first := &testObserver{}
	second := &testObserver{err: errors.New("audit sink failed")}
	third := &testObserver{}
	publisher := NewEvent(zap.NewNop())
	event := &EventAudit{
		Timestamp: 12345678,
		Action:    "shorten",
		UserID:    uuid.New(),
		URL:       "https://example.com/original",
	}

	publisher.Register(first)
	publisher.Register(nil)
	publisher.Register(second)
	publisher.Register(third)
	publisher.Update(event)

	require.Len(t, first.events, 1)
	require.Len(t, second.events, 1)
	require.Len(t, third.events, 1)
	assert.Same(t, event, first.events[0])
	assert.Same(t, event, second.events[0])
	assert.Same(t, event, third.events[0])
}

func TestStorageAuditorWritesJSONLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.log")
	event := &EventAudit{
		Timestamp: 12345678,
		Action:    "follow",
		UserID:    uuid.New(),
		URL:       "https://example.com/original",
	}

	observer, closeFn, err := NewStorageAuditor(path)
	require.NoError(t, err)
	require.NotNil(t, observer)
	require.NotNil(t, closeFn)

	require.NoError(t, observer.UpdateAudit(event))
	require.NoError(t, closeFn())

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	require.Len(t, lines, 1)

	var got EventAudit
	require.NoError(t, json.Unmarshal([]byte(lines[0]), &got))
	assert.Equal(t, *event, got)
}

func TestStorageAuditorDisabledWhenPathIsEmpty(t *testing.T) {
	observer, closeFn, err := NewStorageAuditor("")

	require.NoError(t, err)
	assert.Nil(t, observer)
	assert.Nil(t, closeFn)
}

func TestRetryableHTTPClientPostsAuditEvent(t *testing.T) {
	event := &EventAudit{
		Timestamp: 12345678,
		Action:    "shorten",
		UserID:    uuid.New(),
		URL:       "https://example.com/original",
	}
	received := make(chan EventAudit, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var got EventAudit
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		received <- got
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	observer, err := NewRetryableHTTPClient(server.URL)
	require.NoError(t, err)
	require.NotNil(t, observer)

	require.NoError(t, observer.UpdateAudit(event))
	assert.Equal(t, *event, <-received)
}

func TestRetryableHTTPClientRejectsInvalidURL(t *testing.T) {
	observer, err := NewRetryableHTTPClient("/audit")

	require.Error(t, err)
	assert.Nil(t, observer)
}

func TestRetryableHTTPClientReturnsErrorOnReceiverFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	observer, err := NewRetryableHTTPClient(server.URL)
	require.NoError(t, err)

	err = observer.UpdateAudit(&EventAudit{Action: "shorten", URL: "https://example.com"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "audit receiver returned status")
}
