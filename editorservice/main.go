package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"

	"github.com/golang/glog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

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
	glog.Infof("Received AddNode request: %v", req)
	return &pb.AddNodeResponse{Name: "name", Content: "content"}, nil
}

func (s *editorServiceServer) GetNode(ctx context.Context, req *pb.GetNodeRequest) (*pb.GetNodeResponse, error) {
	glog.Infof("Received GetNode request: %v", req)
	return &pb.GetNodeResponse{Name: "name", Content: "content"}, nil
}

func newServer() *editorServiceServer {
	svr := &editorServiceServer{}
	return svr
}

func main() {
	flag.Parse()
	defer glog.Flush()

	// This needs to be 0.0.0.0 so that it works in a docker container. I
	// used localhost before and it was very frustrating to debug! It'll
	// work locally but the connection is refused if you call it from
	// outside the container.
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", *port))
	if err != nil {
		glog.Exitf("Failed to listen on port %s: %v", *port, err)
	}

	svr := grpc.NewServer()
	pb.RegisterEditorServiceServer(svr, newServer())
	reflection.Register(svr)
	glog.Infof("Starting editor service on port %s", *port)
	if err := svr.Serve(lis); err != nil {
		glog.Exitf("Server exited: %v", err)
	}
}
