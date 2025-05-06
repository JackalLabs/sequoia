package gateway_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JackalLabs/sequoia/api/gateway"
	"github.com/JackalLabs/sequoia/types"
	"github.com/stretchr/testify/require"
)

func TestFolderView(t *testing.T) {
	req := require.New(t)

	// Setup test data
	merkle, err := hex.DecodeString("5eefd1c2857aad83e27e6e6fdef552dddfab38d78b1566673cf69298388a5466822814dc2a401bd80bb9ec875e104e3cdd774290272c274a8470f129f4bcd89b")
	req.NoError(err)

	folderName := "folder"
	exampleFolder := &types.FolderData{
		Name:    folderName,
		Merkle:  merkle,
		Version: 1,
		Children: []types.FileNode{
			{Name: "document.txt", Merkle: []byte("documentMerkle123"), Size: 2048},
			{Name: "image.png", Merkle: []byte("imageMerkle456"), Size: 1024 * 1024 * 3},
			{Name: "data.json", Merkle: []byte("jsonMerkle789"), Size: 4096},
		},
	}

	// Generate HTML bytes
	htmlBytes, err := gateway.GenerateHTML(exampleFolder, "/5eefd1c2857aad83e27e6e6fdef552dddfab38d78b1566673cf69298388a5466822814dc2a401bd80bb9ec875e104e3cdd774290272c274a8470f129f4bcd89b")
	req.NoError(err)

	// Method 1: Use a test server (preferred for unit tests)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rs := bytes.NewReader(htmlBytes)
		http.ServeContent(w, r, "index.html", time.Time{}, rs)
	}))
	defer ts.Close() // This ensures the test server gets closed when the test is done

	// Make a request to the test server
	resp, err := http.Get(ts.URL)
	req.NoError(err)
	//nolint:all
	defer resp.Body.Close()

	// Verify response
	req.Equal(http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	req.NoError(err)
	req.NotEmpty(body)

	t.Log(string(body))

	// Method 2: If you really need a real server for manual testing (not ideal for automated tests)
	if true { // Change to true when you want to manually test
		fmt.Println("Starting server on port 4045, press Ctrl+C to stop...")

		s := http.NewServeMux()
		s.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			rs := bytes.NewReader(htmlBytes)
			http.ServeContent(w, r, "index.html", time.Time{}, rs)
		})

		// Use a goroutine so the test doesn't get stuck
		go func() {
			err := http.ListenAndServe(":4045", s)
			if err != nil {
				fmt.Printf("Server error: %v\n", err)
			}
		}()

		// Keep the server running for a short time for manual testing
		// This is not ideal for automated tests
		time.Sleep(30 * time.Second)
	}
}
