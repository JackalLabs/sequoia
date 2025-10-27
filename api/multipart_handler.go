package api

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	sequoiaTypes "github.com/JackalLabs/sequoia/types"
	"github.com/rs/zerolog/log"
)

const (
	MaxMemoryFileSize = 10 << 20 // 10 MB - files larger than this go to disk
	MaxFileSize       = 32 << 30 // 32 GiB - maximum allowed file size
)

// tempFileReader wraps a temporary file to implement the FileReader interface
type tempFileReader struct {
	file *os.File
	size int64
}

func (t *tempFileReader) Read(p []byte) (n int, err error) {
	return t.file.Read(p)
}

func (t *tempFileReader) Seek(offset int64, whence int) (int64, error) {
	return t.file.Seek(offset, whence)
}

func (t *tempFileReader) Close() error {
	// Close and remove the temporary file
	// nolint:errcheck
	t.file.Close()
	return os.Remove(t.file.Name())
}

// readFormField reads a form field value from a multipart part
func readFormField(part *multipart.Part) (string, error) {
	data, err := io.ReadAll(part)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// processSmallFile handles small files by keeping them in memory
func processSmallFile(part *multipart.Part, peekBuffer []byte, n int) (sequoiaTypes.FileReader, *multipart.FileHeader, error) {
	log.Debug().Int64("size", int64(n)).Msg("Using in-memory processing for small file")

	// Check if file exceeds maximum allowed size
	if int64(n) > MaxFileSize {
		return nil, nil, fmt.Errorf("file size %d exceeds maximum allowed size %d", int64(n), MaxFileSize)
	}

	fileData := make([]byte, n)
	copy(fileData, peekBuffer[:n])

	// Create a FileReader from the buffered data
	file := sequoiaTypes.NewBytesSeeker(fileData)
	fh := &multipart.FileHeader{
		Filename: part.FileName(),
		Header:   part.Header,
		Size:     int64(len(fileData)),
	}

	return file, fh, nil
}

// processLargeFile handles large files by streaming them to a temporary file
func processLargeFile(part *multipart.Part, peekBuffer []byte, n int) (sequoiaTypes.FileReader, *multipart.FileHeader, error) {
	log.Debug().Int64("size", int64(n)).Msg("Using disk streaming for large file")

	// Check if file exceeds maximum allowed size
	if int64(n) > MaxFileSize {
		return nil, nil, fmt.Errorf("file size %d exceeds maximum allowed size %d", int64(n), MaxFileSize)
	}

	tempFile, err := os.CreateTemp("", "sequoia_upload_*")
	if err != nil {
		return nil, nil, fmt.Errorf("error creating temporary file: %w", err)
	}

	// Write the peeked data to the temp file
	_, err = tempFile.Write(peekBuffer[:n])
	if err != nil {
		// nolint:errcheck
		tempFile.Close()
		// nolint:errcheck
		os.Remove(tempFile.Name())
		return nil, nil, fmt.Errorf("error writing to temporary file: %w", err)
	}

	// Stream any remaining multipart data directly to the temporary file with size limit
	// Use io.LimitReader to enforce MaxFileSize
	limitedReader := io.LimitReader(part, MaxFileSize-int64(n))
	size, err := io.Copy(tempFile, limitedReader)
	if err != nil {
		// nolint:errcheck
		tempFile.Close()
		// nolint:errcheck
		os.Remove(tempFile.Name())
		return nil, nil, fmt.Errorf("error streaming file data: %w", err)
	}

	// Check if we hit the size limit (meaning file is too large)
	if size == MaxFileSize-int64(n) {
		// Try to read one more byte to see if there's more data
		buf := make([]byte, 1)
		if _, err := part.Read(buf); err == nil {
			// nolint:errcheck
			tempFile.Close()
			// nolint:errcheck
			os.Remove(tempFile.Name())
			return nil, nil, fmt.Errorf("file size exceeds maximum allowed size %d", MaxFileSize)
		}
	}

	totalSize := int64(n) + size

	// Seek back to the beginning of the temporary file
	_, err = tempFile.Seek(0, io.SeekStart)
	if err != nil {
		// nolint:errcheck
		tempFile.Close()
		// nolint:errcheck
		os.Remove(tempFile.Name())
		return nil, nil, fmt.Errorf("error seeking temporary file: %w", err)
	}

	// Create a FileReader from the temporary file
	file := &tempFileReader{
		file: tempFile,
		size: totalSize,
	}
	fh := &multipart.FileHeader{
		Filename: part.FileName(),
		Header:   part.Header,
		Size:     totalSize,
	}

	return file, fh, nil
}

// processFilePart handles the file part of the multipart form using hybrid memory/disk approach
func processFilePart(part *multipart.Part) (sequoiaTypes.FileReader, *multipart.FileHeader, error) {
	// Read a small chunk to determine if we should use memory or disk
	peekBuffer := make([]byte, MaxMemoryFileSize+1) // Read one extra byte to detect if file is larger
	n, err := io.ReadFull(part, peekBuffer)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, nil, fmt.Errorf("error reading file data: %w", err)
	}

	// Check if file fits in memory (n <= MaxMemoryFileSize)
	if n <= MaxMemoryFileSize {
		return processSmallFile(part, peekBuffer, n)
	} else {
		return processLargeFile(part, peekBuffer, n)
	}
}

// parseMultipartFormStreaming parses multipart form data using a streaming approach
// to reduce memory usage. It extracts form fields and streams the file to a temporary file.
func parseMultipartFormStreaming(req *http.Request) (sender, merkleString, startBlockString, proofTypeString string, file sequoiaTypes.FileReader, fh *multipart.FileHeader, err error) {
	// Parse the multipart form boundary
	reader, err := req.MultipartReader()
	if err != nil {
		return "", "", "", "", nil, nil, fmt.Errorf("cannot create multipart reader: %w", err)
	}

	// Read form parts
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", "", "", "", nil, nil, fmt.Errorf("error reading multipart part: %w", err)
		}

		formName := part.FormName()
		if formName == "" {
			continue
		}

		switch formName {
		case "sender":
			sender, err = readFormField(part)
			if err != nil {
				return "", "", "", "", nil, nil, fmt.Errorf("error reading sender field: %w", err)
			}
		case "merkle":
			merkleString, err = readFormField(part)
			if err != nil {
				return "", "", "", "", nil, nil, fmt.Errorf("error reading merkle field: %w", err)
			}
		case "start":
			startBlockString, err = readFormField(part)
			if err != nil {
				return "", "", "", "", nil, nil, fmt.Errorf("error reading start field: %w", err)
			}
		case "type":
			proofTypeString, err = readFormField(part)
			if err != nil {
				return "", "", "", "", nil, nil, fmt.Errorf("error reading type field: %w", err)
			}
		case "file":
			file, fh, err = processFilePart(part)
			if err != nil {
				return "", "", "", "", nil, nil, err
			}
		}
		// nolint:errcheck
		part.Close()
	}

	if file == nil {
		return "", "", "", "", nil, nil, fmt.Errorf("no file found in multipart form")
	}

	return sender, merkleString, startBlockString, proofTypeString, file, fh, nil
}
