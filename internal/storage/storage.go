package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type MessageStore struct {
	mu       sync.RWMutex
	filePath string
	data     MessageData
}

type MessageData struct {
	Message string `json:"message"`
}

func NewMessageStore(dataPath string) *MessageStore {
	return &MessageStore{
		filePath: filepath.Join(dataPath, "message.json"),
		data:     MessageData{Message: "Hello, World!"},
	}
}

func (s *MessageStore) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		// Create with default message
		return s.saveUnsafe()
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return fmt.Errorf("failed to read message file: %w", err)
	}

	if err := json.Unmarshal(data, &s.data); err != nil {
		return fmt.Errorf("failed to unmarshal message data: %w", err)
	}

	return nil
}

func (s *MessageStore) GetMessage() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.Message
}

func (s *MessageStore) SetMessage(message string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.Message = message
	return s.saveUnsafe()
}

func (s *MessageStore) saveUnsafe() error {
	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal message data: %w", err)
	}

	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write message file: %w", err)
	}

	return nil
}
