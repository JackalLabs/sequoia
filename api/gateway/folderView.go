package gateway

import (
	"bytes"
	_ "embed"
	"encoding/hex"
	"fmt"
	"html/template"

	"github.com/JackalLabs/sequoia/types"
)

//go:embed folderView.html
var tmpl string

// TemplateData is a struct to hold both the folder data and the current path
type TemplateData struct {
	Folder      *types.FolderData
	CurrentPath string
}

// formatSize returns a human-readable string representation of a file size in bytes, using units such as B, KB, MB, GB, etc.
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

// encodeMerkleHex returns the hexadecimal string representation of the given Merkle hash bytes.
func encodeMerkleHex(merkle []byte) string {
	return hex.EncodeToString(merkle)
}

// truncateMerkleHex returns a shortened hexadecimal string of the Merkle hash, displaying the first 6 and last 4 characters separated by ellipsis if the encoded length exceeds 10 characters.
func truncateMerkleHex(merkle []byte) string {
	encoded := encodeMerkleHex(merkle)
	if len(encoded) <= 10 {
		return encoded
	}
	return encoded[:6] + "..." + encoded[len(encoded)-4:]
}

// GenerateHTML renders HTML content for the specified folder data and current path using an embedded template.
// It returns the generated HTML as a byte slice, or an error if template parsing or execution fails.
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
