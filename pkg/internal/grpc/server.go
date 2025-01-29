package grpc

import (
	"net"

	"git.solsynth.dev/hypernet/nexus/pkg/proto"
	wproto "git.solsynth.dev/hypernet/wallet/pkg/proto"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	proto.UnimplementedDirectoryServiceServer
	wproto.UnimplementedPaymentServiceServer
	health.UnimplementedHealthServer

	srv *grpc.Server
}

func NewGrpc() *Server {
	server := &Server{
		srv: grpc.NewServer(),
	}

	proto.RegisterDirectoryServiceServer(server.srv, server)
	wproto.RegisterPaymentServiceServer(server.srv, server)
	health.RegisterHealthServer(server.srv, server)

	reflection.Register(server.srv)

	return server
}

func (v *Server) Listen() error {
	listener, err := net.Listen("tcp", viper.GetString("grpc_bind"))
	if err != nil {
		return err
	}

	return v.srv.Serve(listener)
}
