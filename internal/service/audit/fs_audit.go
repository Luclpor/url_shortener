package audit

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
)

type storageAuditor struct {
	mu      sync.Mutex
	file    *os.File
	encoder *json.Encoder
}

// NewStorageAuditor creates an audit observer that appends JSON events to path.
func NewStorageAuditor(path string) (Observer, func() error, error) {
	if path == "" {
		return nil, nil, nil
	}
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, nil, err
	}
	sa := &storageAuditor{
		file:    file,
		encoder: json.NewEncoder(file),
	}
	return sa, sa.Close, nil
}

func (sa *storageAuditor) UpdateAudit(evAudit *EventAudit) error {
	return sa.saveAuditFS(evAudit)
}

func (sa *storageAuditor) saveAuditFS(evAudit *EventAudit) error {
	sa.mu.Lock()
	defer sa.mu.Unlock()
	err := sa.encoder.Encode(evAudit)
	if err != nil {
		return err
	}
	return nil
}

// Close flushes audit events to disk and closes the file used by the storage auditor.
func (sa *storageAuditor) Close() error {
	sa.mu.Lock()
	defer sa.mu.Unlock()
	return errors.Join(sa.file.Sync(), sa.file.Close())
}
