package audit

import (
	"encoding/json"
	"os"
	"sync"

	"go.uber.org/zap"
)

type storageAuditor struct {
	mu      sync.Mutex
	file    *os.File
	encoder *json.Encoder
}

func NewStorageAuditor(path string) (observer, func() error, error) {
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

func (sa *storageAuditor) updateAudit(evAudit *EventAudit, appLogger *zap.Logger) {
	err := sa.saveAuditFS(evAudit)
	if err != nil {
		appLogger.Error("Failed to save audit event", zap.Error(err))
	}
}

func (as *storageAuditor) saveAuditFS(evAudit *EventAudit) error {
	as.mu.Lock()
	defer as.mu.Unlock()
	err := as.encoder.Encode(evAudit)
	if err != nil {
		return err
	}
	return nil
}

func (as *storageAuditor) Close() error {
	return as.file.Close()
}
