package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"

	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"

	gw "github.com/nick96/cycad/editorservice/pb"
)

var (
	endpoint = flag.String("endpoint", "localhost:9090", "gRPC server endpoint")
	port     = flag.String("port", "8081", "Port to listen on")
)

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Register gRPC server endpoint
	// Note: Make sure the gRPC server is running properly and accessible
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := gw.RegisterEditorServiceHandlerFromEndpoint(ctx, mux, *endpoint, opts)
	if err != nil {
		return err
	}
	glog.Info("Registered editor service at %s", *endpoint)

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	glog.Info("Listening on port %s", *port)
	return http.ListenAndServe(fmt.Sprintf(":%s", *port), mux)
}

func main() {
	flag.Parse()
	defer glog.Flush()

	if err := run(); err != nil {
		glog.Fatal(err)
	}
}
