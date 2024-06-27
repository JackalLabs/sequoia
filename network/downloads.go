package network

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/JackalLabs/sequoia/file_system"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	"github.com/jackalLabs/canine-chain/v4/x/storage/types"
	"github.com/rs/zerolog/log"
)

func DownloadFile(f *file_system.FileSystem, merkle []byte, owner string, start int64, wallet *wallet.Wallet, fileSize int64, myUrl string, chunkSize int64) error {
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

		size, err := DownloadFileFromURL(f, url, merkle, owner, start, wallet.AccAddress(), chunkSize)
		if err != nil {
			log.Info().Msg(fmt.Sprintf("Couldn't get %x from %s, trying again...", merkle, url))
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

	log.Debug().Msg(fmt.Sprintf("Done downloading %x", merkle))

	return nil
}

func DownloadFileFromURL(f *file_system.FileSystem, url string, merkle []byte, owner string, start int64, address string, chunkSize int64) (int, error) {
	log.Info().Msg(fmt.Sprintf("Downloading %x from %s...", merkle, url))
	cli := http.Client{}
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

	size, _, err := f.WriteFile(reader, merkle, owner, start, address, chunkSize)
	if err != nil {
		return 0, err
	}

	return size, nil
}
