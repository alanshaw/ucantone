package transport

import "io"

type Request interface {
	Body() io.Reader
}

type Response interface {
	Body() io.ReadCloser
}
