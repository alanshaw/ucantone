package examples

import (
	"fmt"
	"maps"
	"testing"

	"github.com/alanshaw/ucantone/examples/types"
	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/principal/ed25519"
	"github.com/alanshaw/ucantone/ucan/command"
	"github.com/alanshaw/ucantone/ucan/delegation"
	"github.com/alanshaw/ucantone/ucan/delegation/policy"
	"github.com/alanshaw/ucantone/ucan/delegation/policy/selector"
	"github.com/alanshaw/ucantone/ucan/invocation"
	"github.com/alanshaw/ucantone/validator/capability"
)

func TestCapabilityDefinition(t *testing.T) {
	messageSendCommand, err := command.Parse("/message/send")
	if err != nil {
		panic(err)
	}

	messageSendCapability := capability.New[*types.MessageSendArguments](
		messageSendCommand,
		capability.WithPolicy(
			policy.Not(
				policy.Equal(must(selector.Parse(".to")), []string{}),
			),
		),
	)

	// mailer is an email service that can send emails
	mailer, err := ed25519.Generate()
	if err != nil {
		panic(err)
	}

	alice, err := ed25519.Generate()
	if err != nil {
		panic(err)
	}

	// delegate alice capability to use the email service
	dlg, err := messageSendCapability.Delegate(
		mailer,
		alice,
		delegation.WithSubject(mailer),
	)
	if err != nil {
		panic(err)
	}

	// invoke the capability
	invocation, err := messageSendCapability.Invoke(
		alice,
		mailer,
		&types.MessageSendArguments{
			To:      []string{"bob@example.com"},
			Subject: "Hello!",
			Message: "Hello Bob, How do you do?",
		},
		invocation.WithProofs(dlg.Link()),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(invocation.Link())

	// Now, send the invocation to the service. You'll probably want to put the
	// invocation and delegation in a Container and send a HTTP request...
}

func TestCapabilityDefinitionGenericMap(t *testing.T) {
	messageSendCommand, err := command.Parse("/message/send")
	if err != nil {
		panic(err)
	}

	messageSendCapability := capability.New[*datamodel.Map](
		messageSendCommand,
		capability.WithPolicy(
			policy.Not(
				policy.Equal(must(selector.Parse(".to")), []string{}),
			),
		),
	)

	// mailer is an email service that can send emails
	mailer, err := ed25519.Generate()
	if err != nil {
		panic(err)
	}

	alice, err := ed25519.Generate()
	if err != nil {
		panic(err)
	}

	// delegate alice capability to use the email service
	dlg, err := messageSendCapability.Delegate(
		mailer,
		alice,
		delegation.WithSubject(mailer),
	)
	if err != nil {
		panic(err)
	}

	args := map[string]any{
		"to":      []string{"bob@example.com"},
		"subject": "Hello!",
		"message": "Hello Bob, How do you do?",
	}

	// invoke the capability
	invocation, err := messageSendCapability.Invoke(
		alice,
		mailer,
		datamodel.NewMap(datamodel.WithEntries(maps.All(args))),
		invocation.WithProofs(dlg.Link()),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(invocation.Link())

	// Now, send the invocation to the service. You'll probably want to put the
	// invocation and delegation in a Container and send a HTTP request...
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
