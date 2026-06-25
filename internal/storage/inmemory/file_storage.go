package inmemory

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/Luclpor/url_shortener.git/internal/model"
)

// FileStorage persists URL records as JSON lines in a local file.
type FileStorage struct {
	file    *os.File
	encoder *json.Encoder
}

// NewFileStorage opens or creates a file-backed URL storage.
func NewFileStorage(filePath string) (*FileStorage, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return &FileStorage{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

// Close closes the underlying storage file.
func (s *FileStorage) Close() error {
	return s.file.Close()
}

// SaveInFile appends one URL record to the storage file.
func (s *FileStorage) SaveInFile(url model.ShortenURL) (err error) {
	return s.encoder.Encode(url)
}

// SaveInFileBatch appends several URL records to the storage file.
func (s *FileStorage) SaveInFileBatch(url []model.ShortenURL) (err error) {
	for _, u := range url {
		if err = s.encoder.Encode(u); err != nil {
			return err
		}
	}
	return nil
}

// ScantTo reads stored URL records into the provided slice.
func (s *FileStorage) ScantTo(urls []model.ShortenURL) (err error) {
	scanner := bufio.NewScanner(s.file)
	for scanner.Scan() {
		var u model.ShortenURL
		if err = json.Unmarshal(scanner.Bytes(), &u); err != nil {
			_ = s.file.Close()
			return err
		}
		urls = append(urls, u)
	}
	if err := scanner.Err(); err != nil {
		_ = s.file.Close()
		return err
	}
	return nil
}
