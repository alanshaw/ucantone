package server

import (
	"context"

	"github.com/alanshaw/ucantone/executor"
	"github.com/alanshaw/ucantone/transport"
	"github.com/alanshaw/ucantone/ucan"
)

type HandlerFunc = executor.HandlerFunc

type Server struct {
	executor *executor.Executor
}

func New() *Server {
	return &Server{
		executor: executor.New(),
	}
}

func (s *Server) Handle(cmd ucan.Command, handler HandlerFunc) {
	s.executor.Handle(cmd, handler)
}

// Request unpacks and executes an incoming request.
func (s *Server) Request(ctx context.Context, req transport.Request) (transport.Response, error) {
	panic("not implemented")
}

func (s *Server) Execute(ctx context.Context, request executor.Input, response executor.Output) {
	s.executor.Execute(ctx, request, response)
}
