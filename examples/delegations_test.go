package examples

import (
	"testing"

	"github.com/alanshaw/ucantone/principal/ed25519"
	"github.com/alanshaw/ucantone/ucan/delegation"
	"github.com/alanshaw/ucantone/ucan/delegation/policy"
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
		mailer,
		"/message/send",
	)
	if err != nil {
		panic(err)
	}

	_, err = delegation.Delegate(
		alice,
		bob,
		mailer,
		"/message/send",
		// alice delegates bob capability to use the email service, but only allows
		// bob to send to example.com email addresses
		delegation.WithPolicyBuilder(
			policy.All(".to", policy.Like(".", "*.example.com")),
		),
	)
	if err != nil {
		panic(err)
	}
}
