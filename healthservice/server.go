package main

import (
	"context"

	"github.com/golang/glog"
	"github.com/nick96/cycad/consul"
	pb "github.com/nick96/cycad/healthservice/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type healthServer struct {
	pb.UnimplementedHealthServer
}

func (*healthServer) Check(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	ok, err := consul.HealthCheckService(req.Service)
	if err != nil {
		glog.Infof("Failed to check service %s's status: %v", req.Service, err)
		return nil, status.Errorf(codes.Unavailable, "Failed to get the status of %s", req.Service)
	}

	response := new(pb.HealthCheckResponse)
	if ok {
		response.Status = pb.HealthCheckResponse_SERVING
	} else {
		response.Status = pb.HealthCheckResponse_NOT_SERVING
	}
	return response, nil
}
func (*healthServer) Watch(req *pb.HealthCheckRequest, srv pb.Health_WatchServer) error {
	return status.Errorf(codes.Unimplemented, "method Watch not implemented")
}

func newServer() *healthServer {
	svr := &healthServer{}
	return svr
}
