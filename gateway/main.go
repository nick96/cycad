package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"

	gw "github.com/nick96/cycad/gateway/pb"
)

func run() error {
	port := os.Getenv("GATEWAY_PORT")
	editorEndpoint := os.Getenv("EDITOR_ENDPOINT")
	healthEndpoint := os.Getenv("HEALTH_ENDPOINT")

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}

	if err := gw.RegisterEditorServiceHandlerFromEndpoint(ctx, mux, editorEndpoint, opts); err != nil {
		return err
	}
	glog.Infof("Registered editor service at %s", editorEndpoint)

	if err := gw.RegisterHealthHandlerFromEndpoint(ctx, mux, healthEndpoint, opts); err != nil {
		return err
	}
	glog.Infof("Registered health service at %s", healthEndpoint)

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	glog.Infof("Listening on port %s", port)
	return http.ListenAndServe(fmt.Sprintf(":%s", port), mux)
}

func main() {
	flag.Parse()
	defer glog.Flush()

	if err := run(); err != nil {
		glog.Exitf("Editor service dateway exited: %v", err)
	}
}
