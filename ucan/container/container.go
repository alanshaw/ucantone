package container

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/ucan/container/datamodel"
	"github.com/alanshaw/ucantone/ucan/invocation"
	"github.com/ipfs/go-cid"
)

const (
	Raw           = byte(0x40) // raw bytes, no compression
	Base64        = byte(0x42) // base64 std padding, no compression
	Base64url     = byte(0x43) // base64 url (no padding), no compression
	RawGzip       = byte(0x44) // raw bytes, gzip
	Base64Gzip    = byte(0x45) // base64 std padding, gzip
	Base64urlGzip = byte(0x46) // base64 url (no padding), gzip
)

var ErrNotFound = errors.New("not found")

// FormatCodec converts a container codec code into a human readable string.
func FormatCodec(codec byte) string {
	switch codec {
	case Raw:
		return "raw"
	case Base64:
		return "base64"
	case Base64url:
		return "base64url"
	case RawGzip:
		return "raw+gzip"
	case Base64Gzip:
		return "base64+gzip"
	case Base64urlGzip:
		return "base64url+gzip"
	default:
		return "unknown"
	}
}

type Container struct {
	invs  []ucan.Invocation
	rcpts []ucan.Receipt
	dlgs  []ucan.Delegation
}

func (c *Container) Invocations() []ucan.Invocation {
	return c.invs
}

func (c *Container) Delegations() []ucan.Delegation {
	return c.dlgs
}

func (c *Container) Delegation(root cid.Cid) (ucan.Delegation, error) {
	panic("not implemented")
}

func (c *Container) Receipts() []ucan.Receipt {
	return c.rcpts
}

func (c *Container) Receipt(about cid.Cid) (ucan.Receipt, error) {
	for _, inv := range c.invs {
		if inv.Command() == "/ucan/assert" {
			// TODO: inspect inv.Args to see if `about` field matches
		}
	}
	return nil, ErrNotFound
}

type ContainerConfig struct{}

type Option func(c *Container)

func WithInvocations(invocations ...ucan.Invocation) Option {
	return func(c *Container) {
		c.invs = append(c.invs, invocations...)
	}
}

func WithDelegations(delegations ...ucan.Delegation) Option {
	return func(c *Container) {
		c.dlgs = append(c.dlgs, delegations...)
	}
}

func WithReceipts(receipts ...ucan.Receipt) Option {
	return func(c *Container) {
		c.rcpts = append(c.rcpts, receipts...)
	}
}

// TODO: create and store model on container and add accessor (`Model()`)
func New(options ...Option) *Container {
	ct := Container{}
	for _, opt := range options {
		opt(&ct)
	}
	return &ct
}

func Encode(codec byte, container ucan.Container) ([]byte, error) {
	var tokens [][]byte
	for _, inv := range container.Invocations() {
		b, err := invocation.Encode(inv)
		if err != nil {
			return nil, fmt.Errorf("encoding invocation: %w", err)
		}
		tokens = append(tokens, b)
	}
	// for _, dlg := range container.Delegations() {
	// 	b, err := delegation.Encode(dlg)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("encoding delegation: %w", err)
	// 	}
	// 	tokens = append(tokens, b)
	// }
	// for _, rcpt := range container.Receipts() {
	// 	b, err := receipt.Encode(rcpt)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("encoding receipt: %w", err)
	// 	}
	// 	tokens = append(tokens, b)
	// }
	model := datamodel.ContainerModel{Ctn1: tokens}
	var buf bytes.Buffer
	err := model.MarshalCBOR(&buf)
	if err != nil {
		return nil, fmt.Errorf("marshaling container to CBOR: %w", err)
	}

	var out []byte
	if codec == RawGzip || codec == Base64Gzip || codec == Base64urlGzip {
		var gzbuf bytes.Buffer
		gz := gzip.NewWriter(&gzbuf)
		_, err := io.Copy(gz, &buf)
		if err != nil {
			gz.Close()
			return nil, fmt.Errorf("compressing container data: %w", err)
		}
		if err := gz.Close(); err != nil {
			return nil, fmt.Errorf("closing gzip writer: %w", err)
		}
		out = gzbuf.Bytes()
	} else {
		out = buf.Bytes()
	}

	switch codec {
	case Raw, RawGzip:
		// nothing to do
		break
	case Base64, Base64Gzip:
		out = []byte(base64.StdEncoding.EncodeToString(out))
	case Base64url, Base64urlGzip:
		out = []byte(base64.RawURLEncoding.EncodeToString(out))
	default:
		return nil, fmt.Errorf("unknown codec: 0x%02x", codec)
	}

	return append([]byte{codec}, out...), nil
}

func Decode(input []byte) (*Container, error) {
	if len(input) == 0 {
		return nil, errors.New("empty container bytes")
	}
	codec := input[0]
	var compressed []byte
	switch codec {
	case Raw, RawGzip:
		compressed = input[1:]
	case Base64, Base64Gzip:
		r, err := base64.StdEncoding.DecodeString(string(input[1:]))
		if err != nil {
			return nil, fmt.Errorf("decoding base64: %w", err)
		}
		compressed = r
	case Base64url, Base64urlGzip:
		r, err := base64.RawURLEncoding.DecodeString(string(input[1:]))
		if err != nil {
			return nil, fmt.Errorf("decoding base64: %w", err)
		}
		compressed = r
	default:
		return nil, fmt.Errorf("unknown codec: 0x%02x", codec)
	}

	var raw []byte
	if codec == RawGzip || codec == Base64Gzip || codec == Base64urlGzip {
		gz, err := gzip.NewReader(bytes.NewReader(compressed))
		if err != nil {
			return nil, fmt.Errorf("creating gzip reader: %w", err)
		}
		defer gz.Close()
		if raw, err = io.ReadAll(gz); err != nil {
			return nil, fmt.Errorf("reading gzipped data: %w", err)
		}
	} else {
		raw = compressed // not compressed
	}

	model := &datamodel.ContainerModel{}
	if err := model.UnmarshalCBOR(bytes.NewReader(raw)); err != nil {
		return nil, fmt.Errorf("unmarshalling container from CBOR: %w", err)
	}

	var invs []ucan.Invocation
	var rcpts []ucan.Receipt
	var dlgs []ucan.Delegation
	for _, b := range model.Ctn1 {
		if inv, err := invocation.Decode(b); err == nil {
			invs = append(invs, inv)
			continue
		}
		// TODO: reinstate when implemented
		// if rcpt, err := receipt.Decode(b); err != nil {
		// 	rcpts = append(rcpts, rcpt)
		// 	continue
		// }
		// if dlg, err := delegation.Decode(b); err != nil {
		// 	dlgs = append(dlgs, dlg)
		// 	continue
		// }
	}

	return New(WithInvocations(invs...), WithDelegations(dlgs...), WithReceipts(rcpts...)), nil
}
