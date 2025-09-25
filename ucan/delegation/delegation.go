package delegation

import (
	"github.com/alanshaw/ucantone/ipld"
	"github.com/alanshaw/ucantone/ucan"
)

type Delegation struct{}

func (d *Delegation) Arguments() any {
	panic("unimplemented")
}

func (d *Delegation) Audience() ucan.Principal {
	panic("unimplemented")
}

func (d *Delegation) Command() ucan.Command {
	panic("unimplemented")
}

func (d *Delegation) Expiration() *ucan.UTCUnixTimestamp {
	panic("unimplemented")
}

func (d *Delegation) Issuer() ucan.Principal {
	panic("unimplemented")
}

func (d *Delegation) Metadata() ipld.Map[string, any] {
	panic("unimplemented")
}

func (d *Delegation) Nonce() ucan.Nonce {
	panic("unimplemented")
}

func (d *Delegation) NotBefore() *ucan.UTCUnixTimestamp {
	panic("unimplemented")
}

func (d *Delegation) Policy() []string {
	panic("unimplemented")
}

func (d *Delegation) Proofs() []ucan.Link {
	panic("unimplemented")
}

func (d *Delegation) Signature() ucan.Signature {
	panic("unimplemented")
}

func (d *Delegation) Subject() ucan.Principal {
	panic("unimplemented")
}

var _ ucan.Delegation = (*Delegation)(nil)

func Encode(dlg ucan.Delegation) ([]byte, error) {
	panic("not implemented")
}

func Decode(input []byte) (*Delegation, error) {
	panic("not implemented")
}
