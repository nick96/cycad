package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"

	"github.com/golang/glog"
	"google.golang.org/grpc"

	pb "github.com/nick96/cycad/editorservice/pb"
)

var (
	port = flag.String("port", "9090", "Port to listen on")

	errNotImplemented = errors.New("Not implemented")
)

type editorServiceServer struct {
	pb.UnimplementedEditorServiceServer
}

func (s *editorServiceServer) AddNode(ctx context.Context, req *pb.AddNodeRequest) (*pb.AddNodeResponse, error) {
	return nil, errNotImplemented
}

func (s *editorServiceServer) GetNode(ctx context.Context, req *pb.GetNodeRequest) (*pb.GetNodeResponse, error) {
	return nil, errNotImplemented
}

func newServer() *editorServiceServer {
	svr := &editorServiceServer{}
	return svr
}

func main() {
	flag.Parse()
	defer glog.Flush()

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", *port))
	if err != nil {
		glog.Exit("failed to listen on port %s: %v", *port, err)
	}

	svr := grpc.NewServer()
	pb.RegisterEditorServiceServer(svr, newServer())
	svr.Serve(lis)
}
