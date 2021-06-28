package main

import (
	"context"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"

	"bitbucket.org/maxex/mls_poc/pkg/client"
)

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "address",
				Value: ":8080",
				Usage: "port of server being ",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "upload",
				Aliases: []string{"u"},
				Usage:   "upload a file",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "path",
						Value:    "",
						Usage:    "path of file to be uploaded",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					ctx := context.Background()
					var logger = log.New()

					// Set up connection with rpc server
					var conn *grpc.ClientConn
					conn, err := grpc.Dial(c.String("address"), grpc.WithInsecure())
					if err != nil {
						log.Fatalf("grpc Dial fail: %s/n", err)
					}

					mlsClient := client.NewClient(logger, conn)
					defer mlsClient.CloseConn()

					f, err := os.Open(c.String("path"))
					if err != nil {
						return err
					}

					mlsClient.Upload(ctx, f)
					if err != nil {
						return err
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
