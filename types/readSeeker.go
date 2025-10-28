package types

import (
	"fmt"
	"io"
	"os"
	"runtime"
)

// ReadCloserToReadSeekCloser streams rc to a temp file and returns a seekable reader.
// IMPORTANT: The caller must call Close() to delete the temp file.
func ReadCloserToReadSeekCloser(rc io.ReadCloser) (io.ReadSeekCloser, error) {
	tmpFile, err := os.CreateTemp("", "temp-data-*.tmp")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	_, err = io.Copy(tmpFile, rc)
	if err != nil {
		// nolint:errcheck
		tmpFile.Close()
		// nolint:errcheck
		os.Remove(tmpFile.Name()) // Clean up on error
		return nil, fmt.Errorf("failed to copy data to temp file: %w", err)
	}
	// nolint:errcheck
	rc.Close() // Close the original ReadCloser

	_, err = tmpFile.Seek(0, io.SeekStart) // Rewind to the beginning
	if err != nil {
		// nolint:errcheck
		tmpFile.Close()
		// nolint:errcheck
		os.Remove(tmpFile.Name()) // Clean up on error
		return nil, fmt.Errorf("failed to seek temp file to start: %w", err)
	}

	// Wrap the file with a custom closer that deletes the file
	w := &tempFileReadSeekCloser{File: tmpFile}
	// Best-effort safety net if callers forget Close (not guaranteed timing).
	runtime.SetFinalizer(w, func(tf *tempFileReadSeekCloser) { _ = os.Remove(tf.Name()) })
	return w, nil
}

// tempFileReadSeekCloser wraps os.File and deletes it on Close
type tempFileReadSeekCloser struct {
	*os.File
}

func (tfrsc *tempFileReadSeekCloser) Close() error {
	// nolint:errcheck
	defer os.Remove(tfrsc.Name()) // Ensure deletion
	return tfrsc.File.Close()
}
