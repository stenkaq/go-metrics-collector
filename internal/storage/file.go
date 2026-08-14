package storage

import (
	"encoding/json"
	models "go-metrics-collector/internal/model"
	"os"
)

type FileStorage struct {
	path string
}

func NewFileStorage(path string) *FileStorage {
	return &FileStorage{path: path}
}

func (s *FileStorage) Dump(metrics []models.Metrics) error {
	producer, err := NewProducer(s.path)
	if err != nil {
		return err
	}
	defer producer.Close()

	return producer.WriteMetrics(metrics)
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
