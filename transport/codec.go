package transport

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/alanshaw/ucantone/ipld/codec/dagcbor"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/ucan/container"
)

var (
	DefaultHTTPInboundCodec  = &HTTPInboundCodec{}
	DefaultHTTPOutboundCodec = &HTTPOutboundCodec{}
)

type HTTPInboundCodec struct{}

var _ InboundCodec[*http.Request, *http.Response] = (*HTTPInboundCodec)(nil)

func (h *HTTPInboundCodec) Decode(r *http.Request) (ucan.Container, error) {
	if r.Header.Get("Content-Type") != dagcbor.ContentType {
		return nil, fmt.Errorf("invalid content type %q, expected %q", r.Header.Get("Content-Type"), dagcbor.ContentType)
	}
	buf, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("reading request body: %w", err)
	}
	ct, err := container.Decode(buf)
	if err != nil {
		return nil, fmt.Errorf("decoding container: %w", err)
	}
	return ct, nil
}

func (h *HTTPInboundCodec) Encode(c ucan.Container) (*http.Response, error) {
	buf, err := container.Encode(container.Raw, c)
	if err != nil {
		return nil, fmt.Errorf("encoding response container: %w", err)
	}
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(buf)),
		Header:     make(http.Header),
	}
	resp.Header.Set("Content-Type", dagcbor.ContentType)
	return resp, nil
}

type HTTPOutboundCodec struct{}

var _ OutboundCodec[*http.Request, *http.Response] = (*HTTPOutboundCodec)(nil)

func (h *HTTPOutboundCodec) Encode(c ucan.Container) (*http.Request, error) {
	buf, err := container.Encode(container.Raw, c)
	if err != nil {
		return nil, fmt.Errorf("encoding request container: %w", err)
	}
	req := &http.Request{
		Method: http.MethodPost,
		Body:   io.NopCloser(bytes.NewReader(buf)),
		Header: make(http.Header),
	}
	req.Header.Set("Content-Type", dagcbor.ContentType)
	return req, nil
}

func (h *HTTPOutboundCodec) Decode(r *http.Response) (ucan.Container, error) {
	if r.Header.Get("Content-Type") != dagcbor.ContentType {
		return nil, fmt.Errorf("invalid content type %q, expected %q", r.Header.Get("Content-Type"), dagcbor.ContentType)
	}
	buf, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}
	ct, err := container.Decode(buf)
	if err != nil {
		return nil, fmt.Errorf("decoding container: %w", err)
	}
	return ct, nil
}
