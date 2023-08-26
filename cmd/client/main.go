package main

import (
	"context"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/pigeatgarlic/grpc-file-server/pkg/client"
)

func main() {
	addr := os.Args[2]
	file := os.Args[1]

	ctx := context.Background()
	var logger = log.New()

	// Set up connection with rpc server
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(addr,grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(fmt.Errorf("grpc Dial fail: %s/n", err))
	}

	mlsClient := client.NewClient(logger, conn)
	defer mlsClient.CloseConn()

	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}

	mlsClient.Upload(ctx, f)
	if err != nil {
		panic(err)
	}

	os.Exit(0)
}
