package storage

import (
	"context"
	"io"
)

// Storage defines the interface for file storage operations
type Storage interface {
	// Upload uploads data and returns the public URL
	Upload(ctx context.Context, key string, data []byte, contentType string) (string, error)

	// Download downloads data from storage
	Download(ctx context.Context, key string) ([]byte, error)

	// GetReader returns a reader for streaming
	GetReader(ctx context.Context, key string) (io.ReadCloser, string, error)

	// Delete deletes a file
	Delete(ctx context.Context, key string) error

	// Exists checks if a file exists
	Exists(ctx context.Context, key string) (bool, error)

	// GetPublicURL returns the public URL for a key
	GetPublicURL(key string) string
}
