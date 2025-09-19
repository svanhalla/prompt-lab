package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageStore(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "greetd-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store := NewMessageStore(tmpDir)

	// Test initial load (should create default)
	err = store.Load()
	require.NoError(t, err)

	// Test default message
	message := store.GetMessage()
	assert.Equal(t, "Hello, World!", message)

	// Test setting message
	newMessage := "Hello, Universe!"
	err = store.SetMessage(newMessage)
	require.NoError(t, err)

	// Test getting updated message
	message = store.GetMessage()
	assert.Equal(t, newMessage, message)

	// Test persistence by creating new store
	store2 := NewMessageStore(tmpDir)
	err = store2.Load()
	require.NoError(t, err)

	message = store2.GetMessage()
	assert.Equal(t, newMessage, message)
}

func TestMessageStoreFileExists(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "greetd-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create message file with custom content
	messageFile := filepath.Join(tmpDir, "message.json")
	content := `{"message": "Existing message"}`
	err = os.WriteFile(messageFile, []byte(content), 0644)
	require.NoError(t, err)

	store := NewMessageStore(tmpDir)
	err = store.Load()
	require.NoError(t, err)

	message := store.GetMessage()
	assert.Equal(t, "Existing message", message)
}

func TestMessageStoreConcurrency(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "greetd-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store := NewMessageStore(tmpDir)
	err = store.Load()
	require.NoError(t, err)

	// Test concurrent access
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 100; i++ {
			store.SetMessage("Message from goroutine 1")
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			store.GetMessage()
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// Should not panic or race
}
