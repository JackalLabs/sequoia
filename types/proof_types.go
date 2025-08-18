package types

import "io"

type Hash interface {
	io.Writer
	Sum(b []byte) []byte
}

const (
	ProofTypeDefault    = 0
	ProofTypeIPFSFolder = 1
	ProofTypeBlake3     = 3
)
