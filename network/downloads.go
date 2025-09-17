package network

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/andybalholm/brotli"

	apiTypes "github.com/JackalLabs/sequoia/api/types"

	ipfslite "github.com/hsanjuan/ipfs-lite"

	"github.com/JackalLabs/sequoia/file_system"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	"github.com/jackalLabs/canine-chain/v4/x/storage/types"
	"github.com/rs/zerolog/log"
)

var urlMap map[string]string

func init() {
	log.Info().Msg("Importing url replacement map...")
	data, err := os.ReadFile("urlmap.json")
	if err != nil {
		log.Warn().Err(err).Msg("Could not import URL map.")
		return
	}

	err = json.Unmarshal(data, &urlMap)
	if err != nil {
		log.Warn().Err(err).Msg("Could not parse url map.")
		return
	}
}

// DownloadFile attempts to download a file identified by its Merkle root from a network of providers, excluding the caller's own URL.
// It queries the provider network, tries each available provider until the file is successfully downloaded and matches the expected size, and writes the file to the local file system.
// Returns an error if the file cannot be found or downloaded from any provider.
func DownloadFile(f *file_system.FileSystem, merkle []byte, owner string, start int64, wallet *wallet.Wallet, fileSize int64, myUrl string, chunkSize int64, proofType int64, ipfsParams *ipfslite.AddParams) error {
	queryParams := &types.QueryFindFile{
		Merkle: merkle,
	}

	cl := types.NewQueryClient(wallet.Client.GRPCConn)

	res, err := cl.FindFile(context.Background(), queryParams)
	if err != nil {
		return err
	}

	arr := res.ProviderIps

	if len(arr) == 0 {
		return fmt.Errorf("%x not found on provider network", merkle)
	}

	foundFile := false
	for _, url := range arr {
		if url == myUrl {
			continue
		}

		mappedUrl, found := urlMap[url]
		if found {
			log.Info().Msgf("Swapping internal URL from %s to %s", url, mappedUrl)
			url = mappedUrl
		}

		size, err := DownloadFileFromURL(f, url, merkle, owner, start, chunkSize, proofType, ipfsParams, fileSize)
		if err != nil {
			log.Info().Msg(fmt.Sprintf("Couldn't get %x from %s, trying again... | %s", merkle, url, err.Error()))
			continue
		}
		if fileSize != int64(size) {
			continue
		}

		foundFile = true
		break
	}
	if !foundFile {
		log.Debug().Msg(fmt.Sprintf("Could not find %x on any providers...", merkle))
		return fmt.Errorf("failed to find file on network")
	}

	log.Debug().Msg(fmt.Sprintf("Done downloading %x", merkle))

	return nil
}

// DownloadFileFromURL downloads a file chunk from a provider URL and writes it to the local file system.
//
// The function performs an HTTP GET request to the provider's download endpoint for the specified Merkle root,
// using a dynamically calculated timeout based on the file size. If the download is successful, the data is
// written to the file system using the provided chunking and proof parameters.
//
// Returns the number of bytes written, or an error if the download or write fails. Timeout and HTTP errors are
// reported with detailed messages.
func DownloadFileFromURL(f *file_system.FileSystem, url string, merkle []byte, owner string, start int64, chunkSize int64, proofType int64, ipfsParams *ipfslite.AddParams, fileSize int64) (int, error) {
	log.Info().Msgf("Downloading %x from %s...", merkle, url)

	// Calculate timeout based on file size
	// Base timeout + additional time for large files
	baseTimeout := 30 * time.Second
	bytesPerSecond := int64(1024 * 1024 * 10) // 10MB/s as a conservative estimate

	var timeout time.Duration
	// Add 1 second per MB with an upper limit
	additionalTime := time.Duration(fileSize/bytesPerSecond) * time.Second
	maxTimeout := 30 * time.Minute
	timeout = baseTimeout + additionalTime
	if timeout > maxTimeout {
		timeout = maxTimeout
	}
	log.Debug().Msg(fmt.Sprintf("Using timeout of %v for %d bytes", timeout, fileSize))

	// Create a client with timeout
	transport := &http.Transport{
		ResponseHeaderTimeout: timeout,
		ExpectContinueTimeout: 5 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		IdleConnTimeout:       90 * time.Second,
	}

	cli := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/download/%x", url, merkle), nil)
	if err != nil {
		return 0, err
	}

	req.Header = http.Header{
		"User-Agent":                {"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.62 Safari/537.36"},
		"Upgrade-Insecure-Requests": {"1"},
		"Accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8"},
		"Accept-Encoding":           {"gzip, deflate, br"},
		"Accept-Language":           {"en-US,en;q=0.9"},
		"Connection":                {"keep-alive"},
	}

	// Add context with timeout for more control
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := cli.Do(req)
	if err != nil {
		// Check if the error is a timeout
		if os.IsTimeout(err) || errors.Is(err, context.DeadlineExceeded) {
			return 0, fmt.Errorf("download timed out after %v: %w", timeout, err)
		}
		return 0, err
	}

	if resp.StatusCode != 200 {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return 0, fmt.Errorf("could not read body, %w | code: %d", err, resp.StatusCode)
		}
		var e apiTypes.ErrorResponse
		err = json.Unmarshal(data, &e)
		if err != nil {
			return 0, fmt.Errorf("could not read json body, %w | code: %d", err, resp.StatusCode)
		}

		return 0, fmt.Errorf("could not get file, code: %d | msg: %s", resp.StatusCode, e.Error)
	}
	//nolint:errcheck
	defer resp.Body.Close()

	var bodyReader io.Reader = resp.Body
	contentEncoding := resp.Header.Get("Content-Encoding")
	log.Info().Str("merkle", fmt.Sprintf("%x", merkle)).Msgf("Downloads content encoding: %s", contentEncoding)
	switch contentEncoding {
	case "gzip":
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			return 0, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		//nolint:errcheck
		defer gz.Close()
		bodyReader = gz
	case "deflate":
		deflateReader := flate.NewReader(resp.Body)
		//nolint:errcheck
		defer deflateReader.Close()
		bodyReader = deflateReader
	case "br":
		bodyReader = brotli.NewReader(resp.Body)
	default:
		// No compression or unsupported; use raw body
	}

	buff := bytes.NewBuffer([]byte{})

	// Use TeeReader to monitor for context cancellation while copying
	doneCh := make(chan struct{})
	errCh := make(chan error, 1)

	go func() {
		_, err := io.Copy(buff, bodyReader)
		if err != nil {
			errCh <- err
		}
		close(doneCh)
	}()

	// Wait for either completion or timeout
	select {
	case <-ctx.Done():
		return 0, fmt.Errorf("download timed out after %v", timeout)
	case err := <-errCh:
		return 0, fmt.Errorf("download error: %w", err)
	case <-doneCh:
		// Download completed successfully
	}

	reader := bytes.NewReader(buff.Bytes())

	size, _, err := f.WriteFile(reader, merkle, owner, start, chunkSize, proofType, ipfsParams)
	if err != nil {
		return 0, fmt.Errorf("failed to write file data: %w", err)
	}

	return size, nil
}
