package main

import (
	"fmt"
	"net"
	"os"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/pigeatgarlic/grpc-file-server/pkg/server"
	mlspb "github.com/pigeatgarlic/grpc-file-server/pkg/protobuf"
)

func main() {
	Dir := os.Args[1]

	logger := logrus.New()
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Printf("failed %s\n",err.Error())
		os.Exit(1)
	}
	defer lis.Close()

	logger.Infof("Now serving %s", lis.Addr().String())

	grpcServer := grpc.NewServer()

	s := server.NewMLSServer(logger, Dir)
	mlspb.RegisterMLSServiceServer(grpcServer, s)
	defer grpcServer.GracefulStop()

	if err = grpcServer.Serve(lis); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
