package utils

import (
	"crypto/sha256"
	"fmt"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"io"
)

func MakeFid(reader io.Reader) (string, error) {
	h := sha256.New()
	_, err := io.Copy(h, reader)
	if err != nil {
		return "", err
	}
	hashName := h.Sum(nil)

	return bech32.ConvertAndEncode("jklf", hashName)
}

func MakeCid(signee string, creator string, fid string) (string, error) {
	h := sha256.New()
	_, err := io.WriteString(h, fmt.Sprintf("%s%s%s", signee, creator, fid))
	if err != nil {
		return "", err
	}
	hashName := h.Sum(nil)

	return bech32.ConvertAndEncode("jklc", hashName)
}
