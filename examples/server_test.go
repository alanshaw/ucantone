package examples

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"testing"

	"github.com/alanshaw/ucantone/client"
	"github.com/alanshaw/ucantone/examples/types"
	"github.com/alanshaw/ucantone/execution"
	"github.com/alanshaw/ucantone/execution/bindexec"
	"github.com/alanshaw/ucantone/ipld"
	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/principal/ed25519"
	"github.com/alanshaw/ucantone/result"
	"github.com/alanshaw/ucantone/server"
	"github.com/alanshaw/ucantone/ucan/invocation"
	"github.com/alanshaw/ucantone/validator/bindcap"
	"github.com/alanshaw/ucantone/validator/capability"
)

func TestServer(t *testing.T) {
	echoCapability, err := capability.New("/example/echo")
	if err != nil {
		panic(err)
	}

	serviceID, err := ed25519.Generate()
	if err != nil {
		panic(err)
	}

	ucanSrv := server.NewHTTP(serviceID)

	// Register an echo handler that returns the invocation arguments as the result
	ucanSrv.Handle(echoCapability, func(req execution.Request) (execution.Response, error) {
		args := req.Invocation().Arguments()
		fmt.Printf("Echo: %s\n", args["message"])
		return execution.NewResponse(execution.WithSuccess(args))
	})

	// Start the server on a random available port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}

	httpSrv := http.Server{Handler: ucanSrv}

	go func() {
		err := httpSrv.Serve(listener)
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	serviceURL, err := url.Parse("http://" + listener.Addr().String())
	if err != nil {
		panic(err)
	}
	fmt.Printf("UCAN Server is running at %s\n", serviceURL.String())

	// Server is now running and can accept invocations!

	alice, err := ed25519.Generate()
	if err != nil {
		panic(err)
	}

	// Allow alice to invoke the echo capability
	dlg, err := echoCapability.Delegate(serviceID, alice, serviceID)
	if err != nil {
		panic(err)
	}

	inv, err := echoCapability.Invoke(
		alice,
		serviceID,
		ipld.Map{"message": "Hello, UCAN!"},
		invocation.WithProofs(dlg.Link()),
	)
	if err != nil {
		panic(err)
	}

	// create a client to send the invocation to the server
	c, err := client.NewHTTP(serviceURL)
	if err != nil {
		panic(err)
	}

	resp, err := c.Execute(execution.NewRequest(context.Background(), inv, execution.WithProofs(dlg)))
	if err != nil {
		panic(err)
	}

	result.MatchResultR0(
		resp.Out(),
		func(o ipld.Any) {
			fmt.Printf("Echo response: %+v\n", o)
		},
		func(x ipld.Any) {
			fmt.Printf("Invocation failed: %v\n", x)
		},
	)

	err = httpSrv.Shutdown(context.Background())
	if err != nil {
		panic(err)
	}
}

func TestTypedServer(t *testing.T) {
	echoCapability, err := bindcap.New[*types.EchoArguments]("/example/echo")
	if err != nil {
		panic(err)
	}

	serviceID, err := ed25519.Generate()
	if err != nil {
		panic(err)
	}

	ucanSrv := server.NewHTTP(serviceID)

	// Register an echo handler that returns the invocation arguments as the result
	ucanSrv.Handle(echoCapability, bindexec.NewHandler(func(req *bindexec.Request[*types.EchoArguments]) (*bindexec.Response[*types.EchoArguments], error) {
		args := req.Task().BindArguments()
		fmt.Printf("Echo: %s\n", args.Message)
		return bindexec.NewResponse(bindexec.WithSuccess(args))
	}))

	// Start the server on a random available port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}

	httpSrv := http.Server{Handler: ucanSrv}

	go func() {
		err := httpSrv.Serve(listener)
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	serviceURL, err := url.Parse("http://" + listener.Addr().String())
	if err != nil {
		panic(err)
	}
	fmt.Printf("UCAN Server is running at %s\n", serviceURL.String())

	// Server is now running and can accept invocations!

	alice, err := ed25519.Generate()
	if err != nil {
		panic(err)
	}

	// Allow alice to invoke the echo capability
	dlg, err := echoCapability.Delegate(serviceID, alice, serviceID)
	if err != nil {
		panic(err)
	}

	inv, err := echoCapability.Invoke(
		alice,
		serviceID,
		&types.EchoArguments{Message: "Hello, UCAN!"},
		invocation.WithProofs(dlg.Link()),
	)
	if err != nil {
		panic(err)
	}

	// create a client to send the invocation to the server
	c, err := client.NewHTTP(serviceURL)
	if err != nil {
		panic(err)
	}

	resp, err := c.Execute(execution.NewRequest(context.Background(), inv, execution.WithProofs(dlg)))
	if err != nil {
		panic(err)
	}

	result.MatchResultR0(
		resp.Out(),
		func(o ipld.Any) {
			args := types.EchoArguments{}
			err := datamodel.Rebind(datamodel.NewAny(o), &args)
			if err != nil {
				panic(err)
			}
			fmt.Printf("Echo response: %+v\n", args)
		},
		func(x ipld.Any) {
			fmt.Printf("Invocation failed: %v\n", x)
		},
	)

	err = httpSrv.Shutdown(context.Background())
	if err != nil {
		panic(err)
	}
}
