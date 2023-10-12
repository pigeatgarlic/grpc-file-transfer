package main

import (
	"fmt"
	"net"
	"os"

	mlspb "github.com/pigeatgarlic/grpc-file-server/pkg/protobuf"
	"github.com/pigeatgarlic/grpc-file-server/pkg/server"
	"google.golang.org/grpc"
)

func main() {
	Dir  := os.Args[1]
	Port := os.Args[2]

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s",Port))
	if err != nil {
		fmt.Printf("failed %s\n",err.Error())
		os.Exit(1)
	}
	defer lis.Close()

	fmt.Printf("Now serving %s\n", lis.Addr().String())

	grpcServer := grpc.NewServer()

	s := server.NewMLSServer(Dir)
	mlspb.RegisterMLSServiceServer(grpcServer, s)
	defer grpcServer.GracefulStop()

	if err = grpcServer.Serve(lis); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
