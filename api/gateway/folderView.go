package gateway

import (
	"bytes"
	_ "embed"
	"encoding/hex"
	"fmt"
	"github.com/JackalLabs/sequoia/types"
	"html/template"
)

//go:embed folderView.html
var tmpl string

// TemplateData is a struct to hold both the folder data and the current path
type TemplateData struct {
	Folder      *types.FolderData
	CurrentPath string
}

// formatSize formats file size to a human-readable format
func formatSize(size uint) string {
	if size == 0 {
		return "0 B"
	}

	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}

	div, exp := uint64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// encodeMerkleHex encodes the merkle bytes to a hex string
func encodeMerkleHex(merkle []byte) string {
	return hex.EncodeToString(merkle)
}

// truncateMerkleHex returns a shortened version of the merkle hash for display in hex
func truncateMerkleHex(merkle []byte) string {
	encoded := encodeMerkleHex(merkle)
	if len(encoded) <= 10 {
		return encoded
	}
	return encoded[:6] + "..." + encoded[len(encoded)-4:]
}

// GenerateHTML generates HTML content for the given folder data and returns it as a byte slice
func GenerateHTML(folder *types.FolderData, currentPath string) ([]byte, error) {
	// Create template functions
	funcMap := template.FuncMap{
		"formatSize":     formatSize,
		"encodeMerkle":   encodeMerkleHex,   // Changed to use hex encoding
		"truncateMerkle": truncateMerkleHex, // Changed to use hex encoding
	}

	// Create template data that includes both folder and current path
	data := TemplateData{
		Folder:      folder,
		CurrentPath: currentPath,
	}

	// Parse template with functions
	t, err := template.New("folderTemplate").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return nil, fmt.Errorf("error parsing template: %v", err)
	}

	// Create a buffer to store the rendered HTML
	var buf bytes.Buffer

	// Execute template writing to the buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("error executing template: %v", err)
	}

	return buf.Bytes(), nil
}
