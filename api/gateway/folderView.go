package gateway

import (
	_ "embed"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"sync"

	"github.com/JackalLabs/sequoia/types"
)

//go:embed folderView.html
var tmplSrc string

var (
	folderViewTmpl *template.Template
	tmplOnce       sync.Once
)

func compileTemplate() {
	funcMap := template.FuncMap{
		"formatSize":     formatSize,
		"encodeMerkle":   encodeMerkleHex,
		"truncateMerkle": truncateMerkleHex,
	}
	folderViewTmpl = template.Must(template.New("folderTemplate").Funcs(funcMap).Parse(tmplSrc))
}

func GenerateHTML(folder *types.FolderData, currentPath string) (io.ReadSeekCloser, error) {
	tmplOnce.Do(compileTemplate)
	var buf types.BytesSeeker
	if err := folderViewTmpl.Execute(&buf, TemplateData{Folder: folder, CurrentPath: currentPath}); err != nil {
		return nil, fmt.Errorf("error executing template: %w", err)
	}

	return &buf, nil
}

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
