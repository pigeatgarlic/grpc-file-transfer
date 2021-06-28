package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"

	"github.com/aleitner/grpc-file-server/pkg/server"
)

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "address",
				Value: ":8080",
				Usage: "server address",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "start",
				Aliases: []string{"s"},
				Usage:   "start server",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "path",
						Value:    "",
						Usage:    "path for data to be stored",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					// do a little server startin here
					logger := logrus.New()
					documentDir := c.String("path")

					lis, err := net.Listen("tcp", ":8080")
					if err != nil {
						fmt.Println("failed")
						os.Exit(1)
					}
					defer lis.Close()

					logger.Infof("Now serving %s", lis.Addr().String())

					grpcServer := grpc.NewServer()

					s := server.NewMLSServer(logger, documentDir)
					server.RegisterMLSServer(grpcServer, s)
					defer grpcServer.GracefulStop()

					if err = grpcServer.Serve(lis); err != nil {
						fmt.Println(err)
						os.Exit(1)
					}

					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
