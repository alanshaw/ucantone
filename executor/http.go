package executor

import (
	"io"
	"net/http"
	"net/url"
)

type HTTPRequest interface {
	Input
	URL() url.URL
	Header() http.Header
}

type HTTPResponse interface {
	Output
	SetStatusCode(int)
	SetHeader(http.Header)
	SetBody(io.Reader)
}
