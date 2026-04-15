package inmemory

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/Luclpor/url_shortener.git/internal/model"
)

type FileStorage struct {
	file    *os.File
	encoder *json.Encoder
}

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

func (s *FileStorage) Close() error {
	return s.file.Close()
}

func (s *FileStorage) SaveInFile(url model.ShortenURL) (err error) {
	return s.encoder.Encode(url)
}

func (s *FileStorage) SaveInFileBatch(url []model.ShortenURL) (err error) {
	for _, u := range url {
		if err = s.encoder.Encode(u); err != nil {
			return err
		}
	}
	return nil
}

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
