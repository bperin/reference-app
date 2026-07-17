package storage

import (
	"testing"
)

func TestGetObjectKey(t *testing.T) {
	adapter := &GCSAdapter{}
	userID := "user-uuid"
	contentID := "content-uuid"
	originalName := "file.txt"

	expected := "user-uuid/content-uuid/file.txt"
	got := adapter.GetObjectKey(userID, contentID, originalName)

	if got != expected {
		t.Errorf("GetObjectKey() = %q, want %q", got, expected)
	}
}
