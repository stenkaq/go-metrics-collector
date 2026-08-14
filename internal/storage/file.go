package storage

import (
	"encoding/json"
	"errors"
	models "go-metrics-collector/internal/model"
	"io"
	"os"
	"sync"
)

type FileStorage struct {
	mu   sync.Mutex
	path string
}

func NewFileStorage(path string) *FileStorage {
	return &FileStorage{path: path}
}

func (s *FileStorage) Dump(metrics []models.Metrics) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	producer, err := NewProducer(s.path)
	if err != nil {
		return err
	}
	defer producer.Close()

	return producer.WriteMetrics(metrics)
}

func (s *FileStorage) Load() ([]models.Metrics, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	consumer, err := NewConsumer(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}

		return nil, err
	}
	defer consumer.Close()

	metrics, err := consumer.ReadMetrics()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, nil
		}

		return nil, err
	}

	return metrics, nil
}

type Producer struct {
	file    *os.File
	encoder *json.Encoder
}

type Consumer struct {
	file    *os.File
	decoder *json.Decoder
}

func NewProducer(fileName string) (*Producer, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}

	return &Producer{file: file, encoder: json.NewEncoder(file)}, nil
}

func (p *Producer) WriteMetrics(m []models.Metrics) error {
	return p.encoder.Encode(&m)
}

func (p *Producer) Close() error {
	return p.file.Close()
}

func NewConsumer(fileName string) (*Consumer, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(file)

	return &Consumer{file: file, decoder: decoder}, nil
}

func (c *Consumer) ReadMetrics() ([]models.Metrics, error) {
	metrics := []models.Metrics{}

	if err := c.decoder.Decode(&metrics); err != nil {
		return nil, err
	}

	return metrics, nil
}

func (c *Consumer) Close() error {
	return c.file.Close()
}
