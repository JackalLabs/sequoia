package ipfs

import (
	"fmt"
	"strings"
	"testing"

	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"
)

type test struct {
	domain      string
	shouldError bool
}

func TestIPFSDomains(t *testing.T) {
	r := require.New(t)
	domains := []test{
		{
			domain:      "dns4/ipfs2.squirrellogic.com/tcp/4001",
			shouldError: false,
		},
		{
			domain:      "/dns4/ipfs2.squirrellogic.com/tcp/4001",
			shouldError: false,
		},
	}

	for _, domain := range domains {
		customDomain := domain.domain

		if !strings.Contains(customDomain, "example.com") && len(customDomain) > 2 {
			if !strings.HasPrefix(customDomain, "/") {
				customDomain = fmt.Sprintf("/%s", customDomain)
			}
			_, err := multiaddr.NewMultiaddr(customDomain)
			if domain.shouldError {
				r.Error(err)
			} else {
				r.NoError(err)
			}
		}
	}
}
