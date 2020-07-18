package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"

	"github.com/golang/glog"

	"github.com/jackc/pgtype"
	pgtypeuuid "github.com/jackc/pgtype/ext/gofrs-uuid"
	pgx "github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/nick96/cycad/consul"
	pb "github.com/nick96/cycad/editorservice/pb"
)

const (
	expectedMigrationVersion = 1
)

var (
	port = flag.String("port", "9090", "Port to listen on")
	db   = flag.String("db", "", "Database connection string")

	errNotImplemented = errors.New("Not implemented")
)

func main() {
	flag.Parse()
	defer glog.Flush()

	if *db == "" {
		glog.Exit("'db' flag is required")
	}

	glog.Infof("Connecting to database...")
	dbconfig, err := pgxpool.ParseConfig(*db)
	if err != nil {
		glog.Exitf("Faield to parse config: %v", err)
	}
	dbconfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		conn.ConnInfo().RegisterDataType(pgtype.DataType{
			Value: &pgtypeuuid.UUID{},
			Name:  "uuid",
			OID:   pgtype.UUIDOID,
		})
		return nil
	}

	db, err := pgxpool.Connect(context.Background(), *db)
	if err != nil {
		glog.Exitf("Failed to create session database on %s: %v", *db, err)
	}
	glog.Infof("Successfully connected to database")

	glog.Infof("Checking migration version is the expected version %d", expectedMigrationVersion)
	if err := checkMigrationVersion(db); err != nil {
		glog.Exitf("Database migraiton version check failed: %v", err)
	}
	glog.Infof("Database is migrated to expected version of %d", expectedMigrationVersion)

	// This needs to be 0.0.0.0 so that it works in a docker container. I
	// used localhost before and it was very frustrating to debug! It'll
	// work locally but the connection is refused if you call it from
	// outside the container.
	endpoint := fmt.Sprintf("0.0.0.0:%s", *port)
	glog.Infof("Starting TCP listener on %s...", endpoint)
	lis, err := net.Listen("tcp", endpoint)
	if err != nil {
		glog.Exitf("Failed to listener on %s: %v", endpoint, err)
	}
	glog.Infof("Successfully started listener on %s", endpoint)

	svr := grpc.NewServer()
	pb.RegisterEditorServiceServer(svr, newServer(db))
	var serviceName string
	// There should only be one service registered at the moment so we can
	// get its name from the service info map.
	for name := range svr.GetServiceInfo() {
		serviceName = name
	}
	reflection.Register(svr)

	glog.Info("Registering service with consul")
	if err := consul.RegisterService(serviceName, *port); err != nil {
		glog.Exitf("Failed to register service with consul: %v", err)
	}
	glog.Info("Successfull registered service with consul")

	glog.Infof("Starting editor service on port %s", *port)
	if err := svr.Serve(lis); err != nil {
		glog.Exitf("Server exited: %v", err)
	}
}
