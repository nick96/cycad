package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/golang/glog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/nick96/cycad/consul"
	pb "github.com/nick96/cycad/healthservice/pb"
)

const (
	expectedMigrationVersion = 1
)

var (
	errNotImplemented = errors.New("Not implemented")
)

func main() {
	flag.Parse()
	defer glog.Flush()

	port := os.Getenv("SERVICE_PORT")

	// This needs to be 0.0.0.0 so that it works in a docker container. I
	// used localhost before and it was very frustrating to debug! It'll
	// work locally but the connection is refused if you call it from
	// outside the container.
	endpoint := fmt.Sprintf("0.0.0.0:%s", port)
	glog.Infof("Starting TCP listener on %s...", endpoint)
	lis, err := net.Listen("tcp", endpoint)
	if err != nil {
		glog.Exitf("Failed to listener on %s: %v", endpoint, err)
	}
	glog.Infof("Successfully started listener on %s", endpoint)

	svr := grpc.NewServer()
	pb.RegisterHealthServer(svr, newServer())
	var serviceName string
	// There should only be one service registered at the moment so we can
	// get its name from the service info map.
	for name := range svr.GetServiceInfo() {
		serviceName = name
	}
	reflection.Register(svr)

	glog.Info("Registering service with consul")
	if err := consul.RegisterService(serviceName, port); err != nil {
		glog.Exitf("Failed to register service with consul: %v", err)
	}
	glog.Info("Successfull registered service with consul")

	glog.Infof("Starting health service on port %s", port)
	if err := svr.Serve(lis); err != nil {
		glog.Exitf("Server exited: %v", err)
	}
}
