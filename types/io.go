package types

import "io"

type FileReader interface {
	io.Reader
	io.Seeker
	io.Closer
}
