package network

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/JackalLabs/sequoia/file_system"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	"github.com/dgraph-io/badger/v4"
	"github.com/jackalLabs/canine-chain/v3/x/storage/types"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

func DownloadFile(db *badger.DB, cid string, fid string, wallet *wallet.Wallet, signee string, fileSize int64, myUrl string) error {
	queryParams := &types.QueryFindFileRequest{
		Fid: fid,
	}

	cl := types.NewQueryClient(wallet.Client.GRPCConn)

	res, err := cl.FindFile(context.Background(), queryParams)
	if err != nil {
		return err
	}

	var arr []string // Create an array of IPs from the request.
	err = json.Unmarshal([]byte(res.ProviderIps), &arr)
	if err != nil {
		return err
	}

	if len(arr) == 0 {
		return fmt.Errorf("%s not found on provider network", fid)
	}

	foundFile := false
	for _, url := range arr {
		if url == myUrl {
			continue
		}

		size, err := DownloadFileFromURL(db, url, cid, fid, signee, wallet.AccAddress())
		if err != nil {
			log.Info().Msg(fmt.Sprintf("Couldn't get %s from %s, trying again...", fid, url))
			continue
		}
		if fileSize != int64(size) {
			continue
		}

		foundFile = true
		break
	}
	if !foundFile {
		return fmt.Errorf("failed to find file on network")
	}

	log.Debug().Msg(fmt.Sprintf("Done downloading %s", fid))

	return nil
}

func DownloadFileFromURL(db *badger.DB, url string, cid string, fid string, signee string, address string) (int, error) {
	log.Info().Msg(fmt.Sprintf("Downloading %s from %s...", fid, url))
	cli := http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/download/%s", url, fid), nil)
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

	resp, err := cli.Do(req)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("failed to find file on network")
	}
	defer resp.Body.Close()

	buff := bytes.NewBuffer([]byte{})
	_, err = io.Copy(buff, resp.Body)
	if err != nil {
		return 0, err
	}

	reader := bytes.NewReader(buff.Bytes())

	_, _, _, size, err := file_system.WriteFile(db, reader, signee, address, cid)
	if err != nil {
		return 0, err
	}

	return size, nil
}
