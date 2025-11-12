package examples

import (
	"testing"

	"github.com/alanshaw/ucantone/principal/ed25519"
	"github.com/alanshaw/ucantone/ucan/command"
	"github.com/alanshaw/ucantone/ucan/delegation"
	"github.com/alanshaw/ucantone/ucan/delegation/policy/builder"
)

func TestDelegations(t *testing.T) {
	// mailer is an email service that can send emails
	mailer, err := ed25519.Generate()
	if err != nil {
		panic(err)
	}

	alice, err := ed25519.Generate()
	if err != nil {
		panic(err)
	}

	bob, err := ed25519.Generate()
	if err != nil {
		panic(err)
	}

	// delegate alice capability to use the email service
	_, err = delegation.Delegate(
		mailer,
		alice,
		must(command.Parse("/message/send")),
		delegation.WithSubject(mailer),
	)
	if err != nil {
		panic(err)
	}

	// alice delegates bob capability to use the email service, but only allows
	// bob to send to example.com email addresses
	policy, err := builder.Build(builder.All(".to", builder.Like(".", "*.example.com")))
	if err != nil {
		panic(err)
	}

	_, err = delegation.Delegate(
		alice,
		bob,
		must(command.Parse("/message/send")),
		delegation.WithSubject(mailer),
		delegation.WithPolicy(policy),
	)
	if err != nil {
		panic(err)
	}
}
