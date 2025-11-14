package examples

import (
	"fmt"
	"testing"

	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/principal/ed25519"
	"github.com/alanshaw/ucantone/ucan/command"
	"github.com/alanshaw/ucantone/ucan/container"
	"github.com/alanshaw/ucantone/ucan/delegation"
	"github.com/alanshaw/ucantone/ucan/delegation/policy"
	"github.com/alanshaw/ucantone/ucan/invocation"
)

func TestContainer(t *testing.T) {
	// mailer is an email service that can send emails
	mailer, err := ed25519.Generate()
	if err != nil {
		panic(err)
	}

	alice, err := ed25519.Generate()
	if err != nil {
		panic(err)
	}

	pol, err := policy.Build(policy.All(".to", policy.Like(".", "*.example.com")))
	if err != nil {
		panic(err)
	}

	// delegate alice capability to use the email service, but only allow sending
	// to example.com email addresses
	dlg, err := delegation.Delegate(
		mailer,
		alice,
		must(command.Parse("/message/send")),
		delegation.WithSubject(mailer),
		delegation.WithPolicy(pol),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println("Delegation:", dlg.Link())

	// invoke the capability
	inv, err := invocation.Invoke(
		alice,
		mailer,
		must(command.Parse("/message/send")),
		datamodel.NewMap(
			datamodel.WithEntry("to", []string{"bob@example.com"}),
			datamodel.WithEntry("subject", "Hello!"),
			datamodel.WithEntry("message", "Hello Bob, How do you do?"),
		),
		invocation.WithProofs(dlg.Link()),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println("Invocation:", inv.Link())

	ct, err := container.New(container.WithDelegations(dlg), container.WithInvocations(inv))
	if err != nil {
		panic(err)
	}

	buf, err := container.Encode(container.Base64Gzip, ct)
	if err != nil {
		panic(err)
	}

	// you could put this in a HTTP header if you like!
	fmt.Println("X-Ucan-Container:", string(buf))

	ct2, err := container.Decode(buf)
	if err != nil {
		panic(err)
	}

	fmt.Println("Delegation:", ct2.Delegations()[0].Link())
	fmt.Println("Invocation:", ct2.Invocations()[0].Link())
}
